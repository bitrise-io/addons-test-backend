package firebaseutils

import (
	"net/http"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	iam "google.golang.org/api/iam/v1"
)

// APIModel ...
type APIModel struct {
	ProjectID string
	Bucket    string
	JWT       *JWTModel
}

// JWTModel ...
type JWTModel struct {
	Client *http.Client
	Config *jwt.Config
}

// TestAsset describes a requested test asset
type TestAsset struct {
	UploadURL string `json:"uploadUrl"`
	GcsPath   string `json:"gcsPath"`
	Filename  string `json:"filename"`
}

// TestAssetsAndroid describes needed Android test asset and is used to return Android test asset upload URLs
type TestAssetsAndroid struct {
	Apk        TestAsset   `json:"apk,omitempty"`
	Aab        TestAsset   `json:"aab,omitmepty"`
	TestApk    TestAsset   `json:"testApk,omitempty"`
	RoboScript TestAsset   `json:"roboScript,omitempty"`
	ObbFiles   []TestAsset `json:"obbFiles,omitempty"`
}

// UploadURLRequest ...
type UploadURLRequest struct {
	AppURL     string `json:"appUrl,omitempty"`
	TestAppURL string `json:"testAppUrl,omitempty"`
}

// MetricSampleModel ...
type MetricSampleModel struct {
	CPU         map[string]float64 `json:"cpu_samples"`
	RAM         map[string]float64 `json:"ram_samples"`
	NetworkDown map[string]float64 `json:"nwd_samples"`
	NetworkUp   map[string]float64 `json:"nwu_samples"`
}

// NewJWTModel ...
func NewJWTModel(gcKeyJSON string) (*JWTModel, error) {
	if gcKeyJSON == "" {
		return nil, errors.New("GC key JSON is empty")
	}
	config, err := google.JWTConfigFromJSON([]byte(gcKeyJSON), iam.CloudPlatformScope, "https://www.googleapis.com/auth/firebase")
	if err != nil {
		return nil, err
	}

	client := config.Client(oauth2.NoContext)

	return &JWTModel{Config: config, Client: client}, nil
}
