package pipeline

import (
	"fmt"
	"os/exec"
	"strings"
)

func StartContainer(imageTag, containerPort string) (containerID string, hostPort int, err error) {
	out, err := exec.Command("docker", "run",
		"-d",
		"-p", fmt.Sprintf("0::%s", containerPort),
		"--env", fmt.Sprintf("PORT=%s", containerPort),
		imageTag,
	).Output()
	if err != nil {
		return "", 0, fmt.Errorf("docker run: %w", err)
	}

	containerID = strings.TrimSpace(string(out))

	portOut, err := exec.Command("docker", "port", containerID, containerPort).Output()
	if err != nil {
		return "", 0, fmt.Errorf("docker port: %w", err)
	}

	portStr := strings.TrimSpace(string(portOut))
	parts := strings.Split(portStr, ":")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("unexpected port output: %s", portStr)
	}

	fmt.Sscanf(parts[1], "%d", &hostPort)

	return containerID, hostPort, nil
}

func StopContainer(containerID string) error {
	if err := exec.Command("docker", "stop", containerID).Run(); err != nil {
		return fmt.Errorf("docker stop: %w", err)
	}

	return exec.Command("docker", "rm", containerID).Run()
}
