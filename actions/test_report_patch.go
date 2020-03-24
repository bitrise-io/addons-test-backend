package actions

import (
	"encoding/json"
	"net/http"

	"github.com/bitrise-io/addons-test-backend/env"
	"github.com/bitrise-io/addons-test-backend/models"
	"github.com/bitrise-io/api-utils/httpresponse"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// TestReportPatchHandler ...
func TestReportPatchHandler(appEnv *env.AppEnv, w http.ResponseWriter, r *http.Request) error {
	requestParams := appEnv.RequestParams.Get(r)
	testReportID := requestParams["test_report_id"]

	params := testReportPatchParams{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		return httpresponse.RespondWithBadRequestError(w, "Failed to decode test report data")
	}

	testReport, err := appEnv.TestReportService.Find(&models.TestReport{
		ID: uuid.FromStringOrNil(testReportID),
	})
	switch {
	case gorm.IsRecordNotFoundError(err):
		return httpresponse.RespondWithNotFoundError(w)
	case err != nil:
		return errors.WithMessage(err, "SQL Error")
	}

	if params.Name != "" {
		testReport.Name = params.Name
	}
	if params.Uploaded.Valid {
		testReport.Uploaded = params.Uploaded.Bool
	}

	verrs, err := appEnv.TestReportService.Update(testReport, []string{"Name", "Uploaded"})
	if len(verrs) > 0 {
		return httpresponse.RespondWithUnprocessableEntity(w, verrs)
	}
	if err != nil {
		return errors.WithMessage(err, "SQL Error")
	}

	return httpresponse.RespondWithSuccess(w, testReport)
}
