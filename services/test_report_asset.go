package services

import (
	"github.com/bitrise-io/addons-test-backend/models"
	validation "github.com/bitrise-io/api-utils/models"
	"github.com/jinzhu/gorm"
)

// TestReportAsset ...
type TestReportAsset struct {
	DB *gorm.DB
}

// Create ...
func (s *TestReportAsset) Create(testReportAsset *models.TestReportAsset) (*models.TestReportAsset, []error, error) {
	result := s.DB.Create(testReportAsset)
	verrs := validation.ValidationErrors(result.GetErrors())
	if len(verrs) > 0 {
		return nil, verrs, nil
	}
	if result.Error != nil {
		return nil, nil, result.Error
	}
	return testReportAsset, nil, nil
}
