package router

import (
	"net/http"

	"github.com/bitrise-io/addons-test-backend/actions"
	"github.com/bitrise-io/addons-test-backend/env"
	"github.com/bitrise-io/addons-test-backend/middlewares"
	"github.com/bitrise-io/api-utils/handlers"
	"github.com/justinas/alice"
	"gopkg.in/DataDog/dd-trace-go.v1/contrib/gorilla/mux"
)

// New ...
func New(appEnv *env.AppEnv) *mux.Router {
	// StrictSlash: allow "trim slash"; /x/ REDIRECTS to /x
	r := mux.NewRouter(mux.WithServiceName("addons-test-mux")).StrictSlash(true)

	for _, route := range []struct {
		path           string
		middleware     alice.Chain
		handler        func(e *env.AppEnv, w http.ResponseWriter, r *http.Request) error
		allowedMethods []string
	}{
		{
			path: "/", middleware: middlewares.CommonMiddleware(appEnv),
			handler: actions.RootHandler, allowedMethods: []string{"GET", "OPTIONS"},
		},
		// PROVISIONING
		{
			path: "/provision", middleware: middlewares.AuthenticateForProvisioningMiddleware(appEnv),
			handler: actions.ProvisionPostHandler, allowedMethods: []string{"POST", "OPTIONS"},
		},
		{
			path: "/provision/{app_slug}", middleware: middlewares.AuthenticateForProvisioningMiddleware(appEnv),
			handler: actions.ProvisionPutHandler, allowedMethods: []string{"PUT", "OPTIONS"},
		},
		{
			path: "/provision/{app_slug}", middleware: middlewares.AuthenticateForProvisioningMiddleware(appEnv),
			handler: actions.ProvisionDeleteHandler, allowedMethods: []string{"DELETE", "OPTIONS"},
		},
		// TESTING
		{
			path: "/test/apps/{app_slug}/builds/{build_slug}/test_reports/{token}", middleware: middlewares.AuthorizeForTestReportsMiddleware(appEnv),
			handler: actions.TestReportGetHandler, allowedMethods: []string{"POST", "OPTIONS"},
		},
		{
			path: "/test/apps/{app_slug}/builds/{build_slug}/test_reports/{test_report_id}/{token}", middleware: middlewares.AuthorizeForTestReportManageMiddleware(appEnv),
			handler: actions.TestReportGetHandler, allowedMethods: []string{"PATCH", "OPTIONS"},
		},
		// API
		{
			path: "/api/app", middleware: middlewares.AuthenticateForAppMiddleware(appEnv),
			handler: actions.AppGetHandler, allowedMethods: []string{"GET", "OPTIONS"},
		},
		{
			path: "/api/builds/{build_slug}/steps/{step_id}", middleware: middlewares.AuthenticateForAppMiddleware(appEnv),
			handler: actions.StepGetHandler, allowedMethods: []string{"GET", "OPTIONS"},
		},
		{
			path: "/api/builds/{build_slug}/test_summary", middleware: middlewares.AuthenticateForAppMiddleware(appEnv),
			handler: actions.StepGetHandler, allowedMethods: []string{"GET", "OPTIONS"},
		},
		{
			path: "/api/builds/{build_slug}/test_reports", middleware: middlewares.AuthenticateForAppMiddleware(appEnv),
			handler: actions.TestReportsGetHandler, allowedMethods: []string{"GET", "OPTIONS"},
		},
		{
			path: "/api/builds/{build_slug}/test_reports/ftl", middleware: middlewares.AuthenticateForAppMiddleware(appEnv),
			handler: actions.TestSummaryGetHandler, allowedMethods: []string{"GET", "OPTIONS"},
		},
		{
			path: "/api/builds/{build_slug}/test_reports/{test_report_id}", middleware: middlewares.AuthenticateForAppMiddleware(appEnv),
			handler: actions.TestReportGetHandler, allowedMethods: []string{"GET", "OPTIONS"},
		},
	} {
		r.Handle(route.path, route.middleware.Then(actions.Handler{Env: appEnv, H: route.handler})).
			Methods(route.allowedMethods...)
	}

	r.NotFoundHandler = middlewares.CommonMiddleware(appEnv).Then(&handlers.NotFoundHandler{})
	return r
}
