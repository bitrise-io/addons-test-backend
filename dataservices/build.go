package dataservices

import "github.com/bitrise-io/addons-test-backend/models"

// Build ...
type Build interface {
	Create(*models.Build) (*models.Build, []error, error)
	Find(*models.Build) (*models.Build, error)
}
