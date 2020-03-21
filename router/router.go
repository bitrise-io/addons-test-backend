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
		{
			path: "/api/app", middleware: middlewares.AuthenticatedAppMiddleware(appEnv),
			handler: actions.AppGetHandler, allowedMethods: []string{"GET", "OPTIONS"},
		},
	} {
		r.Handle(route.path, route.middleware.Then(actions.Handler{Env: appEnv, H: route.handler})).
			Methods(route.allowedMethods...)
	}

	r.NotFoundHandler = middlewares.CommonMiddleware(appEnv).Then(&handlers.NotFoundHandler{})
	return r
}
