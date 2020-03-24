package actions

import (
	"net/http"

	"github.com/bitrise-io/addons-test-backend/env"
	"github.com/bitrise-io/addons-test-backend/models"
	"github.com/bitrise-io/api-utils/httpresponse"
	"github.com/pkg/errors"
)

// TestReportResponseItem ...
type TestReportResponseItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// TestReportsListHandler ...
func TestReportsListHandler(appEnv *env.AppEnv, w http.ResponseWriter, r *http.Request) error {
	requestParams := appEnv.RequestParams.Get(r)
	buildSlug := requestParams["build_slug"]
	appSlug, ok := appEnv.Session.Get("app_slug").(string)
	if !ok {
		return errors.New("Failed to get app slug from session")
	}
	testReports, err := appEnv.TestReportService.FindAll(&models.TestReport{
		AppSlug:   appSlug,
		BuildSlug: buildSlug,
		Uploaded:  true,
	})
	if err != nil {
		return errors.WithMessage(err, "SQL Error")
	}

	testReportsResponse := []TestReportResponseItem{}
	for _, tr := range testReports {
		testReportsResponse = append(testReportsResponse, testReportResponseItemFromTestReport(tr))
	}

	build, err := appEnv.BuildService.Find(&models.Build{AppSlug: appSlug, BuildSlug: buildSlug})
	if err != nil {
		return errors.WithMessage(err, "SQL Error")
	}

	if build.TestHistoryID == "" || build.TestExecutionID == "" {
		return httpresponse.RespondWithSuccess(w, testReportsResponse)
	}

	testReportsResponse = append(testReportsResponse, ftlReportItem())

	return httpresponse.RespondWithSuccess(w, testReportsResponse)
}

func testReportResponseItemFromTestReport(tr models.TestReport) TestReportResponseItem {
	return TestReportResponseItem{
		ID:   tr.ID.String(),
		Name: tr.Name,
	}
}

func ftlReportItem() TestReportResponseItem {
	return TestReportResponseItem{
		ID:   "ftl",
		Name: "Firebase TestLab",
	}
}
