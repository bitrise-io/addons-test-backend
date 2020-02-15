package models

import (
	validation "github.com/bitrise-io/api-utils/models"
	"github.com/jinzhu/gorm"
)

// TestReportService ...
type TestReportService struct {
	DB *gorm.DB
}

// Create ...
func (s *TestReportService) Create(testReport *TestReport) (*TestReport, []error, error) {
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
func (s *TestReportService) Find(testReport *TestReport) (*TestReport, error) {
	err := s.DB.Where(testReport).First(testReport).Error
	if err != nil {
		return nil, err
	}
	return testReport, nil
}

// FindAll ...
func (s *TestReportService) FindAll(testReport *TestReport) ([]TestReport, error) {
	var testReports []TestReport
	err := s.DB.Where(testReport).Find(&testReports).Error
	if err != nil {
		return nil, err
	}
	return testReports, nil
}
