package actions

import (
	"net/http"

	"github.com/bitrise-io/addons-test-backend/env"
	"github.com/bitrise-io/addons-test-backend/models"
	"github.com/bitrise-io/api-utils/httpresponse"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

// ProvisionDeleteHandler ...
func ProvisionDeleteHandler(appEnv *env.AppEnv, w http.ResponseWriter, r *http.Request) error {
	requestParams := appEnv.RequestParams.Get(r)
	appSlug := requestParams["app_slug"]
	err := appEnv.AppService.Delete(&models.App{AppSlug: appSlug})
	switch {
	case gorm.IsRecordNotFoundError(err):
		return httpresponse.RespondWithNotFoundError(w)
	case err != nil:
		return errors.WithMessage(err, "SQL Error")
	}

	return httpresponse.RespondWithSuccess(w, nil)
}
