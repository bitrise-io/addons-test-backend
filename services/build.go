package services

import (
	"github.com/bitrise-io/addons-test-backend/models"
	entity "github.com/bitrise-io/api-utils/models"
	"github.com/jinzhu/gorm"
)

// Build ...
type Build struct {
	entity.UpdatableModelService
	DB *gorm.DB
}

// Create ...
func (s *Build) Create(build *models.Build) (*models.Build, []error, error) {
	result := s.DB.Create(build)
	verrs := entity.ValidationErrors(result.GetErrors())
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

// Update ...
func (s *Build) Update(build *models.Build, whitelist []string) ([]error, error) {
	updateData, err := s.UpdateData(*build, whitelist)
	if err != nil {
		return nil, err
	}
	result := s.DB.Model(build).Updates(updateData)
	verrs := entity.ValidationErrors(result.GetErrors())
	if len(verrs) > 0 {
		return verrs, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return nil, nil
}
