package env

import (
	"fmt"
	"os"

	"github.com/bitrise-io/addons-test-backend/analytics"
	"github.com/bitrise-io/addons-test-backend/dataservices"
	"github.com/bitrise-io/addons-test-backend/firebaseutils"
	"github.com/bitrise-io/addons-test-backend/services"
	"github.com/bitrise-io/addons-test-backend/session"
	"github.com/bitrise-io/api-utils/providers"
	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	// ServerEnvProduction ...
	ServerEnvProduction = "production"
	// ServerEnvDevelopment ...
	ServerEnvDevelopment = "development"
)

// AppEnv ...
type AppEnv struct {
	Port                            string
	Environment                     string
	ShouldSkipSessionAuthentication bool
	Logger                          *zap.Logger
	HostName                        string
	AddonAccessToken                string
	SSOToken                        string
	RequestParams                   providers.RequestParamsInterface
	AnalyticsClient                 analytics.Interface
	SessionCookieStore              *sessions.CookieStore
	SessionName                     string
	Session                         session.Interface
	FirebaseAPI                     *firebaseutils.APIModel
	AppService                      dataservices.App
	BuildService                    dataservices.Build
	TestReportService               dataservices.TestReport
}

// New ...
func New(db *gorm.DB) (*AppEnv, error) {
	var ok bool
	env := &AppEnv{}
	env.Port, ok = os.LookupEnv("PORT")
	if !ok {
		env.Port = "80"
	}
	env.Environment, ok = os.LookupEnv("ENVIRONMENT")
	if !ok {
		env.Environment = ServerEnvDevelopment
	}
	env.ShouldSkipSessionAuthentication = os.Getenv("SKIP_SESSION_AUTH") == "yes"
	env.RequestParams = &providers.RequestParams{}
	analyticsClient, err := analytics.NewClient(env.Logger)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize analytics client")
	}
	env.AnalyticsClient = &analyticsClient
	sessionSecret, ok := os.LookupEnv("SESSION_SECRET")
	if !ok && env.Environment == ServerEnvProduction {
		return nil, errors.New("Session secret must be set in production")
	}

	env.SessionCookieStore = sessions.NewCookieStore([]byte(sessionSecret))
	env.SessionName = "_addons-firebase-testlab_session"

	hostName, err := setRequiredEnv("ADDON_HOST")
	if err != nil {
		return nil, err
	}
	env.HostName = hostName
	accessToken, err := setRequiredEnv("ADDON_ACCESS_TOKEN")
	if err != nil {
		return nil, err
	}
	env.AddonAccessToken = accessToken

	gcKeyJSON, err := setRequiredEnv("SERVICE_ACCOUNT_KEY_JSON")
	if err != nil {
		return nil, err
	}
	gcJWTModel, err := firebaseutils.NewJWTModel(gcKeyJSON)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to create Firebase API model")
	}
	gcBucket, err := setRequiredEnv("BUCKET")
	if err != nil {
		return nil, err
	}
	gcProjectID, err := setRequiredEnv("PROJECT_ID")
	if err != nil {
		return nil, err
	}
	fAPI, err := firebaseutils.New(gcJWTModel, gcProjectID, gcBucket, env.Logger)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to create Firebase API model")
	}
	env.FirebaseAPI = fAPI

	env.AppService = &services.App{DB: db}
	env.BuildService = &services.Build{DB: db}
	env.TestReportService = &services.TestReport{DB: db}
	return env, nil
}

func setRequiredEnv(envKey string) (string, error) {
	envVar := os.Getenv(envKey)
	if envVar == "" {
		return "", fmt.Errorf("Environment Variable missing: %s", envKey)
	}
	return envVar, nil
}
