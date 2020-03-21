package services

import (
	"github.com/bitrise-io/addons-test-backend/models"
	validation "github.com/bitrise-io/api-utils/models"
	"github.com/jinzhu/gorm"
)

// Build ...
type Build struct {
	DB *gorm.DB
}

// Create ...
func (s *Build) Create(build *models.Build) (*models.Build, []error, error) {
	result := s.DB.Create(build)
	verrs := validation.ValidationErrors(result.GetErrors())
	if len(verrs) > 0 {
		return nil, verrs, nil
	}
	if result.Error != nil {
		return nil, nil, result.Error
	}
	return build, nil, nil
}

// Find ...
func (s *Build) Find(build *models.Build) (*models.Build, error) {
	var b models.Build
	err := s.DB.Where(build).First(&b).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}
