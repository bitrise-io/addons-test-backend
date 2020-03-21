package services

import (
	"github.com/bitrise-io/addons-test-backend/models"
	validation "github.com/bitrise-io/api-utils/models"
	"github.com/jinzhu/gorm"
)

// TestReport ...
type TestReport struct {
	DB *gorm.DB
}

// Create ...
func (s *TestReport) Create(testReport *models.TestReport) (*models.TestReport, []error, error) {
	result := s.DB.Create(testReport)
	verrs := validation.ValidationErrors(result.GetErrors())
	if len(verrs) > 0 {
		return nil, verrs, nil
	}
	if result.Error != nil {
		return nil, nil, result.Error
	}
	return testReport, nil, nil
}

// Find ...
func (s *TestReport) Find(testReport *models.TestReport) (*models.TestReport, error) {
	var tr models.TestReport
	err := s.DB.Where(testReport).First(&tr).Error
	if err != nil {
		return nil, err
	}
	return &tr, nil
}

// FindAll ...
func (s *TestReport) FindAll(testReport *models.TestReport) ([]models.TestReport, error) {
	var testReports []models.TestReport
	err := s.DB.Where(testReport).Find(&testReports).Error
	if err != nil {
		return nil, err
	}
	return testReports, nil
}
