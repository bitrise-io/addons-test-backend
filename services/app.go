package services

import (
	"github.com/bitrise-io/addons-test-backend/models"
	entity "github.com/bitrise-io/api-utils/models"
	"github.com/jinzhu/gorm"
)

// App ...
type App struct {
	entity.UpdatableModelService
	DB *gorm.DB
}

// Create ...
func (s *App) Create(app *models.App) (*models.App, []error, error) {
	result := s.DB.Create(app)
	verrs := entity.ValidationErrors(result.GetErrors())
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

// Update ...
func (s *App) Update(app *models.App, whitelist []string) ([]error, error) {
	updateData, err := s.UpdateData(*app, whitelist)
	if err != nil {
		return nil, err
	}
	result := s.DB.Model(app).Updates(updateData)
	verrs := entity.ValidationErrors(result.GetErrors())
	if len(verrs) > 0 {
		return verrs, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return nil, nil
}

// Delete ...
func (s *App) Delete(app *models.App) error {
	result := s.DB.Delete(&app)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected < 1 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
