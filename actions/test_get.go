package actions

import (
	"net/http"
	"time"

	"github.com/bitrise-io/addons-test-backend/env"
	"github.com/bitrise-io/addons-test-backend/models"
	"github.com/bitrise-io/api-utils/httpresponse"
	"github.com/bitrise-io/go-utils/sliceutil"
	"github.com/gobuffalo/nulls"
	"github.com/pkg/errors"
)

// TestGetHandler ...
func TestGetHandler(appEnv *env.AppEnv, w http.ResponseWriter, r *http.Request) error {
	requestParams := appEnv.RequestParams.Get(r)

	buildSlug := requestParams["build_slug"]
	appSlug := requestParams["app_slug"]

	build, err := appEnv.BuildService.Find(&models.Build{AppSlug: appSlug, BuildSlug: buildSlug})
	if err != nil {
		return errors.WithMessage(err, "SQL Error")
	}

	if build.TestHistoryID == "" || build.TestExecutionID == "" {
		matrix, err := appEnv.FirebaseAPI.GetHistoryAndExecutionIDByMatrixID(build.TestMatrixID)
		if err != nil {
			matrix, err = appEnv.FirebaseAPI.GetHistoryAndExecutionIDByMatrixID(build.TestMatrixID)
			if err != nil {
				return errors.WithMessage(err, "Failed to get test status")
			}
		}

		if isMessageAnError(matrix.State) {
			return errors.Errorf("Failed to get test status: %s(%s)", matrix.State, matrix.InvalidMatrixDetails)
		}

		if len(matrix.TestExecutions) == 0 {
			build.LastRequest = nulls.NewTime(time.Now())

			verrs, err := appEnv.BuildService.Update(build, []string{"LastRequest"})
			if len(verrs) > 0 {
				return httpresponse.RespondWithUnprocessableEntity(w, verrs)
			}
			if err != nil {
				return errors.WithMessage(err, "SQL Error")
			}
			return httpresponse.RespondWithSuccess(w, map[string]string{"state": matrix.State})
		}

		if matrix.TestExecutions[0].ToolResultsStep == nil {
			build.LastRequest = nulls.NewTime(time.Now())

			verrs, err := appEnv.BuildService.Update(build, []string{"LastRequest"})
			if len(verrs) > 0 {
				return httpresponse.RespondWithUnprocessableEntity(w, verrs)
			}
			if err != nil {
				return errors.WithMessage(err, "SQL Error")
			}
			return httpresponse.RespondWithSuccess(w, map[string]string{"state": matrix.State})
		}

		build.TestHistoryID = matrix.TestExecutions[0].ToolResultsStep.HistoryId
		build.TestExecutionID = matrix.TestExecutions[0].ToolResultsStep.ExecutionId
	}

	steps, err := appEnv.FirebaseAPI.GetTestsByHistoryAndExecutionID(build.TestHistoryID, build.TestExecutionID, appSlug, buildSlug, "steps(state,name,outcome,dimensionValue,testExecutionStep)")
	if err != nil {
		steps, err = appEnv.FirebaseAPI.GetTestsByHistoryAndExecutionID(build.TestHistoryID, build.TestExecutionID, appSlug, buildSlug, "steps(state,name,outcome,dimensionValue,testExecutionStep)")
		if err != nil {
			return errors.WithMessage(err, "Failed to get test status")
		}
	}

	build.LastRequest = nulls.NewTime(time.Now())

	verrs, err := appEnv.BuildService.Update(build, []string{"LastRequest"})
	if len(verrs) > 0 {
		return httpresponse.RespondWithUnprocessableEntity(w, verrs)
	}
	if err != nil {
		return errors.WithMessage(err, "SQL Error")
	}
	return httpresponse.RespondWithSuccess(w, steps)
}

func isMessageAnError(message string) bool {
	errorMessages := []string{
		"ERROR",
		"UNSUPPORTED_ENVIRONMENT",
		"INCOMPATIBLE_ENVIRONMENT",
		"INCOMPATIBLE_ARCHITECTURE",
		"CANCELLED",
		"INVALID",
	}
	return sliceutil.IsStringInSlice(message, errorMessages)
}
