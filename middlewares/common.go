package middlewares

import (
	"github.com/bitrise-io/addons-test-backend/env"
	"github.com/bitrise-io/api-utils/middleware"
	"github.com/justinas/alice"
)

// CommonMiddleware ...
func CommonMiddleware(appEnv *env.AppEnv) alice.Chain {
	baseMiddleware := middleware.CommonMiddleware()

	if appEnv.Environment == env.ServerEnvProduction {
		baseMiddleware = baseMiddleware.Append(
			middleware.CreateRedirectToHTTPSMiddleware(),
		)
	}
	return baseMiddleware.Append(
		middleware.CreateOptionsRequestTerminatorMiddleware(),
		setupSession(appEnv),
		setupLogger(appEnv),
	)
}

// AuthenticatedAppMiddleware ...
func AuthenticatedAppMiddleware(appEnv *env.AppEnv) alice.Chain {
	return CommonMiddleware(appEnv).Append(checkAuthenticatedApp(appEnv))
}
