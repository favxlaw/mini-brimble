package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/favxlaw/mini-brimble/models"
)

func (db *DB) CreateDeployment(d *models.Deployment) error {
	_, err := db.Exec(`
        INSERT INTO deployments (id, name, source_type, source_url, status, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)`,
		d.ID, d.Name, d.SourceType, d.SourceURL, d.Status, d.CreatedAt, d.UpdatedAt,
	)
	return err
}

func (db *DB) GetDeployment(id string) (*models.Deployment, error) {
	row := db.QueryRow(`
        SELECT id, name, source_type, source_url, status,
               image_tag, container_id, host_port, live_url, error,
               created_at, updated_at
        FROM deployments WHERE id = ?`, id)
	return scanDeployment(row)
}

func (db *DB) ListDeployments() ([]*models.Deployment, error) {
	rows, err := db.Query(`
        SELECT id, name, source_type, source_url, status,
               image_tag, container_id, host_port, live_url, error,
               created_at, updated_at
        FROM deployments ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []*models.Deployment
	for rows.Next() {
		d, err := scanDeployment(rows)
		if err != nil {
			return nil, err
		}
		deployments = append(deployments, d)
	}
	return deployments, rows.Err()
}

func (db *DB) UpdateDeploymentStatus(id string, status models.Status) error {
	_, err := db.Exec(
		`UPDATE deployments SET status = ?, updated_at = ? WHERE id = ?`,
		status, time.Now(), id,
	)
	return err
}

func (db *DB) UpdateDeploymentRunning(id, imageTag, containerID string, hostPort int, liveURL string) error {
	_, err := db.Exec(`
        UPDATE deployments
        SET status = ?, image_tag = ?, container_id = ?, host_port = ?, live_url = ?, updated_at = ?
        WHERE id = ?`,
		models.StatusRunning, imageTag, containerID, hostPort, liveURL, time.Now(), id,
	)
	return err
}

func (db *DB) UpdateDeploymentFailed(id, errMsg string) error {
	_, err := db.Exec(`
        UPDATE deployments SET status = ?, error = ?, updated_at = ? WHERE id = ?`,
		models.StatusFailed, errMsg, time.Now(), id,
	)
	return err
}

func (db *DB) AppendLog(deploymentID, stream, line string) error {
	_, err := db.Exec(`
        INSERT INTO logs (deployment_id, stream, line, created_at)
        VALUES (?, ?, ?, ?)`,
		deploymentID, stream, line, time.Now(),
	)
	return err
}

func (db *DB) GetLogs(deploymentID string) ([]*models.LogLine, error) {
	rows, err := db.Query(`
        SELECT id, deployment_id, stream, line, created_at
        FROM logs WHERE deployment_id = ? ORDER BY id ASC`,
		deploymentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.LogLine
	for rows.Next() {
		l := &models.LogLine{}
		if err := rows.Scan(&l.ID, &l.DeploymentID, &l.Stream, &l.Line, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanDeployment(s scanner) (*models.Deployment, error) {
	d := &models.Deployment{}
	var sourceURL, imageTag, containerID, liveURL, errMsg sql.NullString
	var hostPort sql.NullInt64

	err := s.Scan(
		&d.ID, &d.Name, &d.SourceType, &sourceURL, &d.Status,
		&imageTag, &containerID, &hostPort, &liveURL, &errMsg,
		&d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan deployment: %w", err)
	}

	if sourceURL.Valid {
		d.SourceURL = &sourceURL.String
	}
	if imageTag.Valid {
		d.ImageTag = &imageTag.String
	}
	if containerID.Valid {
		d.ContainerID = &containerID.String
	}
	if hostPort.Valid {
		p := int(hostPort.Int64)
		d.HostPort = &p
	}
	if liveURL.Valid {
		d.LiveURL = &liveURL.String
	}
	if errMsg.Valid {
		d.Error = &errMsg.String
	}

	return d, nil
}
