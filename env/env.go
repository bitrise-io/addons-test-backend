package env

import (
	"os"

	"github.com/bitrise-io/addons-test-backend/dataservices"
	"github.com/bitrise-io/addons-test-backend/models"
	"github.com/bitrise-io/api-utils/logging"
	"github.com/jinzhu/gorm"
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
	Port        string
	Environment string
	Logger      *zap.Logger
	AppService  dataservices.App
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

	env.AppService = &models.AppService{DB: db}

	return env, nil
}
