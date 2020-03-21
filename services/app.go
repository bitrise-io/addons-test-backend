package services

import (
	"github.com/bitrise-io/addons-test-backend/models"
	validation "github.com/bitrise-io/api-utils/models"
	"github.com/jinzhu/gorm"
)

// AppInterface ...
type AppInterface interface {
	Find(*models.App) (*models.App, error)
}

// App ...
type App struct {
	DB *gorm.DB
}

// Create ...
func (s *App) Create(app *models.App) (*models.App, []error, error) {
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
func (s *App) Find(app *models.App) (*models.App, error) {
	var a models.App
	err := s.DB.Where(app).First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}
