package firebaseutils

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"google.golang.org/api/googleapi"
	storage "google.golang.org/api/storage/v1"

	storagesu "cloud.google.com/go/storage"
	"github.com/bitrise-io/addons-firebase-testlab/configs"
	"github.com/bitrise-io/addons-firebase-testlab/metrics"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/sliceutil"
	testing "google.golang.org/api/testing/v1"
	toolresults "google.golang.org/api/toolresults/v1beta3"
)

// DevicesCatalog ...
var DevicesCatalog *testing.TestEnvironmentCatalog

const (
	sampleTypeCPU    = 0
	sampleTypeRAM    = 7
	sampleTypeNWUp   = 9
	sampleTypeNWDown = 10
)

// New ...
func New() (*APIModel, error) {
	return &APIModel{
		JWT: configs.GetJWTModel(),
	}, nil
}

//
// TOOLS

func getAndroidVirtualDeviceIDs(devices []*testing.AndroidModel) []string {
	deviceIDs := []string{}
	for _, device := range devices {
		if device.Form == "VIRTUAL" {
			deviceIDs = append(deviceIDs, device.Id)
		}
	}
	return deviceIDs
}

func getIosDeviceIDs(devices []*testing.IosModel) []string {
	deviceIDs := []string{}
	for _, device := range devices {
		deviceIDs = append(deviceIDs, device.Id)
	}
	return deviceIDs
}

// ValidateAndroidDevices ...
func ValidateAndroidDevices(devices []*testing.AndroidDevice) error {
	for _, device := range devices {
		deviceID := device.AndroidModelId
		versionID := device.AndroidVersionId

		deviceFound := false
		for _, catalogDevice := range DevicesCatalog.AndroidDeviceCatalog.Models {
			if catalogDevice.Id == deviceID {
				if catalogDevice.Form != "VIRTUAL" {
					return fmt.Errorf("(%s) is not a virtual device. Available virtual devices: %v", deviceID, getAndroidVirtualDeviceIDs(DevicesCatalog.AndroidDeviceCatalog.Models))
				}
				if !sliceutil.IsStringInSlice(versionID, catalogDevice.SupportedVersionIds) {
					return fmt.Errorf("device (%s) has no version: %s. Available versions: %v", deviceID, versionID, catalogDevice.SupportedVersionIds)
				}
				deviceFound = true
			}
		}
		if !deviceFound {
			return fmt.Errorf("device (%s) not found. Available devices: %v", deviceID, getAndroidVirtualDeviceIDs(DevicesCatalog.AndroidDeviceCatalog.Models))
		}
	}
	return nil
}

// ValidateIosDevices ...
func ValidateIosDevices(devices []*testing.IosDevice) error {
	for _, device := range devices {
		deviceID := device.IosModelId
		versionID := device.IosVersionId

		deviceFound := false
		for _, catalogDevice := range DevicesCatalog.IosDeviceCatalog.Models {
			if catalogDevice.Id == deviceID {
				if !sliceutil.IsStringInSlice(versionID, catalogDevice.SupportedVersionIds) {
					return fmt.Errorf("device (%s) has no version: %s. Available versions: %v", deviceID, versionID, catalogDevice.SupportedVersionIds)
				}
				deviceFound = true
			}
		}
		if !deviceFound {
			return fmt.Errorf("device (%s) not found. Available devices: %v", deviceID, getIosDeviceIDs(DevicesCatalog.IosDeviceCatalog.Models))
		}
	}
	return nil
}

//
// TESTING API

// CancelTestMatrix ...
func CancelTestMatrix(matrixID string) (string, error) {
	testingService, err := testing.New(configs.GetJWTModel().Client)
	if err != nil {
		return "", err
	}

	cancelCall := testingService.Projects.TestMatrices.Cancel(configs.GetProjectID(), matrixID)
	cancelResp, err := cancelCall.Do()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%+v", cancelResp.ServerResponse), nil
}

// GetTagArray ...
func (api *APIModel) GetTagArray() []string {
	return []string{}
}

// GetProfileName ...
func (api *APIModel) GetProfileName() string {
	return "FirebaseAPI"
}

// StartTestMatrix ...
func (api *APIModel) StartTestMatrix(appSlug, buildSlug string, testMatrix *testing.TestMatrix) (*testing.TestMatrix, error) {
	tracker := metrics.NewDogStatsDMetrics("")
	defer tracker.Close()

	testingService, err := testing.New(configs.GetJWTModel().Client)
	if err != nil {
		return nil, fmt.Errorf("Failed to create testing service, error: %s", err)
	}

	toolresultsService, err := toolresults.New(configs.GetJWTModel().Client)
	if err != nil {
		return nil, fmt.Errorf("Failed to create toolresults service, error: %s", err)
	}

	listHistory := toolresultsService.Projects.Histories.List(configs.GetProjectID())
	listHistory = listHistory.FilterByName(appSlug)
	histories, err := listHistory.Do()
	if err != nil {
		return nil, fmt.Errorf("Failed to list histories, error: %s", err)
	}
	tracker.Track(api, "numberOfOutgoingRequests", fmt.Sprintf("appSlug:%s", appSlug), fmt.Sprintf("buildSlug:%s", buildSlug))

	historyID := ""

	if histories != nil && histories.Histories != nil {
		for _, hist := range histories.Histories {
			if hist != nil && hist.Name == appSlug {
				historyID = hist.HistoryId
				break
			}
		}
	}

	if historyID == "" {
		createHistory := toolresultsService.Projects.Histories.Create(configs.GetProjectID(), &toolresults.History{DisplayName: appSlug, Name: appSlug})
		newHistory, err := createHistory.Do()
		if err != nil {
			return nil, fmt.Errorf("Failed to create history, error: %s", err)
		}
		tracker.Track(api, "numberOfOutgoingRequests", fmt.Sprintf("appSlug:%s", appSlug), fmt.Sprintf("buildSlug:%s", buildSlug))
		historyID = newHistory.HistoryId
	}

	if testMatrix.TestSpecification.AndroidInstrumentationTest != nil &&
		(testMatrix.TestSpecification.AndroidInstrumentationTest.AppApk == nil && testMatrix.TestSpecification.AndroidInstrumentationTest.AppBundle == nil) {
		testMatrix.TestSpecification.AndroidInstrumentationTest.AppApk = &testing.FileReference{GcsPath: getAppBucketPath(buildSlug, "app.apk")}
		testMatrix.TestSpecification.AndroidInstrumentationTest.TestApk = &testing.FileReference{GcsPath: getAppBucketPath(buildSlug, "app-test.apk")}
	}
	if testMatrix.TestSpecification.AndroidRoboTest != nil &&
		(testMatrix.TestSpecification.AndroidRoboTest.AppApk == nil && testMatrix.TestSpecification.AndroidRoboTest.AppBundle == nil) {
		testMatrix.TestSpecification.AndroidRoboTest.AppApk = &testing.FileReference{GcsPath: getAppBucketPath(buildSlug, "app.apk")}
	}
	if testMatrix.TestSpecification.AndroidTestLoop != nil &&
		(testMatrix.TestSpecification.AndroidTestLoop.AppApk == nil && testMatrix.TestSpecification.AndroidTestLoop.AppBundle == nil) {
		testMatrix.TestSpecification.AndroidTestLoop.AppApk = &testing.FileReference{GcsPath: getAppBucketPath(buildSlug, "app-test.apk")}
	}

	if testMatrix.TestSpecification.IosXcTest != nil {
		testMatrix.TestSpecification.IosXcTest.TestsZip = &testing.FileReference{GcsPath: getAppBucketPath(buildSlug, "app.apk")}
	}

	testMatrix.ResultStorage = &testing.ResultStorage{GoogleCloudStorage: &testing.GoogleCloudStorage{GcsPath: getResultsBucketPath(buildSlug)}, ToolResultsHistory: &testing.ToolResultsHistory{ProjectId: configs.GetProjectID(), HistoryId: historyID}}

	testMatrixCall := testingService.Projects.TestMatrices.Create(configs.GetProjectID(), testMatrix)
	testMatrixCall.RequestId(buildSlug)

	//get only required fields
	testMatrixCall.Fields("testMatrixId", "timestamp")

	responseMatrix, err := testMatrixCall.Do()
	if err != nil {
		return nil, fmt.Errorf("Failed to create testing service, error: %s", err)
	}
	tracker.Track(api, "numberOfOutgoingRequests", fmt.Sprintf("appSlug:%s", appSlug), fmt.Sprintf("buildSlug:%s", buildSlug))
	tracker.Track(api, "numberOfTests", fmt.Sprintf("appSlug:%s", appSlug))

	return responseMatrix, nil
}

// GetHistoryAndExecutionIDByMatrixID ...
func (api *APIModel) GetHistoryAndExecutionIDByMatrixID(id string) (*testing.TestMatrix, error) {
	testingService, err := testing.New(configs.GetJWTModel().Client)
	if err != nil {
		return nil, err
	}

	matricesCall := testingService.Projects.TestMatrices.Get(configs.GetProjectID(), id)
	matricesCall.Fields("invalidMatrixDetails,state,testExecutions(toolResultsStep(historyId,executionId))")
	return matricesCall.Do()
}

// GetDeviceCatalog ...
func (api *APIModel) GetDeviceCatalog() (*testing.TestEnvironmentCatalog, error) {
	testEnvCatalog := &testing.TestEnvironmentCatalog{}
	testingService, err := testing.New(configs.GetJWTModel().Client)
	if err != nil {
		return nil, err
	}

	androidCatalogCall := testingService.TestEnvironmentCatalog.Get("ANDROID")
	testEnvCatalog, err = androidCatalogCall.Do()
	if err != nil {
		return nil, err
	}

	iosCatalogCall := testingService.TestEnvironmentCatalog.Get("iOS")
	iosCatalog, err := iosCatalogCall.Do()
	if err != nil {
		return nil, err
	}
	testEnvCatalog.IosDeviceCatalog = iosCatalog.IosDeviceCatalog

	return testEnvCatalog, nil
}

// GetDeviceNameByID ...
func GetDeviceNameByID(id string) string {
	for _, model := range DevicesCatalog.AndroidDeviceCatalog.Models {
		if model.Id == id {
			return model.Name
		}
	}
	for _, model := range DevicesCatalog.IosDeviceCatalog.Models {
		if model.Id == id {
			return model.Name
		}
	}
	return id
}

// GetLangByCountryCode ...
func GetLangByCountryCode(countryCode string) string {
	for _, locale := range DevicesCatalog.AndroidDeviceCatalog.RuntimeConfiguration.Locales {
		if locale.Id == countryCode {
			if locale.Region != "" {
				return fmt.Sprintf("%s (%s)", locale.Name, locale.Region)
			}
			return locale.Name
		}
	}
	return countryCode
}

//
// STORAGE

// TestAssetsUploadURLsAndroid returns a signed URLs used to Android upload asset
func (api *APIModel) TestAssetsUploadURLsAndroid(buildSlug string, assetRequests TestAssetsAndroid) (TestAssetsAndroid, error) {
	var assets TestAssetsAndroid

	getAssetURL := func(origFilename, buildSlug, gcsFilename string) (TestAsset, error) {
		uploadURL, err := storagesu.SignedURL(configs.GetGCSBucket(),
			getTrimmedAppBucketPath(buildSlug, gcsFilename),
			api.GetSignedURLCredentials("PUT"))
		if err != nil {
			return TestAsset{}, err
		}

		return TestAsset{
			UploadURL: uploadURL,
			GcsPath:   getAppBucketPath(buildSlug, gcsFilename),
			Filename:  origFilename,
		}, nil
	}

	if assetRequests.Apk.Filename != "" {
		asset, err := getAssetURL(assetRequests.Apk.Filename, buildSlug, "app.apk")
		if err != nil {
			return TestAssetsAndroid{}, err
		}

		assets.Apk = asset
	}

	if assetRequests.Aab.Filename != "" {
		asset, err := getAssetURL(assetRequests.Aab.Filename, buildSlug, "app.aab")
		if err != nil {
			return TestAssetsAndroid{}, err
		}

		assets.Aab = asset
	}

	if assetRequests.TestApk.Filename != "" {
		asset, err := getAssetURL(assetRequests.TestApk.Filename, buildSlug, "app-test.apk")
		if err != nil {
			return TestAssetsAndroid{}, err
		}

		assets.TestApk = asset
	}

	if assetRequests.RoboScript.Filename != "" {
		asset, err := getAssetURL(assetRequests.RoboScript.Filename, buildSlug, "roboscript.json")
		if err != nil {
			return TestAssetsAndroid{}, err
		}

		assets.RoboScript = asset
	}

	for i, obbFile := range assetRequests.ObbFiles {
		asset, err := getAssetURL(obbFile.Filename, buildSlug, fmt.Sprintf("%d.obb", i))
		if err != nil {
			return TestAssetsAndroid{}, err
		}

		assets.ObbFiles = append(assets.ObbFiles, asset)
	}

	return assets, nil
}

// UploadTestAssets ...
func (api *APIModel) UploadTestAssets(buildSlug string) (*UploadURLRequest, error) {
	apkUploadURL, err := storagesu.SignedURL(configs.GetGCSBucket(), getTrimmedAppBucketPath(buildSlug, "app.apk"), api.GetSignedURLCredentials("PUT"))
	if err != nil {
		return nil, err
	}

	testApkUploadURL, err := storagesu.SignedURL(configs.GetGCSBucket(), getTrimmedAppBucketPath(buildSlug, "app-test.apk"), api.GetSignedURLCredentials("PUT"))
	if err != nil {
		return nil, err
	}

	return &UploadURLRequest{
		AppURL:     apkUploadURL,
		TestAppURL: testApkUploadURL,
	}, nil
}

// UploadURLforPath ...
func (api *APIModel) UploadURLforPath(path string) (string, error) {
	uploadURL, err := storagesu.SignedURL(configs.GetGCSBucket(), path, api.GetSignedURLCredentials("PUT"))
	if err != nil {
		return "", err
	}

	return uploadURL, nil
}

// DownloadURLforPath ...
func (api *APIModel) DownloadURLforPath(path string) (string, error) {
	downloadURL, err := storagesu.SignedURL(configs.GetGCSBucket(), path, api.GetSignedURLCredentials("GET"))
	if err != nil {
		return "", err
	}

	return downloadURL, nil
}

// GetSignedURLCredentials ...
func (api *APIModel) GetSignedURLCredentials(method string) *storagesu.SignedURLOptions {
	return &storagesu.SignedURLOptions{
		GoogleAccessID: api.JWT.Config.Email,
		PrivateKey:     api.JWT.Config.PrivateKey,
		Method:         method,
		Expires:        time.Now().Add(time.Minute * 30),
	}
}

// GetSignedURLOfLegacyBucketPath ...
func (api *APIModel) GetSignedURLOfLegacyBucketPath(path string) (string, error) {
	normlizedPath := strings.TrimPrefix(path, "gs://")
	normlizedPath = strings.TrimPrefix(normlizedPath, configs.GetGCSBucket())
	normlizedPath = strings.TrimPrefix(normlizedPath, "/")

	resultFileDownloadURL, err := storagesu.SignedURL(configs.GetGCSBucket(), normlizedPath, api.GetSignedURLCredentials("GET"))
	if err != nil {
		return "", err
	}

	return resultFileDownloadURL, nil
}

func getTrimmedAppBucketPath(buildSlug string, appFileName string) string {
	return fmt.Sprintf("android-tests/%s/%s", buildSlug, appFileName)
}

func getResultsBucketPath(buildSlug string) string {
	return fmt.Sprintf("gs://%s/android-tests/%s/results/", configs.GetGCSBucket(), buildSlug)
}

func getAppBucketPath(buildSlug string, appFileName string) string {
	return fmt.Sprintf("gs://%s/android-tests/%s/%s", configs.GetGCSBucket(), buildSlug, appFileName)
}

//
// TOOLRESULTS

// GetTestsByHistoryAndExecutionID ...
func (api *APIModel) GetTestsByHistoryAndExecutionID(historyID, executionID, appSlug, buildSlug string, fields ...googleapi.Field) (*toolresults.ListStepsResponse, error) {
	tracker := metrics.NewDogStatsDMetrics("")
	defer tracker.Close()

	resultsService, err := toolresults.New(configs.GetJWTModel().Client)
	if err != nil {
		return nil, err
	}

	stepsCall := resultsService.Projects.Histories.Executions.Steps.List(configs.GetProjectID(), historyID, executionID)
	stepsCall.PageSize(50)
	if len(fields) > 0 {
		stepsCall.Fields(fields...)
	}
	steps, err := stepsCall.Do()
	if err != nil {
		return nil, err
	}
	tracker.Track(api, "numberOfOutgoingRequests", fmt.Sprintf("appSlug:%s", appSlug), fmt.Sprintf("buildSlug:%s", buildSlug))

	return steps, nil
}

// GetTestMetricSamples ...
func (api *APIModel) GetTestMetricSamples(historyID, executionID, stepID, appSlug, buildSlug string) (MetricSampleModel, error) {
	types := []int{sampleTypeCPU, sampleTypeRAM, sampleTypeNWUp, sampleTypeNWDown}

	errChannel := make(chan error, 1)
	var wg sync.WaitGroup
	wg.Add(len(types))

	metricSamples := MetricSampleModel{CPU: map[string]float64{}, RAM: map[string]float64{}, NetworkDown: map[string]float64{}, NetworkUp: map[string]float64{}}

	toolresultsService, err := toolresults.New(configs.GetJWTModel().Client)
	if err != nil {
		return MetricSampleModel{}, fmt.Errorf("Failed to create toolresults service, error: %s", err)
	}

	for _, typeID := range types {
		go func(id int) {
			defer func() {
				wg.Done()
			}()
			tracker := metrics.NewDogStatsDMetrics("")
			defer tracker.Close()

			samplesListCall := toolresultsService.Projects.Histories.Executions.Steps.PerfSampleSeries.Samples.List(configs.GetProjectID(), historyID, executionID, stepID, fmt.Sprintf("%d", id))

			perfSamplesResponse, err := samplesListCall.Do()
			if err != nil {
				if len(errChannel) == 0 {
					errChannel <- fmt.Errorf("Failed to get response for perf samples, error: %s", err)
				}
				return
			}
			tracker.Track(api, "numberOfOutgoingRequests", fmt.Sprintf("appSlug:%s", appSlug), fmt.Sprintf("buildSlug:%s", buildSlug))

			samples := MetricSampleModel{}

			data := map[string]float64{}
			initialSeconds := perfSamplesResponse.PerfSamples[0].SampleTime.Seconds
			for _, sample := range perfSamplesResponse.PerfSamples {
				data[fmt.Sprintf("%f", time.Unix(sample.SampleTime.Seconds-initialSeconds, sample.SampleTime.Nanos).Sub(time.Unix(0, 0)).Seconds())] = sample.Value
			}
			samples.CPU = data

			switch id {
			case sampleTypeCPU:
				metricSamples.CPU = fillUnixTimeDataHash(perfSamplesResponse)
			case sampleTypeRAM:
				metricSamples.RAM = fillUnixTimeDataHash(perfSamplesResponse)
			case sampleTypeNWDown:
				metricSamples.NetworkDown = fillUnixTimeDataHash(perfSamplesResponse)
			case sampleTypeNWUp:
				metricSamples.NetworkUp = fillUnixTimeDataHash(perfSamplesResponse)
			}
		}(typeID)
	}

	wg.Wait()
	close(errChannel)

	err = <-errChannel
	if err != nil {
		log.Errorf("Failed to get perf data, error: %s", err)
		return MetricSampleModel{}, err
	}

	return metricSamples, nil
}

func fillUnixTimeDataHash(metricsData *toolresults.ListPerfSamplesResponse) map[string]float64 {
	data := map[string]float64{}
	initialSeconds := metricsData.PerfSamples[0].SampleTime.Seconds
	for _, sample := range metricsData.PerfSamples {
		data[fmt.Sprintf("%f", time.Unix(sample.SampleTime.Seconds-initialSeconds, sample.SampleTime.Nanos).Sub(time.Unix(0, 0)).Seconds())] = sample.Value
	}
	return data
}

// DownloadTestAssets ...
func (api *APIModel) DownloadTestAssets(buildSlug string) (map[string]string, error) {
	tracker := metrics.NewDogStatsDMetrics("")
	defer tracker.Close()

	storageService, err := storage.New(api.JWT.Client)
	if err != nil {
		return nil, err
	}

	rootObject := getResultsPathPrefix(buildSlug)

	objectsListCall := storageService.Objects.List(configs.GetGCSBucket())
	objectsListCall.Prefix(rootObject)
	objects, err := objectsListCall.Do()
	if err != nil {
		return nil, err
	}
	tracker.Track(api, "numberOfOutgoingRequests", fmt.Sprintf("buildSlug:%s", buildSlug))

	files := map[string]string{}

	for _, obj := range objects.Items {
		trimmedFileName := strings.Replace(strings.TrimPrefix(obj.Name, rootObject), "/", "_", -1)

		resultFileDownloadURL, err := storagesu.SignedURL(configs.GetGCSBucket(), obj.Name, api.GetSignedURLCredentials("GET"))
		if err != nil {
			return nil, err
		}

		files[trimmedFileName] = resultFileDownloadURL
	}

	return files, nil
}

func getResultsPathPrefix(buildSlug string) string {
	return fmt.Sprintf("android-tests/%s/results/", buildSlug)
}
