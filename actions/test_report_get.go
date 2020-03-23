package actions

import (
	"net/http"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"

	"github.com/bitrise-io/addons-test-backend/env"
	"github.com/bitrise-io/addons-test-backend/junit"
	"github.com/bitrise-io/addons-test-backend/models"
	"github.com/bitrise-io/addons-test-backend/testreportfiller"
	"github.com/bitrise-io/api-utils/httpresponse"
	"github.com/pkg/errors"
)

// TestReportGetHandler ...
func TestReportGetHandler(appEnv *env.AppEnv, w http.ResponseWriter, r *http.Request) error {
	requestParams := appEnv.RequestParams.Get(r)

	buildSlug := requestParams["build_slug"]
	testReportID := requestParams["test_report_id"]
	status := requestParams["status"]
	appSlug, ok := appEnv.Session.Get("app_slug").(string)
	if !ok {
		return errors.New("Failed to get session data(app_slug)")
	}

	testReport, err := appEnv.TestReportService.Find(&models.TestReport{
		ID:        uuid.FromStringOrNil(testReportID),
		AppSlug:   appSlug,
		BuildSlug: buildSlug,
		Uploaded:  true,
	})
	switch {
	case gorm.IsRecordNotFoundError(err):
		return httpresponse.RespondWithNotFoundError(w)
	case err != nil:
		return errors.WithMessage(err, "SQL Error")
	}

	parser := &junit.Client{}
	testReportFiller := testreportfiller.Filler{}

	testReportWithTestSuite, err := testReportFiller.FillOne(*testReport, appEnv.FirebaseAPI, parser, &http.Client{}, status)
	if err != nil {
		return errors.WithMessage(err, "Failed to enrich test report with JUNIT results")
	}

	return httpresponse.RespondWithSuccess(w, testReportWithTestSuite)
}
