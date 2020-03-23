package actions

import (
	"encoding/json"
	"net/http"

	"github.com/bitrise-io/addons-test-backend/env"
	"github.com/bitrise-io/addons-test-backend/models"
	"github.com/bitrise-io/api-utils/httprequest"
	"github.com/bitrise-io/api-utils/httpresponse"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

// ChangePlanData ...
type ChangePlanData struct {
	Plan string `json:"plan"`
}

// ProvisionPutHandler ...
func ProvisionPutHandler(appEnv *env.AppEnv, w http.ResponseWriter, r *http.Request) error {
	requestParams := appEnv.RequestParams.Get(r)
	appSlug := requestParams["app_slug"]
	params := ChangePlanData{}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		return httpresponse.RespondWithBadRequestError(w, "Failed to decode provisioning data")
	}
	defer httprequest.BodyCloseWithErrorLog(r)

	app, err := appEnv.AppService.Find(&models.App{AppSlug: appSlug})
	switch {
	case gorm.IsRecordNotFoundError(err):
		return httpresponse.RespondWithNotFoundError(w)
	case err != nil:
		return errors.WithMessage(err, "SQL Error")
	}

	app.Plan = params.Plan
	verrs, err := appEnv.AppService.Update(app, []string{"Plan"})
	if len(verrs) > 0 {
		return httpresponse.RespondWithUnprocessableEntity(w, verrs)
	}
	if err != nil {
		return errors.WithMessage(err, "SQL Error")
	}

	return httpresponse.RespondWithSuccess(w, nil)
}
