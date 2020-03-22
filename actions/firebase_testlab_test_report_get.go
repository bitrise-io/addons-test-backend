package actions

import (
	"net/http"

	"github.com/bitrise-io/addons-test-backend/env"
	"github.com/bitrise-io/addons-test-backend/models"
	"github.com/bitrise-io/api-utils/httpresponse"
	"github.com/jinzhu/gorm"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// FirebaseTestlabTestReportGetHandler ...
func FirebaseTestlabTestReportGetHandler(appEnv *env.AppEnv, w http.ResponseWriter, r *http.Request) error {
	requestParams := appEnv.RequestParams.Get(r)
	buildSlug := requestParams["build_slug"]
	status := requestParams["status"]

	appSlug, ok := appEnv.Session.Get("app_slug").(string)
	if !ok {
		return errors.New("Failed to get session data(app_slug)")
	}

	build, err := appEnv.BuildService.Find(&models.Build{AppSlug: appSlug, BuildSlug: buildSlug})
	switch {
	case gorm.IsRecordNotFoundError(err):
		return httpresponse.RespondWithNotFoundError(w)
	case err != nil:
		return errors.WithMessage(err, "SQL Error")
	}

	if build.TestHistoryID == "" || build.TestExecutionID == "" {
		appEnv.Logger.Error("No TestHistoryID or TestExecutionID found for build", zap.String("build_slug", build.BuildSlug))
		return httpresponse.RespondWithNotFoundError(w)
	}

	details, err := appEnv.FirebaseAPI.GetTestsByHistoryAndExecutionID(build.TestHistoryID, build.TestExecutionID, appSlug, buildSlug)
	if err != nil {
		return errors.WithMessage(err, "Failed to get test details")
	}

	//
	// prepare data structure
	testDetails, err := fillTestDetails(details, appEnv.FirebaseAPI, appEnv.Logger)
	if err != nil {
		return errors.WithMessage(err, "Failed to prepare test details data structure")
	}

	if status != "" {
		testDetails = filterTestsByStatus(testDetails, status)
	}

	return httpresponse.RespondWithSuccess(w, testDetails)
}
