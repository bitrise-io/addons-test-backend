package models

// GitData ...
type GitData struct {
	Provider      string `json:"provider"`
	SrcBranch     string `json:"src_branch"`
	DstBranch     string `json:"dst_branch"`
	PullRequestID int    `json:"pull_request_id"`
}
