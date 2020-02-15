package models

import (
	validation "github.com/bitrise-io/api-utils/models"
	"github.com/jinzhu/gorm"
)

// BuildService ...
type BuildService struct {
	DB *gorm.DB
}

// Create ...
func (s *BuildService) Create(build *Build) (*Build, []error, error) {
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
func (s *BuildService) Find(build *Build) (*Build, error) {
	err := s.DB.Where(build).First(build).Error
	if err != nil {
		return nil, err
	}
	return build, nil
}
