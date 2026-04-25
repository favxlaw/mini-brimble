package pipeline

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/favxlaw/mini-brimble/config"
	"github.com/favxlaw/mini-brimble/db"
	"github.com/favxlaw/mini-brimble/models"
)

func Deploy(store *db.DB, broadcaster *LogBroadcaster, cfg *config.Config, deployment *models.Deployment) {
	fail := func(err error) {
		msg := err.Error()
		store.AppendLog(deployment.ID, "system", "ERROR: "+msg)
		broadcaster.Publish(deployment.ID, "ERROR: "+msg)
		store.UpdateDeploymentFailed(deployment.ID, msg)
	}

	sourceDir, err := os.MkdirTemp("", fmt.Sprintf("brimble-deploy-%s", deployment.ID))
	if err != nil {
		fail(fmt.Errorf("create temp dir: %w", err))
		return
	}
	defer os.RemoveAll(sourceDir)

	store.AppendLog(deployment.ID, "system", "Starting deployment...")
	broadcaster.Publish(deployment.ID, "Starting deployment...")

	if deployment.SourceType == models.SourceTypeGit {
		store.AppendLog(deployment.ID, "system", fmt.Sprintf("Cloning %s...", *deployment.SourceURL))
		broadcaster.Publish(deployment.ID, fmt.Sprintf("Cloning %s...", *deployment.SourceURL))

		out, err := exec.Command("git", "clone", *deployment.SourceURL, sourceDir).CombinedOutput()
		if err != nil {
			fail(fmt.Errorf("git clone: %s", string(out)))
			return
		}

		store.AppendLog(deployment.ID, "system", "Clone complete.")
		broadcaster.Publish(deployment.ID, "Clone complete.")
	}

	if err := store.UpdateDeploymentStatus(deployment.ID, models.StatusBuilding); err != nil {
		fail(fmt.Errorf("update status building: %w", err))
		return
	}

	store.AppendLog(deployment.ID, "system", "Building image with Railpack...")
	broadcaster.Publish(deployment.ID, "Building image with Railpack...")

	imageTag, err := Build(store, broadcaster, deployment.ID, sourceDir)
	if err != nil {
		fail(fmt.Errorf("build: %w", err))
		return
	}

	store.AppendLog(deployment.ID, "system", fmt.Sprintf("Image built: %s", imageTag))
	broadcaster.Publish(deployment.ID, fmt.Sprintf("Image built: %s", imageTag))

	if err := store.UpdateDeploymentStatus(deployment.ID, models.StatusDeploying); err != nil {
		fail(fmt.Errorf("update status deploying: %w", err))
		return
	}

	store.AppendLog(deployment.ID, "system", "Starting container...")
	broadcaster.Publish(deployment.ID, "Starting container...")

	containerID, hostPort, err := StartContainer(imageTag, cfg.ContainerPort)
	if err != nil {
		fail(fmt.Errorf("start container: %w", err))
		return
	}

	store.AppendLog(deployment.ID, "system", fmt.Sprintf("Container started on port %d", hostPort))
	broadcaster.Publish(deployment.ID, fmt.Sprintf("Container started on port %d", hostPort))

	store.AppendLog(deployment.ID, "system", "Configuring routing...")
	broadcaster.Publish(deployment.ID, "Configuring routing...")

	if err := AddRoute(cfg.CaddyAdminURL, deployment.ID, hostPort); err != nil {
		fail(fmt.Errorf("add route: %w", err))
		return
	}

	liveURL := cfg.LiveURL(deployment.ID)

	if err := store.UpdateDeploymentRunning(deployment.ID, imageTag, containerID, hostPort, liveURL); err != nil {
		fail(fmt.Errorf("update deployment running: %w", err))
		return
	}

	store.AppendLog(deployment.ID, "system", fmt.Sprintf("Deployment live at %s", liveURL))
	broadcaster.Publish(deployment.ID, fmt.Sprintf("Deployment live at %s", liveURL))
}
