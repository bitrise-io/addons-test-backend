package models

// Test ...
type Test struct {
	DeviceName   string         `json:"device_name,omitempty"`
	APILevel     string         `json:"api_level,omitempty"`
	Status       string         `json:"status,omitempty"` //pending,inProgress,complete
	TestResults  []TestResults  `json:"test_results,omitempty"`
	Outcome      string         `json:"outcome,omitempty"` //failure,inconclusive,success,skipped?
	Orientation  string         `json:"orientation,omitempty"`
	Locale       string         `json:"locale,omitempty"`
	StepID       string         `json:"step_id,omitempty"`
	OutputURLs   OutputURLModel `json:"output_urls,omitempty"`
	TestType     string         `json:"test_type,omitempty"`
	TestIssues   []TestIssue    `json:"test_issues,omitempty"`
	StepDuration int            `json:"step_duration_in_seconds,omitempty"`
}

// TestIssue ...
type TestIssue struct {
	Name       string `json:"name,omitempty"`
	Summary    string `json:"summary,omitempty"`
	Stacktrace string `json:"stacktrace,omitempty"`
}

// OutputURLModel ...
type OutputURLModel struct {
	ScreenshotURLs  []string          `json:"screenshot_urls,omitempty"`
	VideoURL        string            `json:"video_url,omitempty"`
	ActivityMapURL  string            `json:"activity_map_url,omitempty"`
	TestSuiteXMLURL string            `json:"test_suite_xml_url,omitempty"`
	LogURLs         []string          `json:"log_urls,omitempty"`
	AssetURLs       map[string]string `json:"asset_urls,omitempty"`
}

// TestResults ...
type TestResults struct {
	Skipped int `json:"in_progress,omitempty"`
	Failed  int `json:"failed,omitempty"`
	Total   int `json:"total,omitempty"`
}
