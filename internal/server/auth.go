package server

import (
	"context"
	"net/http"
)

// UserValidatorFunc is a function that validates a user.
type UserValidatorFunc func(username, password string) bool

type authenticatedUserKeyType struct{}

var authenticatedUserKey = authenticatedUserKeyType{}

// WithAuthenticatedUser returns a new request with the authenticated user set.
func WithAuthenticatedUser(r *http.Request, username string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), authenticatedUserKey, username))
}

// AuthenticatedUser returns the authenticated user from a request context.
func AuthenticatedUser(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(authenticatedUserKey).(string)
	return username, ok
}

// AllowAnyone is a UserValidatorFunc that allows any user.
func AllowAnyone(username, password string) bool {
	return len(username) > 0
}

// Authenticate returns middlware that authenticates requests.
func Authenticate(validator UserValidatorFunc) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()
			if !ok || !validator(username, password) {
				w.Header().Set("WWW-Authenticate", `Basic realm="Pizza Bakery"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			r = WithAuthenticatedUser(r, username)
			next.ServeHTTP(w, r)
		})
	}
}
