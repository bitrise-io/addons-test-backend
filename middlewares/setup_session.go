package middlewares

import (
	"net/http"

	"github.com/bitrise-io/addons-test-backend/env"
	"github.com/bitrise-io/addons-test-backend/session"
	"github.com/bitrise-io/api-utils/httpresponse"
)

func setupSession(appEnv *env.AppEnv) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sess, err := appEnv.SessionCookieStore.Get(r, appEnv.SessionName)
			if err != nil {
				httpresponse.RespondWithInternalServerError(w, "Failed to get session")
				return
			}
			sessionClient := session.NewClient(sess, r, w)
			appEnv.Session = &sessionClient

			h.ServeHTTP(r, w)
		})
	}
}
