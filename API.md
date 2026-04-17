# API Reference — s101_api

This document describes every HTTP endpoint exposed by the backend. Use it as a single source of truth when wiring the frontend (`buyelp_web`) to the API.

---

## 1. Conventions

### Base URL
All endpoints (except `/healthz`) are mounted under `/api`.

```
{BASE_URL}/api/...
```

### Content type
- All request bodies are JSON: `Content-Type: application/json`.
- All responses are JSON unless documented otherwise (`204 No Content` returns an empty body).

### Authentication
Two kinds of bearer tokens, both JWT (HS256), sent in the `Authorization` header:

```
Authorization: Bearer <jwt>
```

The token's `typ` claim is either:
- `"user"` — issued via `POST /api/auth/verify-code`
- `"admin"` — issued via `POST /api/admin/auth/login`

Three middleware modes are used by routes:
- **Public** — no auth required.
- **RequireUser** — must be a valid user JWT and the user must not be blocked.
- **RequireAdmin** — must be a valid admin JWT.
- **OptionalAuth** — used only on `GET /api/places/:id`. Token is parsed if present, but the request is not rejected when missing.

### Error format
Any 4xx/5xx response uses this shape:

```json
{
  "error": "bad_request",
  "message": "human-readable detail"
}
```

`error` codes used by the backend:

| HTTP | `error` code      |
|------|-------------------|
| 400  | `bad_request`     |
| 401  | `unauthorized`    |
| 403  | `forbidden`       |
| 404  | `not_found`       |
| 409  | `conflict`        |
| 500  | `internal_error`  |

### Pagination
Endpoints that return lists accept:

| Query   | Default | Max | Notes                       |
|---------|---------|-----|-----------------------------|
| `page`  | `1`     | —   | 1-based page number.        |
| `limit` | `20`    | 100 | Capped server-side at 100.  |

Paginated responses use this envelope:

```json
{
  "items": [ ... ],
  "page": 1,
  "limit": 20,
  "total": 137
}
```

### Status enum (places & claims)
The backend uses a numeric `status` field in two places:

| Value | Meaning   |
|-------|-----------|
| `0`   | Pending   |
| `10`  | Approved  |
| `-10` | Rejected  |

### Internationalized text (`I18nText`)
Wherever you see a "translated" field (e.g. `address`, `description`, category `name`/`desc`), the JSON shape is:

```json
{ "en": "English text", "uz": "O'zbek tili" }
```

### Geo
- `lat` / `lon` are float degrees (WGS84).
- A place also stores a GeoJSON `location` field used for `$nearSphere` queries: `{ "type": "Point", "coordinates": [lon, lat] }`.

---

## 2. Core Models

### User (private — only returned to `/auth/me` & admins)
```json
{
  "id": "uuid",
  "name": "string",
  "username": "string|null",
  "phone": "string",
  "avatar_url": "string",
  "blocked": false,
  "created_at": "2026-04-17T10:00:00Z",
  "owns_place": true
}
```

### PublicUser (returned to public endpoints)
```json
{
  "id": "uuid",
  "name": "string",
  "username": "string|null",
  "avatar_url": "string",
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
  "images": ["https://..."],
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
> `is_open` is computed at response time from `weekly_hours`. It is present on `GET /api/places` and `GET /api/places/:id` but **not** on admin or write responses.

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
  "images": ["https://..."],
  "latest": true,
  "created_at": "2026-04-17T10:00:00Z"
}
```
> Only the user's *latest* review per place participates in `avg_rating` / `review_count`. History is preserved (`latest=false`).

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
Liveness probe. **Public.**

**200**
```json
{ "ok": true }
```

---

## 4. Public Auth (Telegram OTP)

The user requests a code through the Telegram bot. They paste that 6-digit code into the web app, which exchanges it for a JWT.

### `POST /api/auth/verify-code`
Exchange a 6-digit OTP for a JWT. Creates the user on first login.

**Request**
```json
{ "code": "123456" }
```
- `code` (string, required, exactly 6 chars).

**200**
```json
{
  "token": "<jwt>",
  "user": {
    "id": "uuid",
    "name": "string",
    "username": "string|null",
    "avatar_url": "",
    "created_at": "..."
  }
}
```

**Errors**
- `401 unauthorized` — code is invalid, expired, or already used.
- `403 forbidden` — `account is blocked`.

---

### `GET /api/auth/me`
Returns the authenticated user's full private profile.

**Auth:** RequireUser.

**200**
```json
{
  "id": "uuid",
  "name": "string",
  "username": "string|null",
  "phone": "+998...",
  "avatar_url": "string",
  "created_at": "...",
  "owns_place": true,
  "blocked": false
}
```
> `owns_place` is `true` if the user has at least one place where `claimed_by == user.id`.

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

**Errors**
- `401 unauthorized` — invalid credentials.

---

### `GET /api/admin/auth/me`
Return current admin profile. **Auth:** RequireAdmin.

**200** → `Admin`

---

## 6. Users (Public Profiles & Self-Service)

### `GET /api/users/:id`
Public profile of a user (no phone, no telegram id).

**Auth:** Public.

**200**
```json
{
  "user": {
    "id": "uuid",
    "name": "string",
    "username": "string|null",
    "avatar_url": "string",
    "created_at": "..."
  },
  "review_count": 12
}
```
> `review_count` only counts the user's *latest* reviews.

**Errors:** `404 not_found`.

---

### `GET /api/users/:id/reviews`
List of a user's latest reviews. Paginated.

**Auth:** Public.

**Query:** `page`, `limit`.

**200**
```json
{
  "items": [ Review, ... ],
  "page": 1,
  "limit": 20,
  "total": 12
}
```

---

### `PUT /api/users/me`
Update own profile.

**Auth:** RequireUser.

**Request** (all fields optional)
```json
{
  "name": "New name",
  "avatar_url": "https://..."
}
```

**200** → `PublicUser`

---

### `DELETE /api/users/me`
Hard-delete the current user.

**Auth:** RequireUser.

Side effects:
- Reviews authored by the user are kept but their `user_id` becomes `null`.
- Places created by the user keep their record but `created_by` becomes `null`.
- If the user had `claimed_by` ownership, it is unset.
- Bookmarks and pending claim requests are deleted.

**204 No Content.**

---

## 7. Categories

### `GET /api/categories`
List every category, sorted by slug.

**Auth:** Public.

**200** → `[ Category, ... ]`

---

## 8. Places

### `GET /api/places`
List approved places.

**Auth:** Public.

**Query**

| Param      | Description                                                                 |
|------------|-----------------------------------------------------------------------------|
| `query`    | Full-text search across name/description (uses Mongo `$text`).              |
| `category` | Category UUID **or** slug. If unknown, returns an empty list.               |
| `sort`     | `top` (default), `recent`, `nearest`. `nearest` requires `near=lat,lon`.    |
| `near`     | `lat,lon` pair. When set, results are sorted by distance even for `sort=top`. |
| `page`     | See pagination.                                                             |
| `limit`    | See pagination.                                                             |

Sort behavior:
- `top` — `avg_rating` desc, then `review_count` desc.
- `recent` — `created_at` desc.
- `nearest` — geospatial ascending distance from `near`.

**200**
```json
{
  "items": [ Place (with is_open), ... ],
  "page": 1,
  "limit": 20,
  "total": 137
}
```

**Errors:** `400 bad_request` if `sort=nearest` is used without `near`.

---

### `GET /api/places/:id`
Get one place. `:id` may be its UUID **or** its slug.

**Auth:** OptionalAuth. Pending/rejected places are visible only to:
- the admin (any admin token), or
- the place's `created_by` user, or
- the place's `claimed_by` user.

Otherwise, they appear as `404 not_found`.

**200** → `Place` (with `is_open`).

---

### `POST /api/places/create`
Create a new place. The created place starts in `status=0` (pending).

**Auth:** RequireUser.

**Request**
```json
{
  "name": "Blue Cafe",
  "category_id": "uuid-or-slug",
  "address": { "en": "...", "uz": "..." },
  "phone": "+998...",
  "description": { "en": "...", "uz": "..." },
  "lat": 41.31,
  "lon": 69.28,
  "images": ["https://..."],
  "weekly_hours": {
    "mon": [{ "open": "09:00", "close": "22:00" }]
  }
}
```
Required: `name`, `category_id`, `address`, `lat`, `lon`.

**201** → `Place`.

**Errors**
- `400 bad_request` — `invalid category` or invalid body.
- `500 internal_error` — slug collision overflow.

---

### `PUT /api/places/:id`
Edit a place. **Only the `claimed_by` user may call this.**

`:id` accepts UUID or slug.

**Auth:** RequireUser.

**Request** (all optional, only sent fields are updated)
```json
{
  "phone": "+998...",
  "description": { "en": "...", "uz": "..." },
  "weekly_hours": { ... },
  "images": ["https://..."]
}
```

**200** → `Place` (with `is_open`).

**Errors**
- `403 forbidden` — `only the claimant can edit this place`.
- `404 not_found`.

---

## 9. Reviews

### `GET /api/places/:id/reviews`
List reviews for a place. Paginated. By default returns only `latest=true` reviews.

`:id` accepts UUID or slug.

**Auth:** Public.

**Query:** `page`, `limit`, optional `all=true` to include the full history (older revisions of users' reviews).

**200**
```json
{
  "items": [ Review, ... ],
  "page": 1,
  "limit": 20,
  "total": 23
}
```

---

### `POST /api/places/:id/reviews`
Create or replace the user's review for the given place.

The backend keeps an append-only history: posting a new review demotes the previous one to `latest=false` and inserts a new `latest=true` row. Only the latest review for `(place, user)` counts toward ratings.

**Auth:** RequireUser. Place must be `status=10` (approved).

**Request**
```json
{
  "star_rating": 5,
  "price_rating": 4,
  "quality_rating": 5,
  "text": "Great service",
  "images": ["https://..."]
}
```
- `star_rating`: required, integer 1–5.
- `price_rating`, `quality_rating`: optional, integer 1–5 each.

**201** → `Review`.

**Errors**
- `403 forbidden` — `cannot review a non-approved place`.
- `404 not_found` — place missing.

---

### `DELETE /api/reviews/:id`
Delete one of your own reviews.

**Auth:** RequireUser (must be the author).

If the deleted review was the latest one, the next-most-recent review by the same user for the same place is auto-promoted to `latest=true`. The place's aggregates are recomputed.

**204 No Content.**

**Errors**
- `403 forbidden` — `not your review`.
- `404 not_found`.

---

## 10. Bookmarks (User-Private)

All bookmark endpoints require **RequireUser**.

### `GET /api/bookmarks`
List the user's bookmarks **and** the corresponding places in one call.

**Query:** `page`, `limit` (paginates the bookmarks).

**200**
```json
{
  "bookmarks": [ Bookmark, ... ],
  "places": [ Place, ... ]
}
```
> Bookmarks are sorted by `created_at` desc. The `places` array contains the `Place` records referenced by those bookmarks (no guaranteed order — match by `place_id`).

---

### `POST /api/bookmarks/:placeId`
Bookmark a place.

**201** → `Bookmark`.

If the user already bookmarked it:
**200**
```json
{ "ok": true, "already": true }
```

**Errors:** `404 not_found` if the place doesn't exist.

---

### `DELETE /api/bookmarks/:placeId`
Remove a bookmark.

**204 No Content.**

---

## 11. Claim Requests

A user can claim ownership of an *existing* place. Approval flips the place's `claimed_by`. After that, only the claimant can edit the place via `PUT /api/places/:id`.

### `POST /api/claims`
Submit a new claim.

**Auth:** RequireUser.

**Request**
```json
{
  "place_id": "uuid",
  "phone": "+998...",
  "note": "optional contact note"
}
```

**201** → `ClaimRequest` (status = 0).

**Errors**
- `404 not_found` — place missing.
- `409 conflict` — `this place is already claimed` or `you already have a pending claim for this place`.

---

### `GET /api/claims/mine`
List the current user's claim requests, newest first.

**Auth:** RequireUser.

**200** → `[ ClaimRequest, ... ]` (not paginated).

---

## 12. Admin Endpoints

All endpoints below are mounted under `/api/admin` and require **RequireAdmin**.

### Places

#### `GET /api/admin/places`
List places (any status). Paginated.

**Query**
- `status` — optional, one of `0`, `10`, `-10`. Other values are ignored (no filter).
- `page`, `limit`.

**200**
```json
{ "items": [ Place, ... ], "page": 1, "limit": 20, "total": 42 }
```

---

#### `PUT /api/admin/places/:id/status`
Change a place's moderation status.

**Request**
```json
{ "status": 10 }
```
- `status`: required, one of `0`, `10`, `-10`.

**200**
```json
{ "ok": true }
```

**Errors**
- `400 bad_request` — `status must be 0, 10, or -10`.
- `404 not_found`.

---

#### `PUT /api/admin/places/:id`
Admin can edit any field. All fields optional. Sending both `lat` and `lon` updates the geo `location` too.

**Request**
```json
{
  "name": "...",
  "category_id": "uuid-or-slug",
  "address": { "en": "...", "uz": "..." },
  "phone": "...",
  "description": { "en": "...", "uz": "..." },
  "lat": 41.31,
  "lon": 69.28,
  "images": ["..."],
  "weekly_hours": { ... }
}
```

**200**
```json
{ "ok": true }
```

**Errors**
- `400 bad_request` — `invalid category`.
- `404 not_found`.

---

#### `DELETE /api/admin/places/:id`
Hard-delete the place plus its reviews, bookmarks, and claim requests.

**204 No Content.**

**Errors:** `404 not_found`.

---

### Reviews

#### `GET /api/admin/reviews`
List reviews. Paginated.

**Query:** `place_id` (optional filter), `page`, `limit`.

**200**
```json
{ "items": [ Review, ... ], "page": 1, "limit": 20, "total": 100 }
```

---

#### `DELETE /api/admin/reviews/:id`
Delete any review. Restores the `latest` invariant and recomputes the place's aggregates.

**204 No Content.**

**Errors:** `404 not_found`.

---

### Users

#### `GET /api/admin/users`
List users. Paginated. Phone numbers **are** included for admins.

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
      "avatar_url": "string",
      "blocked": false,
      "created_at": "..."
    }
  ],
  "page": 1, "limit": 20, "total": 1
}
```

---

#### `PUT /api/admin/users/:id/block`
Block or unblock a user.

**Request**
```json
{ "blocked": true }
```

**200**
```json
{ "ok": true, "blocked": true }
```

**Errors:** `404 not_found`.

---

### Claims

#### `GET /api/admin/claims`
List claim requests. Paginated.

**Query:** `status` (`0`, `10`, `-10`, optional), `page`, `limit`.

**200**
```json
{ "items": [ ClaimRequest, ... ], "page": 1, "limit": 20, "total": 5 }
```

---

#### `PUT /api/admin/claims/:id`
Approve or reject a claim. On approve, the place's `claimed_by` is set to the claim's `user_id`.

**Request**
```json
{ "status": 10 }
```
- `status`: required, `10` (approve) or `-10` (reject).

**200**
```json
{ "ok": true, "status": 10 }
```

**Errors**
- `400 bad_request` — `status must be 10 or -10`.
- `404 not_found` — claim or place missing.
- `409 conflict` — `place already claimed by another user`.

---

### Categories

#### `GET /api/admin/categories`
Same payload as the public `GET /api/categories` (kept separate for future divergence).

**200** → `[ Category, ... ]`

---

#### `PUT /api/admin/categories/:id`
Edit a category's name/description. **Slug is immutable.** All fields optional.

**Request**
```json
{
  "name": { "en": "Cafe", "uz": "Kafe" },
  "desc": { "en": "...", "uz": "..." }
}
```

**200**
```json
{ "ok": true }
```

**Errors:** `404 not_found`.

---

## 13. Quick Endpoint Map

| Method | Path                                | Auth         |
|--------|-------------------------------------|--------------|
| GET    | `/healthz`                          | public       |
| POST   | `/api/auth/verify-code`             | public       |
| GET    | `/api/auth/me`                      | user         |
| POST   | `/api/admin/auth/login`             | public       |
| GET    | `/api/admin/auth/me`                | admin        |
| GET    | `/api/users/:id`                    | public       |
| GET    | `/api/users/:id/reviews`            | public       |
| PUT    | `/api/users/me`                     | user         |
| DELETE | `/api/users/me`                     | user         |
| GET    | `/api/categories`                   | public       |
| GET    | `/api/places`                       | public       |
| GET    | `/api/places/:id`                   | optional     |
| POST   | `/api/places/create`                | user         |
| PUT    | `/api/places/:id`                   | user (claimant) |
| GET    | `/api/places/:id/reviews`           | public       |
| POST   | `/api/places/:id/reviews`           | user         |
| DELETE | `/api/reviews/:id`                  | user (author)|
| GET    | `/api/bookmarks`                    | user         |
| POST   | `/api/bookmarks/:placeId`           | user         |
| DELETE | `/api/bookmarks/:placeId`           | user         |
| POST   | `/api/claims`                       | user         |
| GET    | `/api/claims/mine`                  | user         |
| GET    | `/api/admin/places`                 | admin        |
| PUT    | `/api/admin/places/:id/status`      | admin        |
| PUT    | `/api/admin/places/:id`             | admin        |
| DELETE | `/api/admin/places/:id`             | admin        |
| GET    | `/api/admin/reviews`                | admin        |
| DELETE | `/api/admin/reviews/:id`            | admin        |
| GET    | `/api/admin/users`                  | admin        |
| PUT    | `/api/admin/users/:id/block`        | admin        |
| GET    | `/api/admin/claims`                 | admin        |
| PUT    | `/api/admin/claims/:id`             | admin        |
| GET    | `/api/admin/categories`             | admin        |
| PUT    | `/api/admin/categories/:id`         | admin        |

---

## 14. CORS

The server allows all origins (`*`), the methods `GET POST PUT PATCH DELETE OPTIONS`, and the headers `Origin Content-Type Authorization Accept`. Credentials are allowed; preflight cache is 12h. The frontend can call the API directly from the browser.
