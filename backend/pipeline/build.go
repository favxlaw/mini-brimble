package pipeline

import (
	"bufio"
	"fmt"
	"os/exec"

	"github.com/favxlaw/mini-brimble/db"
)

func ensureBuildKit() error {
	out, _ := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", "buildkit").Output()
	if string(out) == "true\n" {
		return nil
	}

	cmd := exec.Command("docker", "run",
		"--rm", "--privileged", "-d",
		"--name", "buildkit",
		"moby/buildkit",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		// Already exists but not running — ignore
		_ = out
	}

	return nil
}

func Build(store *db.DB, broadcaster *LogBroadcaster, deploymentID, sourceDir string) (string, error) {
	imageTag := fmt.Sprintf("brimble-%s", deploymentID)

	if err := ensureBuildKit(); err != nil {
		return "", fmt.Errorf("ensure buildkit: %w", err)
	}

	cmd := exec.Command("railpack", "build", sourceDir, "--name", imageTag)

	cmd.Env = append(cmd.Environ(),
		"BUILDKIT_HOST=docker-container://buildkit",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("stdout pipe: %w", err)
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("start railpack: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		store.AppendLog(deploymentID, "build", line)
		broadcaster.Publish(deploymentID, line)
	}

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("railpack build failed: %w", err)
	}

	return imageTag, nil
}
