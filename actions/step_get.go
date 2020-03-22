package actions

import (
	"net/http"

	"github.com/bitrise-io/addons-test-backend/env"
	"github.com/bitrise-io/addons-test-backend/models"
	"github.com/bitrise-io/api-utils/httpresponse"
	"github.com/pkg/errors"
)

// StepGetHandler ...
func StepGetHandler(appEnv *env.AppEnv, w http.ResponseWriter, r *http.Request) error {
	requestParams := appEnv.RequestParams.Get(r)
	stepID := requestParams["step_id"]
	buildSlug := requestParams["build_slug"]
	appSlug, ok := appEnv.Session.Get("app_slug").(string)
	if !ok {
		return errors.New("Failed to get app slug from session")
	}

	build, err := appEnv.BuildService.Find(&models.Build{AppSlug: appSlug, BuildSlug: buildSlug})
	if err != nil {
		return errors.WithMessage(err, "SQL Error")
	}

	samples, err := appEnv.FirebaseAPI.GetTestMetricSamples(build.TestHistoryID, build.TestExecutionID, stepID, appSlug, buildSlug)
	if err != nil {
		return errors.WithMessage(err, "Failed to get sample data")
	}

	return httpresponse.RespondWithSuccess(w, samples)
}
