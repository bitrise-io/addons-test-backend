package models

import (
	"fmt"
	"time"

	validation "github.com/bitrise-io/api-utils/models"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// TestReportAsset ...
type TestReportAsset struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Filename     string     `json:"filename" db:"filename"`
	Filesize     int        `json:"filesize" db:"filesize"`
	Uploaded     bool       `json:"uploaded" db:"uploaded"`
	TestReport   TestReport `belongs_to:"test_report" json:"-" db:"-"`
	TestReportID uuid.UUID  `json:"test_report_id" db:"test_report_id"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// BeforeCreate ...
func (tra *TestReportAsset) BeforeCreate(scope *gorm.Scope) error {
	if uuid.Equal(tra.ID, uuid.UUID{}) {
		tra.ID = uuid.NewV4()
	}
	t := time.Now()
	tra.CreatedAt = t
	tra.UpdatedAt = t
	return nil
}

// BeforeSave ...
func (tra *TestReportAsset) BeforeSave(scope *gorm.Scope) error {
	var err error
	if len(tra.Filename) == 0 {
		err = scope.DB().AddError(validation.NewValidationError("filename: cannot be empty"))
	}
	if tra.Filesize > 0 {
		err = scope.DB().AddError(validation.NewValidationError("filesize: must be greater than 0"))
	}
	if err != nil {
		return errors.New("Validation failed")
	}
	return nil
}

// PathInBucket ...
func (tra *TestReportAsset) PathInBucket() string {
	return fmt.Sprintf("builds/%s/test_reports/%s/assets/%s", tra.TestReport.BuildSlug, tra.TestReportID, tra.Filename)
}
