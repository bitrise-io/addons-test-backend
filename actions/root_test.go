package actions_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bitrise-io/addons-test-backend/actions"
	"github.com/bitrise-io/addons-test-backend/env"
	"github.com/c2fo/testify/require"
)

func Test_RootHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	require.NoError(t, actions.RootHandler(&env.AppEnv{}, rr, req))

	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, `{"message":"Welcome to Bitrise Test Add-on!"}`+"\n", rr.Body.String())
}
