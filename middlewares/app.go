package middlewares

import (
	"net/http"
	"os"

	"github.com/bitrise-io/addons-test-backend/env"
	"github.com/bitrise-io/addons-test-backend/models"
	"github.com/bitrise-io/api-utils/httpresponse"
	"github.com/jinzhu/gorm"
	"github.com/justinas/alice"
	"github.com/pkg/errors"
)

// AuthenticateForAppMiddleware ...
func AuthenticateForAppMiddleware(appEnv *env.AppEnv) alice.Chain {
	return CommonMiddleware(appEnv).Append(checkAuthenticatedApp(appEnv))
}

func checkAuthenticatedApp(appEnv *env.AppEnv) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if appEnv.ShouldSkipSessionAuthentication {
				appEnv.Session.Set("app_slug", os.Getenv("BITRISE_APP_SLUG"))
			}

			sessionAppSlug, ok := appEnv.Session.Get("app_slug").(string)
			if ok {
				_, err := appEnv.AppService.Find(&models.App{AppSlug: sessionAppSlug})
				switch {
				case gorm.IsRecordNotFoundError(err):
					httpresponse.RespondWithNotFoundErrorNoErr(w)
					return
				case err != nil:
					httpresponse.RespondWithInternalServerError(w, errors.WithMessage(err, "SQL Error"))
					return
				}
				return
			}
			httpresponse.RespondWithForbiddenNoErr(w)
		})
	}
}
