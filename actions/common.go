package actions

import (
	"net/http"

	"github.com/bitrise-io/addons-test-backend/env"

	"github.com/bitrise-io/addons-firebase-testlab/junit"
	"github.com/bitrise-io/addons-firebase-testlab/testreportfiller"
	"github.com/bitrise-io/addons-test-backend/firebaseutils"
	"github.com/bitrise-io/addons-test-backend/models"
	"github.com/pkg/errors"
)

// Totals ...
type Totals struct {
	Tests        int `json:"tests"`
	Passed       int `json:"passed"`
	Skipped      int `json:"skipped"`
	Failed       int `json:"failed"`
	Inconclusive int `json:"inconclusive"`
}

// GetTotals ...
func GetTotals(env *env.AppEnv, appSlug, buildSlug string) (Totals, error) {
	testReportRecords, err := env.TestReportService.FindAll(&models.TestReport{AppSlug: appSlug, BuildSlug: buildSlug})
	if err != nil {
		return Totals{}, errors.Wrap(err, "Failed to find test reports in DB")
	}

	fAPI, err := firebaseutils.New()
	if err != nil {
		return Totals{}, errors.Wrap(err, "Failed to create Firebase API model")
	}
	parser := &junit.Client{}
	testReportFiller := testreportfiller.Filler{}

	testReportsWithTestSuites, err := testReportFiller.FillMore(testReportRecords, fAPI, parser, &http.Client{}, "")
	if err != nil {
		return Totals{}, errors.Wrap(err, "Failed to enrich test reports with JUNIT results")
	}

	var totals Totals

	for _, testReport := range testReportsWithTestSuites {
		for _, testSuite := range testReport.TestSuites {
			totals.Passed = totals.Passed + testSuite.Totals.Passed
			totals.Failed = totals.Failed + testSuite.Totals.Failed + testSuite.Totals.Error
			totals.Skipped = totals.Skipped + testSuite.Totals.Skipped
			totals.Tests = totals.Tests + testSuite.Totals.Tests
		}
	}

	build, err := env.BuildService.Find(&models.Build{AppSlug: appSlug, BuildSlug: buildSlug})
	if err != nil {
		// no Firebase tests, it's fine, we can return
		return totals, nil
	}

	if build.TestHistoryID == "" || build.TestExecutionID == "" {
		// no Firebase tests, it's fine, we can return
		return totals, nil
	}

	details, err := fAPI.GetTestsByHistoryAndExecutionID(build.TestHistoryID, build.TestExecutionID, appSlug, buildSlug)
	if err != nil {
		return Totals{}, errors.Wrap(err, "Failed to get test details")
	}

	testDetails, err := fillTestDetails(details, fAPI, env.Logger)
	if err != nil {
		return Totals{}, errors.Wrap(err, "Failed to prepare test details data structure")
	}

	for _, testDetail := range testDetails {
		switch testDetail.Outcome {
		case "success":
			totals.Passed++
		case "failure":
			totals.Failed++
		case "skipped":
			totals.Skipped++
		case "inconclusive":
			totals.Inconclusive++
		}
	}
	return totals, nil
}
