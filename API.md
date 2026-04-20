# API Reference — s101_api

Single source of truth for frontend wiring. Be precise — the backend changes fast.

---

## 1. Conventions

### Base URL
All endpoints (except `/healthz` and `/static/*`) are mounted under `/api`.

```
{BASE_URL}/api/...
```

### Content type
- Request bodies: JSON (`Content-Type: application/json`), except `POST /api/files/upload` which is `multipart/form-data`.
- Responses: JSON. `204 No Content` → empty body.

### Authentication
Bearer JWT (HS256) in `Authorization: Bearer <jwt>`.

Token `typ` claim:
- `"user"` — from `POST /api/auth/verify-code`
- `"admin"` — from `POST /api/admin/auth/login`

Middleware modes:
- **Public** — no auth.
- **RequireUser** — valid user JWT, user not blocked.
- **RequireAdmin** — valid admin JWT.
- **OptionalAuth** — only on `GET /api/places/:id`. Parses token if present.

### Error format
Every 4xx/5xx:

```json
{ "error": "bad_request", "message": "human-readable detail" }
```

| HTTP | `error` code     |
|------|------------------|
| 400  | `bad_request`    |
| 401  | `unauthorized`   |
| 403  | `forbidden`      |
| 404  | `not_found`      |
| 409  | `conflict`       |
| 500  | `internal_error` |

### Pagination

| Query   | Default | Max | Notes                 |
|---------|---------|-----|-----------------------|
| `page`  | `1`     | —   | 1-based.              |
| `limit` | `20`    | 100 | Capped server-side.   |

Envelope:
```json
{ "items": [ ... ], "page": 1, "limit": 20, "total": 137 }
```

### Status enum (places & claims)
| Value | Meaning   |
|-------|-----------|
| `0`   | Pending   |
| `10`  | Approved  |
| `-10` | Rejected  |

### I18nText
```json
{ "en": "English", "uz": "O'zbek" }
```

### Geo
- `lat` / `lon` float degrees (WGS84).
- Place also stores GeoJSON `location`: `{ "type": "Point", "coordinates": [lon, lat] }`.

### Images & avatars (file keys)
User avatars and place/review images are **file keys** (not full URLs), e.g. `"a8f3...e21.jpg"`.

To upload: call `POST /api/files/upload` (§12). It returns `{ file_id, key, url, usage }`.
- Save `key` into `avatar_key` (user), `logo_key` (place), or append to `images` (place/review).
- To display: prepend the static base → `{BASE_URL}/static/{key}`.

---

## 2. Core Models

### User (private — `/auth/me` & admin)
```json
{
  "id": "uuid",
  "name": "string",
  "username": "string|null",
  "phone": "+998...",
  "avatar_key": "string|null",
  "blocked": false,
  "created_at": "2026-04-17T10:00:00Z",
  "owns_place": true
}
```

### PublicUser (public endpoints)
```json
{
  "id": "uuid",
  "name": "string",
  "username": "string|null",
  "avatar_key": "string|null",
  "created_at": "2026-04-17T10:00:00Z"
}
```

### Place
```json
{
  "id": "uuid",
  "slug": "blue-cafe",
  "atc_id": "uz1726",
  "name": "Blue Cafe",
  "category_id": "uuid",
  "address": { "en": "...", "uz": "..." },
  "phone": "+998...",
  "description": { "en": "...", "uz": "..." },
  "lat": 41.31,
  "lon": 69.28,
  "location": { "type": "Point", "coordinates": [69.28, 41.31] },
  "logo_key": "<file-key>",
  "images": ["<file-key>"],
  "weekly_hours": {
    "mon": [{ "open": "09:00", "close": "22:00" }],
    "tue": [{ "open": "09:00", "close": "22:00" }],
    "wed": [], "thu": [], "fri": [], "sat": [], "sun": []
  },
  "status": 10,
  "avg_rating": 4.6,
  "review_count": 23,
  "created_by": "uuid|null",
  "claimed_by": "uuid|null",
  "created_at": "2026-04-17T10:00:00Z",
  "updated_at": "2026-04-17T10:00:00Z",
  "is_open": true
}
```
> `is_open` computed from `weekly_hours`. Present on `GET /api/places` and `GET /api/places/:id`; **not** on admin/write responses.

### Category
```json
{
  "id": "uuid",
  "slug": "cafe",
  "name": { "en": "Cafe", "uz": "Kafe" },
  "desc": { "en": "...", "uz": "..." },
  "created_at": "...",
  "updated_at": "..."
}
```

### Review
```json
{
  "id": "uuid",
  "place_id": "uuid",
  "user_id": "uuid|null",
  "star_rating": 5,
  "price_rating": 4,
  "quality_rating": 5,
  "text": "Great place",
  "images": ["<file-key>"],
  "latest": true,
  "created_at": "2026-04-17T10:00:00Z"
}
```
> Only user's *latest* review per place counts toward aggregates. History kept (`latest=false`).

### Bookmark
```json
{
  "id": "uuid",
  "user_id": "uuid",
  "place_id": "uuid",
  "created_at": "..."
}
```

### ClaimRequest
```json
{
  "id": "uuid",
  "place_id": "uuid",
  "user_id": "uuid",
  "phone": "+998...",
  "note": "optional",
  "status": 0,
  "reviewed_by": "uuid|null",
  "created_at": "...",
  "updated_at": "..."
}
```

### Admin
```json
{
  "id": "uuid",
  "username": "admin",
  "name": "Admin",
  "created_at": "..."
}
```

---

## 3. Health

### `GET /healthz`
**Public.** `200 → { "ok": true }`.

---

## 4. Public Auth (Telegram OTP)

User gets 6-digit code via Telegram bot, pastes it in web, exchanges for JWT.

### `POST /api/auth/verify-code`
Exchange OTP for JWT. Creates user on first login.

**Request**
```json
{ "code": "123456" }
```
- `code`: required, exactly 6 chars.

**200**
```json
{
  "token": "<jwt>",
  "user": {
    "id": "uuid",
    "name": "string",
    "username": "string|null",
    "avatar_key": null,
    "created_at": "..."
  }
}
```

**Errors**
- `400 bad_request` — `code must be 6 digits`.
- `401 unauthorized` — `invalid or expired code`.
- `403 forbidden` — `account is blocked`.

---

### `GET /api/auth/me`
Auth user's full profile.

**Auth:** RequireUser.

**200**
```json
{
  "id": "uuid",
  "name": "string",
  "username": "string|null",
  "phone": "+998...",
  "avatar_key": "string|null",
  "created_at": "...",
  "owns_place": true,
  "blocked": false
}
```
> `owns_place` = true iff user has ≥1 place with `claimed_by == user.id`.

---

## 5. Admin Auth

### `POST /api/admin/auth/login`
**Public.**

**Request**
```json
{ "username": "admin", "password": "secret" }
```

**200**
```json
{
  "token": "<jwt>",
  "admin": {
    "id": "uuid",
    "username": "admin",
    "name": "Admin",
    "created_at": "..."
  }
}
```

**Errors:** `401 unauthorized` — invalid credentials.

---

### `GET /api/admin/auth/me`
**Auth:** RequireAdmin. **200** → `Admin`.

---

## 6. Users

### `GET /api/users/:id`
Public profile (no phone, no telegram).

**Auth:** Public.

**200**
```json
{
  "user": {
    "id": "uuid",
    "name": "string",
    "username": "string|null",
    "avatar_key": "string|null",
    "created_at": "..."
  },
  "review_count": 12
}
```
> `review_count` counts only `latest=true` reviews.

**Errors:** `404 not_found`.

---

### `GET /api/users/:id/reviews`
User's latest reviews, paginated.

**Auth:** Public. **Query:** `page`, `limit`.

**200** → `Page<Review>`.

---

### `PUT /api/users/me`
Update own profile.

**Auth:** RequireUser.

**Request** (all optional; sent fields only)
```json
{
  "name": "New name",
  "username": "new_handle",
  "avatar_key": "<file-key>"
}
```
- `username`: lowercased + trimmed; must match `^[a-z0-9_]+$`. Must be unique.

**200** → `PublicUser`.

**Errors**
- `400 bad_request` — invalid body/username format.
- `409 conflict` — username taken.
- `404 not_found`.

---

### `DELETE /api/users/me`
Hard-delete current user.

**Auth:** RequireUser.

Side effects:
- Reviews kept but `user_id → null`.
- Places kept, `created_by → null`. `claimed_by` unset if matched.
- Bookmarks and pending claims deleted.

**204 No Content.**

---

## 7. Categories

### `GET /api/categories`
List all, sorted by slug.

**Auth:** Public. **200** → `[ Category, ... ]`.

---

## 8. Places

### `GET /api/places`
List approved places.

**Auth:** Public.

**Query**

| Param      | Description                                                          |
|------------|----------------------------------------------------------------------|
| `query`    | Full-text search on name/description (Mongo `$text`).                |
| `category` | Category UUID. Unknown → empty list.                         |
| `sort`     | `top` (default), `recent`, `nearest`. `nearest` needs `near`.        |
| `near`     | `lat,lon`. When set, sorts by distance even if `sort=top`.           |
| `page`     | See pagination.                                                      |
| `limit`    | See pagination.                                                      |

Sort:
- `top` — `avg_rating` desc, `review_count` desc.
- `recent` — `created_at` desc.
- `nearest` — geo ascending from `near`.

**200** → `Page<Place + is_open>`.

**Errors:** `400 bad_request` — `sort=nearest` without `near`.

---

### `GET /api/places/:id`
`:id` = UUID or slug.

**Auth:** OptionalAuth. Non-approved places visible only to: any admin, `created_by`, or `claimed_by`. Otherwise `404`.

**200** → `Place` (with `is_open`).

---

### `POST /api/places/create`
Create place (starts `status=0` pending).

**Auth:** RequireUser.

**Request**
```json
{
  "name": "Blue Cafe",
  "category_id": "uuid",
  "address": { "en": "...", "uz": "..." },
  "phone": "+998...",
  "description": { "en": "...", "uz": "..." },
  "lat": 41.31,
  "lon": 69.28,
  "logo_key": "<file-key>",
  "images": ["<file-key>"],
  "weekly_hours": { "mon": [{ "open": "09:00", "close": "22:00" }] }
}
```
Required: `name`, `category_id` (strict UUID), `address`, `lat`, `lon`.

**201** → `Place`.

**Errors**
- `400 bad_request` — `invalid category` / bad body.
- `500 internal_error` — slug collision overflow.

---

### `PUT /api/places/:id`
Edit. **Only `claimed_by` may call.** `:id` = **UUID only** (writes address stable IDs; slugs are read-only).

**Auth:** RequireUser.

**Request** (all optional)
```json
{
  "phone": "+998...",
  "description": { "en": "...", "uz": "..." },
  "weekly_hours": { ... },
  "logo_key": "<file-key>",
  "images": ["<file-key>"]
}
```

**200** → `Place` (with `is_open`).

**Errors**
- `403 forbidden` — `only the claimant can edit this place`.
- `404 not_found`.

---

## 9. Reviews

### `GET /api/places/:id/reviews`
Paginated. Default `latest=true` only. `:id` = UUID or slug.

**Auth:** Public. **Query:** `page`, `limit`, `all=true` for history.

**200** → `Page<Review>`.

---

### `POST /api/places/:id/reviews`
Create or replace user's review. Prior latest demoted to `latest=false`. `:id` = **UUID only**.

**Auth:** RequireUser. Place must be `status=10`.

**Request**
```json
{
  "star_rating": 5,
  "price_rating": 4,
  "quality_rating": 5,
  "text": "Great",
  "images": ["<file-key>"]
}
```
- `star_rating`: required, int 1–5.
- `price_rating`, `quality_rating`: optional, int 1–5.

**201** → `Review`.

**Errors**
- `403 forbidden` — `cannot review a non-approved place`.
- `404 not_found`.

---

### `DELETE /api/reviews/:id`
Delete own review. If was latest, next-most-recent is auto-promoted; aggregates recomputed.

**Auth:** RequireUser (author).

**204 No Content.**

**Errors:** `403 forbidden`, `404 not_found`.

---

## 10. Bookmarks (user-private)

All require **RequireUser**.

### `GET /api/bookmarks`
Bookmarks with referenced places nested inside. **Query:** `page`, `limit`.

**200** → `Page<BookmarkView>`.

```json
{
  "items": [
    {
      "id": "uuid",
      "user_id": "uuid",
      "place_id": "uuid",
      "created_at": "...",
      "place": { "id": "uuid", "name": "...", "is_open": true, ... }
    }
  ],
  "page": 1,
  "limit": 20,
  "total": 1
}
```
> Bookmarks sorted `created_at` desc.

---

### `POST /api/bookmarks/:placeId`
**201 Created** → `Bookmark`.
**208 Already Reported** → empty body (if already bookmarked).

**Errors:** `404 not_found`.

---

### `DELETE /api/bookmarks/:placeId`
**204 No Content.**

---

## 11. Claim Requests

Claim ownership of existing place. On approve, place's `claimed_by` flips; claimant can then edit via `PUT /api/places/:id`.

### `POST /api/claims`
**Auth:** RequireUser.

**Request**
```json
{ "place_id": "uuid", "phone": "+998...", "note": "optional" }
```

**201** → `ClaimRequest` (status=0).

**Errors**
- `404 not_found`.
- `409 conflict` — `this place is already claimed` / `you already have a pending claim for this place`.

---

### `GET /api/claims/mine`
User's claims, newest first. Not paginated.

**Auth:** RequireUser. **200** → `[ ClaimRequest, ... ]`.

---

## 12. Files

### `POST /api/files/upload`
Upload an image/asset.

**Auth:** RequireUser. **Content-Type:** `multipart/form-data`.

**Form fields**
| Field   | Type   | Notes                                    |
|---------|--------|------------------------------------------|
| `file`  | file   | required.                                |
| `usage` | string | required. One of `avatar`, `review`, `place`. |

**201**
```json
{
  "file_id": "uuid",
  "key": "uuid.jpg",
  "url": "/static/uuid.jpg",
  "usage": "avatar"
}
```

Use `key`:
- as `avatar_key` in `PUT /api/users/me`, or
- in `images[]` for place/review payloads.

Display via `{BASE_URL}{url}` (i.e. `{BASE_URL}/static/{key}`).

**Errors**
- `400 bad_request` — `file is required` / `invalid usage`.
- `500 internal_error` — upload failed.

---

### `GET /static/:key`
Serves uploaded files. **Public.** Static passthrough (not under `/api`).

---

## 13. Admin Endpoints

All under `/api/admin`, **RequireAdmin**.

### Places

#### `GET /api/admin/places`
List any status. Paginated.

**Query:** `status` (`0`, `10`, `-10`; other → no filter), `page`, `limit`.

**200** → `Page<Place>`.

---

#### `PUT /api/admin/places/:id/status`

**Request**
```json
{ "status": 10 }
```

**200** → `{ "ok": true, "status": 10 }`.

**Errors**
- `400 bad_request` — `status must be 0, 10, or -10`.
- `404 not_found`.

---

#### `PUT /api/admin/places/:id`
Edit any field; all optional. Sending both `lat` and `lon` updates geo `location`. `:id` and `category_id` are **UUID only**.

**Request**
```json
{
  "name": "...",
  "category_id": "uuid",
  "address": { "en": "...", "uz": "..." },
  "phone": "...",
  "description": { "en": "...", "uz": "..." },
  "lat": 41.31,
  "lon": 69.28,
  "logo_key": "<file-key>",
  "images": ["<file-key>"],
  "weekly_hours": { ... }
}
```

**200** → `{ "ok": true }`.

**Errors**
- `400 bad_request` — `invalid category`.
- `404 not_found`.

---

#### `DELETE /api/admin/places/:id`
Hard-deletes place + its reviews, bookmarks, claims.

**204 No Content.** **Errors:** `404 not_found`.

---

### Reviews

#### `GET /api/admin/reviews`
**Query:** `place_id` (optional), `page`, `limit`.

**200** → `Page<Review>`.

---

#### `DELETE /api/admin/reviews/:id`
Deletes any review; restores `latest` invariant; recomputes aggregates.

**204 No Content.** **Errors:** `404 not_found`.

---

### Users

#### `GET /api/admin/users`
Paginated. Phone + telegram included.

**200**
```json
{
  "items": [
    {
      "id": "uuid",
      "name": "string",
      "username": "string|null",
      "telegram_id": "string",
      "phone": "+998...",
      "avatar_key": "string|null",
      "blocked": false,
      "created_at": "..."
    }
  ],
  "page": 1, "limit": 20, "total": 1
}
```

---

#### `PUT /api/admin/users/:id/block`

**Request** `{ "blocked": true }`

**200** → `{ "ok": true, "blocked": true }`.

**Errors:** `404 not_found`.

---

### Claims

#### `GET /api/admin/claims`
**Query:** `status` (`0`, `10`, `-10`, optional), `page`, `limit`.

**200** → `Page<ClaimRequest>`.

---

#### `PUT /api/admin/claims/:id`
Approve/reject. On approve, sets place's `claimed_by = claim.user_id`.

**Request**
```json
{ "status": 10 }
```
- `status`: `10` (approve) or `-10` (reject).

**200** → `{ "ok": true, "status": 10 }`.

**Errors**
- `400 bad_request` — `status must be 10 or -10`.
- `404 not_found`.
- `409 conflict` — `place already claimed by another user`.

---

### Categories

#### `GET /api/admin/categories`
Same as public `GET /api/categories`. **200** → `[ Category, ... ]`.

---

#### `PUT /api/admin/categories/:id`
Edit `name`/`desc`. **Slug immutable.** All optional.

**Request**
```json
{
  "name": { "en": "Cafe", "uz": "Kafe" },
  "desc": { "en": "...", "uz": "..." }
}
```

**200** → `{ "ok": true }`. **Errors:** `404 not_found`.

---

## 14. Quick Endpoint Map

| Method | Path                                | Auth            |
|--------|-------------------------------------|-----------------|
| GET    | `/healthz`                          | public          |
| GET    | `/static/:key`                      | public          |
| POST   | `/api/auth/verify-code`             | public          |
| GET    | `/api/auth/me`                      | user            |
| POST   | `/api/admin/auth/login`             | public          |
| GET    | `/api/admin/auth/me`                | admin           |
| GET    | `/api/users/:id`                    | public          |
| GET    | `/api/users/:id/reviews`            | public          |
| PUT    | `/api/users/me`                     | user            |
| DELETE | `/api/users/me`                     | user            |
| GET    | `/api/categories`                   | public          |
| GET    | `/api/places`                       | public          |
| GET    | `/api/places/:id`                   | optional        |
| POST   | `/api/places/create`                | user            |
| PUT    | `/api/places/:id`                   | user (claimant) |
| GET    | `/api/places/:id/reviews`           | public          |
| POST   | `/api/places/:id/reviews`           | user            |
| DELETE | `/api/reviews/:id`                  | user (author)   |
| GET    | `/api/bookmarks`                    | user            |
| POST   | `/api/bookmarks/:placeId`           | user            |
| DELETE | `/api/bookmarks/:placeId`           | user            |
| POST   | `/api/claims`                       | user            |
| GET    | `/api/claims/mine`                  | user            |
| POST   | `/api/files/upload`                 | user            |
| GET    | `/api/admin/places`                 | admin           |
| PUT    | `/api/admin/places/:id/status`      | admin           |
| PUT    | `/api/admin/places/:id`             | admin           |
| DELETE | `/api/admin/places/:id`             | admin           |
| GET    | `/api/admin/reviews`                | admin           |
| DELETE | `/api/admin/reviews/:id`            | admin           |
| GET    | `/api/admin/users`                  | admin           |
| PUT    | `/api/admin/users/:id/block`        | admin           |
| GET    | `/api/admin/claims`                 | admin           |
| PUT    | `/api/admin/claims/:id`             | admin           |
| GET    | `/api/admin/categories`             | admin           |
| PUT    | `/api/admin/categories/:id`         | admin           |

---

## 15. CORS

All origins (`*`). Methods: `GET POST PUT PATCH DELETE OPTIONS`. Headers: `Origin Content-Type Authorization Accept`. Credentials allowed. Preflight cache 12h. Frontend can call the API directly from the browser.
