package models

// AppData ...
type AppData struct {
	AppSlug                string  `json:"app_slug"`
	BuildSlug              string  `json:"build_slug"`
	BuildNumber            int     `json:"build_number"`
	BuildStatus            int     `json:"build_status"`
	BuildTriggeredWorkflow string  `json:"build_triggered_workflow"`
	Git                    GitData `json:"git"`
}
