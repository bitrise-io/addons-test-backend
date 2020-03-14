package services

import (
	"github.com/bitrise-io/addons-test-backend/models"
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

// Find ...
func (s *App) Find(app *models.App) (*models.App, error) {
	var a models.App
	err := s.DB.Where(app).Find(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}
