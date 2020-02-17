package middlewares

import (
	"net/http"

	"github.com/bitrise-io/addons-test-backend/env"
	"github.com/justinas/alice"
)

// SetupSessionMiddleware ...
func SetupSessionMiddleware(appEnv *env.AppEnv) alice.Chain {
	return CommonMiddleware(appEnv).Append(setupSession(appEnv))
}

func setupSession(appEnv *env.AppEnv) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		})
	}
}
