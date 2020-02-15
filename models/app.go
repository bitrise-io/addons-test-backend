package models

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// App ...
type App struct {
	ID                uuid.UUID `json:"id" db:"id"`
	Plan              string    `json:"plan" db:"plan"`
	EncryptedSecret   []byte    `json:"-" db:"encrypted_secret"`
	EncryptedSecretIV []byte    `json:"-" db:"encrypted_secret_iv"`
	AppSlug           string    `json:"app_slug" db:"app_slug"`
	BitriseAPIToken   string    `json:"-" db:"bitrise_api_token"` // to have authentication when making requests to Bitrise API
	APIToken          string    `json:"api_token" db:"api_token"` // to authenticate incoming requests from running builds
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// BeforeCreate ...
func (a *App) BeforeCreate(scope *gorm.Scope) error {
	if uuid.Equal(a.ID, uuid.UUID{}) {
		a.ID = uuid.NewV4()
	}
	t := time.Now()
	a.CreatedAt = t
	a.UpdatedAt = t
	return nil
}

// BeforeSave ...
func (a *App) BeforeSave(scope *gorm.Scope) error {
	var err error
	if len(a.Plan) == 0 {
		err = scope.DB().AddError(NewValidationError("plan: cannot be empty"))
	}
	if len(a.AppSlug) == 0 {
		err = scope.DB().AddError(NewValidationError("app_slug: cannot be empty"))
	}
	if err != nil {
		return errors.New("Validation failed")
	}
	return nil
}
