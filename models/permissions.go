package models

type PermissionRequest struct {
	Subject string `json:"subject"`
	Object  string `json:"object"`
	Action  string `json:"action"`
	Effect  string `json:"effect"`
}
