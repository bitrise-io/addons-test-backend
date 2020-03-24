package actions

import (
	"net/http"

	"github.com/bitrise-io/api-utils/httpresponse"

	"github.com/bitrise-io/addons-test-backend/env"
	"github.com/pkg/errors"
)

// TestSummaryResponseModel ...
type TestSummaryResponseModel struct {
	Totals Totals `json:"totals"`
}

// TestSummaryGetHandler ...
func TestSummaryGetHandler(appEnv *env.AppEnv, w http.ResponseWriter, r *http.Request) error {
	requestParams := appEnv.RequestParams.Get(r)
	buildSlug := requestParams["build_slug"]

	appSlug, ok := appEnv.Session.Get("app_slug").(string)
	if !ok {
		return errors.New("Failed to get session data(app_slug)")
	}

	totals, err := getTotals(appEnv, appSlug, buildSlug)
	if err != nil {
		return errors.WithMessage(err, "Failed to get totals")
	}

	return httpresponse.RespondWithSuccess(w, TestSummaryResponseModel{
		Totals: totals,
	})
}
