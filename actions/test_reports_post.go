package actions

import (
	"encoding/json"
	"net/http"

	"github.com/bitrise-io/addons-test-backend/env"
	"github.com/bitrise-io/addons-test-backend/models"
	"github.com/bitrise-io/api-utils/httprequest"
	"github.com/bitrise-io/api-utils/httpresponse"
	"github.com/gobuffalo/nulls"
	"github.com/pkg/errors"
)

// TestReportsPostHandler ...
func TestReportsPostHandler(appEnv *env.AppEnv, w http.ResponseWriter, r *http.Request) error {
	requestParams := appEnv.RequestParams.Get(r)
	appSlug := requestParams["app_slug"]
	buildSlug := requestParams["build_slug"]

	params := testReportPostParams{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		return httpresponse.RespondWithBadRequestError(w, "Failed to decode test report data")
	}
	defer httprequest.BodyCloseWithErrorLog(r)

	stepInfo, err := json.Marshal(params.Step)
	if err != nil {
		return errors.WithMessage(err, "Failed to marshal step info")
	}

	testReport := models.TestReport{
		Name:      params.Name,
		Filename:  params.Filename,
		Filesize:  params.Filesize,
		Step:      stepInfo,
		Uploaded:  false,
		AppSlug:   appSlug,
		BuildSlug: buildSlug,
	}

	tr, verrs, err := appEnv.TestReportService.Create(&models.TestReport{
		Name:      params.Name,
		Filename:  params.Filename,
		Filesize:  params.Filesize,
		Step:      stepInfo,
		Uploaded:  false,
		AppSlug:   appSlug,
		BuildSlug: buildSlug,
	})
	if len(verrs) > 0 {
		return httpresponse.RespondWithUnprocessableEntity(w, verrs)
	}
	if err != nil {
		return errors.WithMessage(err, "SQL Error")
	}

	preSignedURL, err := appEnv.FirebaseAPI.UploadURLforPath(testReport.PathInBucket())
	if err != nil {
		return errors.WithMessage(err, "Failed to create upload url")
	}

	testReportWithUploadURL := newTestReportWithUploadURL(*tr, preSignedURL)

	testReportAssets := []testReportAssetWithUploadURL{}
	for _, testReportAssetParam := range params.TestReportAssets {
		tra, verrs, err := appEnv.TestReportAssetService.Create(&models.TestReportAsset{
			TestReport:   testReport,
			TestReportID: testReport.ID,
			Filename:     testReportAssetParam.Filename,
			Filesize:     testReportAssetParam.Filesize,
		})
		if len(verrs) > 0 {
			return httpresponse.RespondWithUnprocessableEntity(w, verrs)
		}
		if err != nil {
			return errors.WithMessage(err, "SQL Error")
		}
		preSignedURL, err := appEnv.FirebaseAPI.UploadURLforPath(tra.PathInBucket())
		if err != nil {
			return errors.WithMessage(err, "Failed to create upload url")
		}
		testReportAssets = append(testReportAssets, testReportAssetWithUploadURL{
			TestReportAsset: *tra,
			UploadURL:       preSignedURL,
		})
	}

	response := testReportPostResponse{
		testReportWithUploadURL: testReportWithUploadURL,
		TestReportAssets:        testReportAssets,
	}

	return httpresponse.RespondWithCreated(w, response)
}

type testReportAssetPostParams struct {
	Filename string `json:"filename"`
	Filesize int    `json:"filesize"`
}

type testReportPostParams struct {
	Name             string                      `json:"name"`
	Filename         string                      `json:"filename"`
	Filesize         int                         `json:"filesize"`
	Step             models.StepInfo             `json:"step"`
	TestReportAssets []testReportAssetPostParams `json:"assets"`
}

type testReportPatchParams struct {
	Name     string     `json:"name"`
	Uploaded nulls.Bool `json:"uploaded"`
}

type testReportWithUploadURL struct {
	models.TestReport
	UploadURL string `json:"upload_url"`
}

type testReportAssetWithUploadURL struct {
	models.TestReportAsset
	UploadURL string `json:"upload_url"`
}

type testReportPostResponse struct {
	testReportWithUploadURL
	TestReportAssets []testReportAssetWithUploadURL `json:"assets"`
}

func newTestReportWithUploadURL(testReport models.TestReport, uploadURL string) testReportWithUploadURL {
	return testReportWithUploadURL{
		testReport,
		uploadURL,
	}
}
