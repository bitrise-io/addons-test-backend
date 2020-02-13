package actions

import (
	"net/http"

	"github.com/bitrise-io/addons-test-backend/env"
	"github.com/bitrise-io/api-utils/httpresponse"
)

// RootHandler ...
func RootHandler(env *env.AppEnv, w http.ResponseWriter, r *http.Request) error {
	return httpresponse.RespondWithSuccess(w, map[string]string{"message": "Welcome to Bitrise Test Add-on!"})
}
