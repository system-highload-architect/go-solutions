// Package appsec provides lightweight, zero‑dependency application security helpers
// for common web security tasks: safe redirects, HTML sanitisation, HMAC‑signed URLs,
// and input validation.
package appsec

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/url"
	"regexp"
	"strings"
)

// ----------------------------------------------------------------------------
// Redirect protection
// ----------------------------------------------------------------------------

// SafeRedirect validates that the given rawURL belongs to one of the allowed hosts.
// If the URL is relative (starts with '/'), it is considered safe.
// Returns the cleaned URL string or an error.
func SafeRedirect(rawURL string, allowedHosts []string) (string, error) {
	if rawURL == "" {
		return "", errors.New("appsec: empty redirect URL")
	}

	// Relative URLs are always safe (within the same origin).
	if strings.HasPrefix(rawURL, "/") && !strings.HasPrefix(rawURL, "//") {
		return rawURL, nil
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if parsed.Host == "" {
		// scheme-less or path-only, treat as relative
		return rawURL, nil
	}

	host := strings.ToLower(parsed.Hostname())
	for _, allowed := range allowedHosts {
		if strings.ToLower(allowed) == host {
			return rawURL, nil
		}
	}
	return "", errors.New("appsec: redirect to untrusted host")
}

// ----------------------------------------------------------------------------
// HTML Sanitisation
// ----------------------------------------------------------------------------

var (
	tagPattern     = regexp.MustCompile(`(?i)<\s*/?\s*(script|iframe|object|embed|form|link|meta|style|applet|frame|ilayer|layer|bgsound|title|base)[^>]*>`)
	attrPattern    = regexp.MustCompile(`(?i)\s*on\w+\s*=\s*"[^"]*"`)
	commentPattern = regexp.MustCompile(`<!--[\s\S]*?-->`)
)

// SanitizeHTML removes dangerous HTML tags and event handler attributes from the input.
// This is a basic implementation suitable for non‑critical contexts. For production
// user‑generated content, consider using a dedicated library like bluemonday.
func SanitizeHTML(input string) string {
	s := input
	s = commentPattern.ReplaceAllString(s, "")
	s = tagPattern.ReplaceAllString(s, "")
	s = attrPattern.ReplaceAllString(s, "")
	return s
}

// ----------------------------------------------------------------------------
// HMAC‑signed URL parameters
// ----------------------------------------------------------------------------

// SignURLParams appends an HMAC‑SHA256 signature to the given base URL and query parameters.
// The signature covers the path and sorted parameters, preventing tampering.
func SignURLParams(baseURL string, params map[string]string, secret []byte) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	q := parsed.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	// Build the canonical string: path + sorted encoded query
	canonical := parsed.Path + "?" + q.Encode()

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(canonical))
	sig := hex.EncodeToString(mac.Sum(nil))
	q.Set("sig", sig)
	parsed.RawQuery = q.Encode()
	return parsed.String(), nil
}

// VerifyURLParams checks the HMAC signature of a signed URL and returns the verified
// query parameters. If the signature is missing or invalid, an error is returned.
func VerifyURLParams(fullURL string, secret []byte) (url.Values, error) {
	parsed, err := url.Parse(fullURL)
	if err != nil {
		return nil, err
	}

	q := parsed.Query()
	sig := q.Get("sig")
	if sig == "" {
		return nil, errors.New("appsec: missing signature")
	}
	q.Del("sig")

	canonical := parsed.Path + "?" + q.Encode()
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(canonical))
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return nil, errors.New("appsec: invalid signature")
	}
	return q, nil
}

// ----------------------------------------------------------------------------
// Input validation
// ----------------------------------------------------------------------------

var safeIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// ValidID returns true if the string contains only alphanumeric characters,
// hyphens, and underscores. It is useful for validating user‑supplied IDs.
func ValidID(id string) bool {
	return safeIDPattern.MatchString(id)
}

var numberPattern = regexp.MustCompile(`^\d+$`)

// ValidNumber returns true if the string consists entirely of digits.
func ValidNumber(s string) bool {
	return numberPattern.MatchString(s)
}
