package env

import (
	"os"

	"github.com/bitrise-io/addons-test-backend/analytics"
	"github.com/bitrise-io/addons-test-backend/dataservices"
	"github.com/bitrise-io/addons-test-backend/models"
	"github.com/bitrise-io/addons-test-backend/session"
	"github.com/bitrise-io/api-utils/logging"
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
	Port               string
	Environment        string
	Logger             *zap.Logger
	SSOToken           string
	RequestParams      providers.RequestParamsInterface
	AnalyticsClient    analytics.Interface
	SessionCookieStore *sessions.CookieStore
	SessionName        string
	Session            session.Interface
	AppService         dataservices.App
	BuildService       dataservices.Build
	TestReportService  dataservices.TestReport
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
	env.Logger = logging.WithContext(nil)
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

	env.AppService = &models.AppService{DB: db}
	env.BuildService = &models.BuildService{DB: db}
	env.TestReportService = &models.TestReportService{DB: db}
	return env, nil
}
