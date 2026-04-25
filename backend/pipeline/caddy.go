package pipeline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func AddRoute(caddyAdminURL, deploymentID string, hostPort int) error {
	route := map[string]any{
		"match": []map[string]any{
			{"path": []string{fmt.Sprintf("/deploys/%s/*", deploymentID)}},
		},
		"handle": []map[string]any{
			{
				"handler": "reverse_proxy",
				"upstreams": []map[string]any{
					{"dial": fmt.Sprintf("host.docker.internal:%d", hostPort)},
				},
				"rewrite": map[string]any{
					"strip_path_prefix": fmt.Sprintf("/deploys/%s", deploymentID),
				},
			},
		},
	}

	body, err := json.Marshal(route)
	if err != nil {
		return fmt.Errorf("marshal route: %w", err)
	}

	url := fmt.Sprintf("%s/config/apps/http/servers/srv0/routes", caddyAdminURL)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("caddy post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("caddy returned %d", resp.StatusCode)
	}

	return nil
}

func RemoveRoute(caddyAdminURL, deploymentID string) error {
	url := fmt.Sprintf("%s/config/apps/http/servers/srv0/routes/%s", caddyAdminURL, deploymentID)

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("build delete request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("caddy delete: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
