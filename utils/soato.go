package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/itpu-student/s101_api/config"
)

// ResolveSOATOID calls the configured SOATO lookup API with the given
// coordinates and returns a "uz_{SOATO_ID}" string.
// If SOATO_LOOKUP_URL is unset, returns "uz_unknown" so local/dev flows work.
func ResolveSOATOID(ctx context.Context, lat, lon float64) (string, error) {
	base := config.Cfg.SOATOLookupURL
	if base == "" {
		return "uz_unknown", nil
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("lat", fmt.Sprintf("%f", lat))
	q.Set("lon", fmt.Sprintf("%f", lon))
	u.RawQuery = q.Encode()

	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(cctx, http.MethodGet, u.String(), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("soato: status %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var out struct {
		SOATOID string `json:"soato_id"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", err
	}
	if out.SOATOID == "" {
		return "uz_unknown", nil
	}
	return "uz_" + out.SOATOID, nil
}
