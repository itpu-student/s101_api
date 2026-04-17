package utils

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var slugNonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify turns any string into a URL-friendly lowercase slug.
func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	// strip diacritics
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	if cleaned, _, err := transform.String(t, s); err == nil {
		s = cleaned
	}
	s = slugNonAlnum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "place"
	}
	return s
}
