package models

import "time"

type Status string

const (
	StatusPending   Status = "pending"
	StatusBuilding  Status = "building"
	StatusDeploying Status = "deploying"
	StatusRunning   Status = "running"
	StatusFailed    Status = "failed"
)

type SourceType string

const (
	SourceTypeGit    SourceType = "git"
	SourceTypeUpload SourceType = "upload"
)

type Deployment struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	SourceType  SourceType `json:"source_type"`
	SourceURL   *string    `json:"source_url,omitempty"`
	Status      Status     `json:"status"`
	ImageTag    *string    `json:"image_tag,omitempty"`
	ContainerID *string    `json:"container_id,omitempty"`
	HostPort    *int       `json:"host_port,omitempty"`
	LiveURL     *string    `json:"live_url,omitempty"`
	Error       *string    `json:"error,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type LogLine struct {
	ID           int64     `json:"id"`
	DeploymentID string    `json:"deployment_id"`
	Stream       string    `json:"stream"`
	Line         string    `json:"line"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateDeploymentRequest struct {
	Name       string     `json:"name"`
	SourceType SourceType `json:"source_type"`
	SourceURL  string     `json:"source_url"`
}
