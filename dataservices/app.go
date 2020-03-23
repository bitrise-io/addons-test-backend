package dataservices

import "github.com/bitrise-io/addons-test-backend/models"

// App ...
type App interface {
	Create(app *models.App) (*models.App, []error, error)
	Find(*models.App) (*models.App, error)
	Update(*models.App, []string) ([]error, error)
}
