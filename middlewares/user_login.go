package middlewares

import (
	"net/http"
	"os"

	"github.com/bitrise-io/addons-firebase-testlab/database"
	"github.com/bitrise-io/addons-test-backend/env"
	"github.com/bitrise-io/api-utils/httpresponse"
	"github.com/pkg/errors"
)

func checkAuthenticatedApp(appEnv *env.AppEnv) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if appEnv.ShouldSkipSessionAuthentication {
				appEnv.Session.Set("app_slug", os.Getenv("BITRISE_APP_SLUG"))
			}

			sessionAppSlug, ok := appEnv.Session.Get("app_slug").(string)
			if ok {
				exists, err := database.IsAppExists(sessionAppSlug)
				if err != nil {
					httpresponse.RespondWithInternalServerError(w, errors.WithMessage(err, "SQL Error"))
					return
				}
				if exists {
					return
				}
			}
			httpresponse.RespondWithForbiddenNoErr(w)
		})
	}
}
