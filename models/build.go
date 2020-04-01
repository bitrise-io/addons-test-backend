package models

import (
	"time"

	validation "github.com/bitrise-io/api-utils/models"
	"github.com/gobuffalo/nulls"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// Build ...
type Build struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	AppSlug             string     `json:"app_slug" db:"app_slug"`
	BuildSlug           string     `json:"build_slug" db:"build_slug"`
	BuildSessionEnabled bool       `json:"build_session_enabled" db:"build_session_enabled"`
	TestStartTime       nulls.Time `json:"test_start_time" db:"test_start_time"`
	TestEndTime         nulls.Time `json:"test_end_time" db:"test_end_time"`
	TestMatrixID        string     `json:"test_matrix_id" db:"test_matrix_id"`
	TestHistoryID       string     `json:"test_history_id" db:"test_history_id"`
	TestExecutionID     string     `json:"test_execution_id" db:"test_execution_id"`
	LastRequest         nulls.Time `json:"last_request" db:"last_request"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}

// BeforeCreate ...
func (b *Build) BeforeCreate(scope *gorm.Scope) error {
	if uuid.Equal(b.ID, uuid.UUID{}) {
		b.ID = uuid.NewV4()
	}
	b.CreatedAt = time.Now()
	return nil
}

// BeforeSave ...
func (b *Build) BeforeSave(scope *gorm.Scope) error {
	var err error
	if len(b.AppSlug) == 0 {
		err = scope.DB().AddError(validation.NewValidationError("app_slug: cannot be empty"))
	}
	if len(b.BuildSlug) == 0 {
		err = scope.DB().AddError(validation.NewValidationError("build_slug: cannot be empty"))
	}
	if err != nil {
		return errors.New("Validation failed")
	}
	b.UpdatedAt = time.Now()
	return nil
}
