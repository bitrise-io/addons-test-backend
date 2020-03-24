package dataservices

import "github.com/bitrise-io/addons-test-backend/models"

// TestReport ...
type TestReport interface {
	Create(*models.TestReport) (*models.TestReport, []error, error)
	Find(*models.TestReport) (*models.TestReport, error)
	FindAll(*models.TestReport) ([]models.TestReport, error)
	Update(*models.TestReport, []string) ([]error, error)
}
