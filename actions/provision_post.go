package actions

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bitrise-io/api-utils/httprequest"
	"github.com/jinzhu/gorm"

	"github.com/bitrise-io/addons-test-backend/bitrise"
	"github.com/bitrise-io/addons-test-backend/env"
	"github.com/bitrise-io/addons-test-backend/models"
	"github.com/bitrise-io/addons-test-backend/utils"
	"github.com/bitrise-io/api-utils/httpresponse"
	"github.com/pkg/errors"
)

// ProvisionData ...
type ProvisionData struct {
	Plan            string `json:"plan"`
	AppSlug         string `json:"app_slug"`
	BitriseAPIToken string `json:"api_token"`
}

// Env ...
type Env struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ProvisionPostHandler ...
func ProvisionPostHandler(appEnv *env.AppEnv, w http.ResponseWriter, r *http.Request) error {
	params := ProvisionData{}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		return httpresponse.RespondWithBadRequestError(w, "Failed to decode provisioning data")
	}
	defer httprequest.BodyCloseWithErrorLog(r)

	app, err := appEnv.AppService.Find(&models.App{AppSlug: params.AppSlug})
	if !gorm.IsRecordNotFoundError(err) && err != nil {
		return errors.WithMessage(err, "SQL Error")
	}
	if app != nil {
		appEnv.Logger.Warn("  [!] App already exists")
	}

	envs := map[string][]Env{}
	envs["envs"] = append(envs["envs"], Env{Key: "ADDON_VDTESTING_API_URL", Value: fmt.Sprintf("%s/test", appEnv.HostName)})

	if app == nil {
		app = &models.App{
			AppSlug:         params.AppSlug,
			Plan:            params.Plan,
			BitriseAPIToken: params.BitriseAPIToken,
			APIToken:        utils.GenerateRandomHash(50),
		}

		app, verrs, err := appEnv.AppService.Create(app)
		if len(verrs) > 0 {
			return httpresponse.RespondWithUnprocessableEntity(w, verrs)
		}
		if err != nil {
			return errors.WithMessage(err, "SQL Error")
		}

		client := bitrise.NewClient(app.BitriseAPIToken)
		_, err = client.RegisterWebhook(app)
		if err != nil {
			return errors.WithMessage(err, "Failed to register webhook for app")
		}

		envs["envs"] = append(envs["envs"], Env{Key: "ADDON_VDTESTING_API_TOKEN", Value: app.APIToken})
		return httpresponse.RespondWithSuccess(w, envs)
	}

	envs["envs"] = append(envs["envs"], Env{Key: "ADDON_VDTESTING_API_TOKEN", Value: app.APIToken})
	return httpresponse.RespondWithSuccess(w, envs)
}
