package models

import (
	validation "github.com/bitrise-io/api-utils/models"
	"github.com/jinzhu/gorm"
)

// AppService ...
type AppService struct {
	DB *gorm.DB
}

// Create ...
func (s *AppService) Create(app *App) (*App, []error, error) {
	result := s.DB.Create(app)
	verrs := validation.ValidationErrors(result.GetErrors())
	if len(verrs) > 0 {
		return nil, verrs, nil
	}
	if result.Error != nil {
		return nil, nil, result.Error
	}
	return app, nil, nil
}

// Find ...
func (s *AppService) Find(app *App) (*App, error) {
	err := s.DB.Where(app).First(app).Error
	if err != nil {
		return nil, err
	}
	return app, nil
}
