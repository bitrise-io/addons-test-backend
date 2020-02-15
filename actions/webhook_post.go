package actions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bitrise-io/api-utils/httprequest"

	"github.com/bitrise-io/addons-firebase-testlab/database"
	"github.com/bitrise-io/addons-firebase-testlab/junit"
	"github.com/bitrise-io/addons-firebase-testlab/testreportfiller"
	"github.com/bitrise-io/addons-test-backend/env"
	"github.com/bitrise-io/addons-test-backend/firebaseutils"
	"github.com/bitrise-io/addons-test-backend/models"
	"github.com/bitrise-io/api-utils/httpresponse"
	"github.com/pkg/errors"
	"github.com/xtgo/uuid"
	"go.uber.org/zap"
)

const (
	abortedBuildStatus      int    = 3
	buildTriggeredEventType string = "build/triggered"
	buildFinishedEventType  string = "build/finished"
)

// WebhookHandler ...
func WebhookHandler(env *env.AppEnv, w http.ResponseWriter, r *http.Request) error {
	buildType := r.Header.Get("Bitrise-Event-Type")

	if buildType != buildTriggeredEventType && buildType != buildFinishedEventType {
		return errors.New("Invalid Bitrise event type")
	}

	appData := &models.AppData{}
	defer httprequest.BodyCloseWithErrorLog(r)
	if err := json.NewDecoder(r.Body).Decode(appData); err != nil {
		return httpresponse.RespondWithBadRequestError(w, "Request body has invalid format")
	}

	app, err := env.AppService.Find(&models.App{AppSlug: appData.AppSlug})
	if err != nil {
		return errors.Wrap(err, "SQL Error")
	}

	switch buildType {
	case buildFinishedEventType:
		build := (*models.Build)(nil)
		if appData.BuildStatus == abortedBuildStatus {
			var err error
			build, err = env.BuildService.Find(&models.Build{AppSlug: app.AppSlug, BuildSlug: appData.BuildSlug})
			if err != nil {
				return httpresponse.RespondWithNotFoundError(w)
			}
			if build.TestExecutionID != "" {
				_, err := firebaseutils.CancelTestMatrix(build.TestMatrixID)
				if err != nil {
					return fmt.Errorf("Failed to cancel test matrix(id: %s), error: %+v", build.TestMatrixID, err)
				}
			}
		}

		totals, err := GetTotals(env, app.AppSlug, appData.BuildSlug, logger)
		if err != nil {
			env.Logger.Warn("Failed to get totals of test", zap.Any("app_data", appData), zap.Error(err))
			return httpresponse.RespondWithSuccess(w, app)
		}

		switch {
		case totals.Failed > 0 || totals.Inconclusive > 0:
			env.AnalyticsClient.TestReportSummaryGenerated(app.AppSlug, appData.BuildSlug, "fail", totals.Tests, time.Now())
		case totals != (Totals{}):
			env.AnalyticsClient.TestReportSummaryGenerated(app.AppSlug, appData.BuildSlug, "success", totals.Tests, time.Now())
		case totals == (Totals{}):
			env.AnalyticsClient.TestReportSummaryGenerated(app.AppSlug, appData.BuildSlug, "empty", totals.Tests, time.Now())
		default:
			env.AnalyticsClient.TestReportSummaryGenerated(app.AppSlug, appData.BuildSlug, "null", totals.Tests, time.Now())
		}

		testReportRecords := []models.TestReport{}
		err = database.GetTestReports(&testReportRecords, app.AppSlug, appData.BuildSlug)
		if err != nil {
			return errors.Wrap(err, "Failed to find test reports in DB")
		}

		env.AnalyticsClient.NumberOfTestReports(app.AppSlug, appData.BuildSlug, len(testReportRecords), time.Now())

		fAPI, err := firebaseutils.New()
		if err != nil {
			return errors.Wrap(err, "Failed to create Firebase API model")
		}
		parser := &junit.Client{}
		testReportFiller := testreportfiller.Filler{}

		testReportsWithTestSuites, err := testReportFiller.FillMore(testReportRecords, fAPI, parser, &http.Client{}, "")
		if err != nil {
			return errors.Wrap(err, "Failed to enrich test reports with JUNIT results")
		}
		for _, tr := range testReportsWithTestSuites {
			result := "success"
			for _, ts := range tr.TestSuites {
				if ts.Totals.Failed > 0 || totals.Inconclusive > 0 {
					result = "fail"
					break
				}
			}
			env.AnalyticsClient.TestReportResult(app.AppSlug, appData.BuildSlug, result, "unit", tr.ID, time.Now())
		}

		if build != nil && build.TestHistoryID != "" && build.TestExecutionID != "" {
			details, err := fAPI.GetTestsByHistoryAndExecutionID(build.TestHistoryID, build.TestExecutionID, app.AppSlug, appData.BuildSlug)
			if err != nil {
				return errors.Wrap(err, "Failed to get test details")
			}

			testDetails, err := fillTestDetails(details, fAPI, logger)
			if err != nil {
				return errors.Wrap(err, "Failed to prepare test details data structure")
			}
			result := "success"
			for _, detail := range testDetails {
				outcome := detail.Outcome
				if outcome == "failure" {
					result = "failed"
				}
				if result != "failed" {
					if outcome == "skipped" || outcome == "inconclusive" {
						result = outcome
					}
				}
			}

			env.AnalyticsClient.TestReportResult(app.AppSlug, appData.BuildSlug, result, "ui", uuid.UUID{}, time.Now())
		}
	case buildTriggeredEventType:
		// Don't care
	default:
		return errors.Errorf("Invalid build type: %s", buildType)
	}

	return c.Render(200, r.JSON(app))
}
