package actions

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/bitrise-io/addons-test-backend/env"
	"go.uber.org/zap"
	toolresults "google.golang.org/api/toolresults/v1beta3"

	"github.com/bitrise-io/addons-test-backend/firebaseutils"
	"github.com/bitrise-io/addons-test-backend/junit"
	"github.com/bitrise-io/addons-test-backend/models"
	"github.com/bitrise-io/addons-test-backend/testreportfiller"
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
func GetTotals(appEnv *env.AppEnv, appSlug, buildSlug string) (Totals, error) {
	testReportRecords, err := appEnv.TestReportService.FindAll(&models.TestReport{AppSlug: appSlug, BuildSlug: buildSlug})
	if err != nil {
		return Totals{}, errors.Wrap(err, "Failed to find test reports in DB")
	}

	parser := &junit.Client{}
	testReportFiller := testreportfiller.Filler{}

	testReportsWithTestSuites, err := testReportFiller.FillMore(testReportRecords, appEnv.FirebaseAPI, parser, &http.Client{}, "")
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

	build, err := appEnv.BuildService.Find(&models.Build{AppSlug: appSlug, BuildSlug: buildSlug})
	if err != nil {
		// no Firebase tests, it's fine, we can return
		return totals, nil
	}

	if build.TestHistoryID == "" || build.TestExecutionID == "" {
		// no Firebase tests, it's fine, we can return
		return totals, nil
	}

	details, err := appEnv.FirebaseAPI.GetTestsByHistoryAndExecutionID(build.TestHistoryID, build.TestExecutionID, appSlug, buildSlug)
	if err != nil {
		return Totals{}, errors.Wrap(err, "Failed to get test details")
	}

	testDetails, err := fillTestDetails(details, appEnv.FirebaseAPI, appEnv.Logger)
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

func fillTestDetails(details *toolresults.ListStepsResponse, fAPI *firebaseutils.APIModel, logger *zap.Logger) ([]*models.Test, error) {
	testDetails := make([]*models.Test, len(details.Steps))

	var wg sync.WaitGroup
	wg.Add(len(details.Steps))
	errChannel := make(chan error, 1)

	for index, d := range details.Steps {
		go func(detail *toolresults.Step, i int) {
			defer func() {
				wg.Done()
			}()
			test := &models.Test{}
			for _, dimension := range detail.DimensionValue {
				switch dimension.Key {
				case "Model":
					test.DeviceName = firebaseutils.GetDeviceNameByID(dimension.Value)
				case "Version":
					prefixByPlatform := "API Level"
					if strings.Contains(strings.ToLower(detail.Name), "ios") {
						prefixByPlatform = "iOS"
					}
					test.APILevel = fmt.Sprintf("%s %s", prefixByPlatform, dimension.Value)
				case "Locale":
					test.Locale = firebaseutils.GetLangByCountryCode(dimension.Value)
				case "Orientation":
					test.Orientation = dimension.Value
				}
			}

			if detail.Outcome != nil {
				test.Outcome = detail.Outcome.Summary
			}
			test.Status = detail.State
			test.StepID = detail.StepId

			if detail.TestExecutionStep != nil {
				if len(detail.TestExecutionStep.TestIssues) > 0 {
					test.TestIssues = []models.TestIssue{}
					for _, issue := range detail.TestExecutionStep.TestIssues {
						testIssue := models.TestIssue{Name: issue.ErrorMessage}
						if issue.StackTrace != nil {
							testIssue.Stacktrace = issue.StackTrace.Exception
						}
						test.TestIssues = append(test.TestIssues, testIssue)
					}
				}
				outputURLs := models.OutputURLModel{}
				outputURLs.ScreenshotURLs = []string{}
				outputURLs.AssetURLs = map[string]string{}
				if detail.TestExecutionStep.TestTiming != nil {
					if detail.TestExecutionStep.TestTiming.TestProcessDuration != nil {
						test.StepDuration = int(detail.TestExecutionStep.TestTiming.TestProcessDuration.Seconds)
					}
				}

				test.TestResults = []models.TestResults{}
				for _, overview := range detail.TestExecutionStep.TestSuiteOverviews {
					testResult := models.TestResults{Total: int(overview.TotalCount), Failed: int(overview.FailureCount), Skipped: int(overview.SkippedCount)}
					test.TestResults = append(test.TestResults, testResult)
				}

				if detail.TestExecutionStep.ToolExecution != nil {
					//get logcat
					for _, testlog := range detail.TestExecutionStep.ToolExecution.ToolLogs {
						//create signed url for assets
						signedURL, err := fAPI.GetSignedURLOfLegacyBucketPath(testlog.FileUri)
						if err != nil {
							logger.Error("Failed to get signed url",
								zap.String("file_uri", testlog.FileUri),
								zap.Any("error", errors.WithStack(err)),
							)
							if len(errChannel) == 0 {
								errChannel <- err
							}
							return
						}

						outputURLs.LogURLs = append(outputURLs.LogURLs, signedURL)
					}

					// parse output files by type
					for _, output := range detail.TestExecutionStep.ToolExecution.ToolOutputs {
						{
							if strings.Contains(output.Output.FileUri, "results/") {
								//create signed url for asset
								signedURL, err := fAPI.GetSignedURLOfLegacyBucketPath(output.Output.FileUri)
								if err != nil {
									logger.Error("Failed to get signed url",
										zap.String("output_file_uri", output.Output.FileUri),
										zap.Any("error", errors.WithStack(err)),
									)
									if len(errChannel) == 0 {
										errChannel <- err
									}
									return
								}
								resultAbsPath := strings.Join(strings.Split(strings.Split(output.Output.FileUri, "results/")[1], "/")[1:], "/")
								outputURLs.AssetURLs[resultAbsPath] = signedURL
							}
						}

						if strings.HasSuffix(output.Output.FileUri, "video.mp4") {
							//create signed url for asset
							signedURL, err := fAPI.GetSignedURLOfLegacyBucketPath(output.Output.FileUri)
							if err != nil {
								logger.Error("Failed to get signed url",
									zap.String("output_file_uri", output.Output.FileUri),
									zap.Any("error", errors.WithStack(err)),
								)
								if len(errChannel) == 0 {
									errChannel <- err
								}
								return
							}
							outputURLs.VideoURL = signedURL
						}

						if strings.HasSuffix(output.Output.FileUri, "sitemap.png") {
							//create signed url for asset
							signedURL, err := fAPI.GetSignedURLOfLegacyBucketPath(output.Output.FileUri)
							if err != nil {
								logger.Error("Failed to get signed url",
									zap.String("output_file_uri", output.Output.FileUri),
									zap.Any("error", errors.WithStack(err)),
								)
								if len(errChannel) == 0 {
									errChannel <- err
								}
								return
							}
							outputURLs.ActivityMapURL = signedURL
						}

						if strings.HasSuffix(output.Output.FileUri, ".png") && !strings.HasSuffix(output.Output.FileUri, "sitemap.png") {
							//create signed url for asset
							signedURL, err := fAPI.GetSignedURLOfLegacyBucketPath(output.Output.FileUri)
							if err != nil {
								logger.Error("Failed to get signed url",
									zap.String("output_file_uri", output.Output.FileUri),
									zap.Any("error", errors.WithStack(err)),
								)
								if len(errChannel) == 0 {
									errChannel <- err
								}
								return
							}
							outputURLs.ScreenshotURLs = append(outputURLs.ScreenshotURLs, signedURL)
						}
					}
				}
				if detail.TestExecutionStep.TestSuiteOverviews != nil {
					//get xmls
					for _, overview := range detail.TestExecutionStep.TestSuiteOverviews {
						//create signed url for assets
						signedURL, err := fAPI.GetSignedURLOfLegacyBucketPath(overview.XmlSource.FileUri)
						if err != nil {
							logger.Error("Failed to get signed url",
								zap.String("xml_source_file_uri", overview.XmlSource.FileUri),
								zap.Any("error", errors.WithStack(err)),
							)
							if len(errChannel) == 0 {
								errChannel <- err
							}
							return
						}

						outputURLs.TestSuiteXMLURL = signedURL
					}
				}
				test.OutputURLs = outputURLs
			}

			if test.OutputURLs.ActivityMapURL != "" {
				test.TestType = "robo"
			}
			if test.OutputURLs.TestSuiteXMLURL != "" {
				test.TestType = "instrumentation"
			}

			testDetails[i] = test
		}(d, index)
	}
	wg.Wait()
	close(errChannel)

	var err error
	err = <-errChannel
	return testDetails, err
}

func filterTestsByStatus(tests []*models.Test, status string) []*models.Test {
	filteredTests := []*models.Test{}

	for _, test := range tests {
		if statusMatch(test.Outcome, status) || test.Status == "inProgress" { // include currently running tests too
			filteredTests = append(filteredTests, test)
		}
	}

	return filteredTests
}

func statusMatch(testStatus string, expected string) bool {
	if testStatus == expected {
		return true
	}

	if testStatus == "success" && expected == "passed" {
		return true
	}

	if testStatus == "failure" && expected == "failed" {
		return true
	}

	return false
}
