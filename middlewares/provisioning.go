package middlewares

import (
	"net/http"

	"github.com/bitrise-io/api-utils/httpresponse"

	"github.com/bitrise-io/addons-test-backend/env"
)

func authenticateForProvisioning(appEnv *env.AppEnv) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authentication") != appEnv.AddonAccessToken {
				httpresponse.RespondWithForbidden(w)
			}
		})
	}
}
