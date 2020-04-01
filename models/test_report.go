package models

import (
	"encoding/json"
	"fmt"
	"time"

	validation "github.com/bitrise-io/api-utils/models"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// TestReport ...
type TestReport struct {
	ID               uuid.UUID        `json:"id" db:"id"`
	Name             string           `json:"name" db:"name"`
	Filename         string           `json:"filename" db:"filename"`
	Filesize         int              `json:"filesize" db:"filesize"`
	Step             json.RawMessage  `json:"step" db:"step"`
	Uploaded         bool             `json:"uploaded" db:"uploaded"`
	AppSlug          string           `json:"app_slug" db:"app_slug"`
	BuildSlug        string           `json:"build_slug" db:"build_slug"`
	CreatedAt        time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time        `json:"-" db:"updated_at"`
	TestReportAssets TestReportAssets `has_many:"test_report_assets" json:"assets" db:"-"`
}

// TestReportAssets ...
type TestReportAssets []TestReportAsset

// BeforeCreate ...
func (tr *TestReport) BeforeCreate(scope *gorm.Scope) error {
	if uuid.Equal(tr.ID, uuid.UUID{}) {
		tr.ID = uuid.NewV4()
	}
	tr.CreatedAt = time.Now()
	return nil
}

// BeforeSave ...
func (tr *TestReport) BeforeSave(scope *gorm.Scope) error {
	var err error
	if len(tr.Filename) == 0 {
		err = scope.DB().AddError(validation.NewValidationError("filename: cannot be empty"))
	}
	if tr.Filesize > 0 {
		err = scope.DB().AddError(validation.NewValidationError("filesize: must be greater than 0"))
	}
	if err != nil {
		return errors.New("Validation failed")
	}
	tr.UpdatedAt = time.Now()
	return nil
}

// PathInBucket ...
func (tr *TestReport) PathInBucket() string {
	return fmt.Sprintf("builds/%s/test_reports/%s/%s", tr.BuildSlug, tr.ID, tr.Filename)
}
