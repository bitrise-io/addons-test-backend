package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/bitrise-io/addons-test-backend/env"
	"github.com/bitrise-io/addons-test-backend/router"
	"github.com/bitrise-io/api-utils/database"
	"github.com/bitrise-io/api-utils/logging"
	"go.uber.org/zap"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// "github.com/bitrise-io/api-utils/database"

func main() {
	logger := logging.WithContext(nil)
	defer func() {
		err := logger.Sync()
		if err != nil {
			fmt.Printf("Failed to sync logger: %#v", err)
		}
	}()

	tracer.Start(tracer.WithServiceName("addons-test"))
	defer tracer.Stop()

	postgresDB := database.PostgresDatabase{}
	err := postgresDB.InitializeConnection(true)
	if err != nil {
		logger.Error("Failed to initialize DB connection", zap.Any("error", err))
		os.Exit(1)
	}
	defer postgresDB.Close()
	log.Println(" [OK] Database connection established")

	appEnv, err := env.New(postgresDB.GetDB())
	if err != nil {
		logger.Error("Failed to initialize Application Environment object", zap.Any("error", err))
		os.Exit(1)
	}

	// Routing
	http.Handle("/", router.New(appEnv))

	log.Println("Starting - using port:", appEnv.Port)
	if err := http.ListenAndServe(":"+appEnv.Port, nil); err != nil {
		logger.Error("Failed to initialize Ship Addon Backend", zap.Any("error", err))
	}
}
