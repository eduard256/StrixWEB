package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func createIssue(token, repo string, c ContributeRequest) (string, error) {
	title := fmt.Sprintf("[Contribution] %s - %s %s", c.Brand, c.Protocol, c.URL)

	body := "```yaml\n"
	body += fmt.Sprintf("brand: %s\n", c.Brand)
	if c.Model != "" {
		body += fmt.Sprintf("model: %s\n", c.Model)
	}
	body += fmt.Sprintf("url: %s\n", c.URL)
	body += fmt.Sprintf("protocol: %s\n", c.Protocol)
	body += fmt.Sprintf("port: %d\n", c.Port)
	if c.MACPrefix != "" {
		body += fmt.Sprintf("mac_prefix: %s\n", c.MACPrefix)
	}
	if c.Comment != "" {
		body += fmt.Sprintf("comment: %s\n", c.Comment)
	}
	body += "```"

	payload := map[string]any{
		"title":  title,
		"body":   body,
		"labels": []string{"contribution"},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/issues", repo)
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("github: %d: %s", resp.StatusCode, b)
	}

	var result struct {
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.HTMLURL, nil
}
