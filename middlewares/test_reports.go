package middlewares

import (
	"net/http"

	"github.com/bitrise-io/api-utils/httpresponse"
	"github.com/justinas/alice"

	"github.com/bitrise-io/addons-test-backend/bitrise"
	"github.com/bitrise-io/addons-test-backend/env"
	"github.com/bitrise-io/addons-test-backend/models"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

// AuthorizeForTestReportsMiddleware ...
func AuthorizeForTestReportsMiddleware(appEnv *env.AppEnv) alice.Chain {
	return CommonMiddleware(appEnv).Append(
		authenticateForAppAccess(appEnv),
		authorizeForRunningBuildViaBitriseAPI(appEnv),
	)
}

func authenticateForAppAccess(appEnv *env.AppEnv) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestParams := appEnv.RequestParams.Get(r)
			_, err := appEnv.AppService.Find(&models.App{
				AppSlug:  requestParams["app_slug"],
				APIToken: requestParams["token"],
			})
			switch {
			case gorm.IsRecordNotFoundError(err):
				httpresponse.RespondWithForbiddenNoErr(w)
				return
			case err != nil:
				httpresponse.RespondWithInternalServerError(w, errors.WithMessage(err, "SQL Error"))
				return
			}
		})
	}
}

func authorizeForRunningBuildViaBitriseAPI(appEnv *env.AppEnv) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if appEnv.ShouldSkipBuildAuthorizationWithBitriseAPI {
				return
			}

			requestParams := appEnv.RequestParams.Get(r)
			appSlug := requestParams["app_slug"]
			buildSlug := requestParams["build_slug"]

			app, err := appEnv.AppService.Find(&models.App{AppSlug: appSlug})
			switch {
			case gorm.IsRecordNotFoundError(err):
				httpresponse.RespondWithNotFoundErrorNoErr(w)
				return
			case err != nil:
				httpresponse.RespondWithInternalServerError(w, errors.WithMessage(err, "SQL Error"))
				return
			}

			client := bitrise.NewClient(app.BitriseAPIToken)
			resp, build, err := client.GetBuildOfApp(buildSlug, appSlug)
			if err != nil {
				httpresponse.RespondWithInternalServerError(w, errors.WithMessage(err, "Failed to get build from Bitrise API"))
				return
			}

			if resp.StatusCode != http.StatusOK {
				httpresponse.RespondWithForbidden(w)
				return
			}

			if build.Status != 0 {
				httpresponse.RespondWithForbidden(w)
				return
			}
		})
	}
}
