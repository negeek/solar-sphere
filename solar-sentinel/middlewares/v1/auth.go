package v1

import (
	"context"
	"net/http"
	"strings"

	"github.com/negeek/solar-sphere/solar-spectrum/accesskey"
	"github.com/negeek/solar-sphere/solar-spectrum/httpapi"
)

type contextKey string

const emailContextKey contextKey = "authenticated_email"

// EmailFromContext returns the authenticated caller's email, set by
// Authentication once a request has passed all checks.
func EmailFromContext(ctx context.Context) string {
	email, _ := ctx.Value(emailContextKey).(string)
	return email
}

// AccessChecker is the subset of the repository the auth middleware needs.
type AccessChecker interface {
	IsKeyRevoked(ctx context.Context, key string) (bool, error)
	UserExists(ctx context.Context, email string) (bool, error)
}

// Authentication verifies the bearer access key on every request: valid
// signature and expiry, not revoked, and belongs to a known user. It is a
// constructor (not the middleware itself) so the checker and verification
// key are explicit dependencies instead of hidden globals.
func Authentication(checker AccessChecker, verificationKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				httpapi.JsonResponse(w, false, http.StatusUnauthorized, "Provide access key", nil)
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				httpapi.JsonResponse(w, false, http.StatusUnauthorized, "Invalid Authorisation Header", nil)
				return
			}
			key := parts[1]

			claims, err := accesskey.Verify(key, verificationKey)
			if err != nil {
				httpapi.JsonResponse(w, false, http.StatusUnauthorized, "Invalid access key", nil)
				return
			}

			revoked, err := checker.IsKeyRevoked(r.Context(), key)
			if err != nil {
				httpapi.JsonResponse(w, false, http.StatusInternalServerError, "Error verifying access key", nil)
				return
			}
			if revoked {
				httpapi.JsonResponse(w, false, http.StatusUnauthorized, "Access key has been revoked", nil)
				return
			}

			exists, err := checker.UserExists(r.Context(), claims.Email)
			if err != nil {
				httpapi.JsonResponse(w, false, http.StatusInternalServerError, "Error verifying user", nil)
				return
			}
			if !exists {
				httpapi.JsonResponse(w, false, http.StatusUnauthorized, "Invalid user", nil)
				return
			}

			ctx := context.WithValue(r.Context(), emailContextKey, claims.Email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
