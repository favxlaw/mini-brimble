package pipeline

import (
	"fmt"
	"os/exec"
	"strings"
)

func StartContainer(imageTag, containerPort, dockerNetwork string) (containerID string, containerIP string, hostPort int, err error) {
	out, err := exec.Command("docker", "run",
		"-d",
		"--network", dockerNetwork,
		"-p", fmt.Sprintf("::%s", containerPort),
		"--env", fmt.Sprintf("PORT=%s", containerPort),
		imageTag,
	).Output()
	if err != nil {
		return "", "", 0, fmt.Errorf("docker run: %w", err)
	}

	containerID = strings.TrimSpace(string(out))

	// Use json format and parse it to avoid template issues with hyphenated network names
	ipOut, err := exec.Command("docker", "inspect",
		"--format", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}",
		containerID,
	).Output()
	if err != nil {
		return "", "", 0, fmt.Errorf("docker inspect ip: %w", err)
	}
	containerIP = strings.TrimSpace(string(ipOut))

	portOut, err := exec.Command("docker", "port", containerID, containerPort).Output()
	if err != nil {
		return "", "", 0, fmt.Errorf("docker port: %w", err)
	}

	portStr := strings.TrimSpace(string(portOut))
	lastColon := strings.LastIndex(portStr, ":")
	if lastColon == -1 {
		return "", "", 0, fmt.Errorf("unexpected port output: %s", portStr)
	}

	fmt.Sscanf(portStr[lastColon+1:], "%d", &hostPort)

	return containerID, containerIP, hostPort, nil
}

func StopContainer(containerID string) error {
	if err := exec.Command("docker", "stop", containerID).Run(); err != nil {
		return fmt.Errorf("docker stop: %w", err)
	}

	return exec.Command("docker", "rm", containerID).Run()
}
