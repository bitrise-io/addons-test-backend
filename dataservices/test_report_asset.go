package dataservices

import "github.com/bitrise-io/addons-test-backend/models"

// TestReportAsset ...
type TestReportAsset interface {
	Create(*models.TestReportAsset) (*models.TestReportAsset, []error, error)
}
