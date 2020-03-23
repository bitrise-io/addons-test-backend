package middlewares

import (
	"net/http"

	"github.com/bitrise-io/addons-test-backend/env"
	"github.com/bitrise-io/addons-test-backend/session"
	"github.com/bitrise-io/api-utils/httpresponse"
	"github.com/bitrise-io/api-utils/logging"
	"github.com/bitrise-io/api-utils/middleware"
	"github.com/justinas/alice"
	"github.com/pkg/errors"
)

// CommonMiddleware ...
func CommonMiddleware(appEnv *env.AppEnv) alice.Chain {
	baseMiddleware := middleware.CommonMiddleware()

	if appEnv.Environment == env.ServerEnvProduction {
		baseMiddleware = baseMiddleware.Append(
			middleware.CreateRedirectToHTTPSMiddleware(),
		)
	}
	return baseMiddleware.Append(
		middleware.CreateOptionsRequestTerminatorMiddleware(),
		setupSession(appEnv),
		setupLogger(appEnv),
	)
}

func setupSession(appEnv *env.AppEnv) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sess, err := appEnv.SessionCookieStore.Get(r, appEnv.SessionName)
			if err != nil {
				httpresponse.RespondWithInternalServerError(w, errors.WithMessage(err, "Failed to get session"))
				return
			}
			sessionClient := session.NewClient(sess, r, w)
			appEnv.Session = &sessionClient
		})
	}
}

func setupLogger(appEnv *env.AppEnv) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			appEnv.Logger = logging.WithContext(r.Context())
		})
	}
}
