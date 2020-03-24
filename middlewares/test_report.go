package middlewares

import (
	"net/http"

	"github.com/bitrise-io/addons-test-backend/env"
	"github.com/bitrise-io/addons-test-backend/models"
	"github.com/bitrise-io/api-utils/httpresponse"
	"github.com/jinzhu/gorm"
	"github.com/justinas/alice"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// AuthorizeForTestReportManageMiddleware ...
func AuthorizeForTestReportManageMiddleware(appEnv *env.AppEnv) alice.Chain {
	return AuthorizeForTestReportManageMiddleware(appEnv).Append(
		authorizeForTestReport(appEnv),
	)
}

func authorizeForTestReport(appEnv *env.AppEnv) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestParams := appEnv.RequestParams.Get(r)
			buildSlug := requestParams["build_slug"]
			testReportID := requestParams["test_report_id"]
			_, err := appEnv.TestReportService.Find(&models.TestReport{ID: uuid.FromStringOrNil(testReportID), BuildSlug: buildSlug})
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
