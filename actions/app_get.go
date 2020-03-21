package actions

import (
	"net/http"

	"github.com/bitrise-io/addons-test-backend/env"
	"github.com/bitrise-io/api-utils/httpresponse"

	"github.com/pkg/errors"
)

// AppGetResponse ...
type AppGetResponse struct {
	AppSlug  string `json:"app_slug"`
	AppTitle string `json:"app_title"`
}

// AppGetHandler ...
func AppGetHandler(appEnv *env.AppEnv, w http.ResponseWriter, r *http.Request) error {
	appSlug, ok := appEnv.Session.Get("app_slug").(string)
	if ok {
		return errors.New("Failed to get app slug from session")
	}
	appTitle, ok := appEnv.Session.Get("app_title").(string)
	if ok {
		return errors.New("Failed to get app title from session")
	}

	return httpresponse.RespondWithSuccess(w, AppGetResponse{
		AppSlug:  appSlug,
		AppTitle: appTitle,
	})
}
