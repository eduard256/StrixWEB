package main

import (
	"encoding/json"
	"net/http"
)

type ContributeRequest struct {
	Brand     string `json:"brand"`
	URL       string `json:"url"`
	Protocol  string `json:"protocol"`
	Port      int    `json:"port"`
	Model     string `json:"model,omitempty"`
	MACPrefix string `json:"mac_prefix,omitempty"`
	Comment   string `json:"comment,omitempty"`
}

// POST /api/contribute
func apiContribute(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ContributeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, map[string]any{"ok": false, "error": "invalid json"})
		return
	}

	// validate required fields
	if req.Brand == "" || req.URL == "" || req.Protocol == "" {
		writeJSON(w, map[string]any{"ok": false, "error": "brand, url, protocol are required"})
		return
	}

	if req.Port < 0 || req.Port > 65535 {
		writeJSON(w, map[string]any{"ok": false, "error": "port must be 0-65535"})
		return
	}

	// validate field lengths
	if len(req.Brand) > 200 || len(req.URL) > 500 || len(req.Protocol) > 20 ||
		len(req.Model) > 200 || len(req.MACPrefix) > 20 || len(req.Comment) > 1000 {
		writeJSON(w, map[string]any{"ok": false, "error": "field too long"})
		return
	}

	issueURL, err := createIssue(githubToken, githubRepo, req)
	if err != nil {
		writeJSON(w, map[string]any{"ok": false, "error": err.Error()})
		return
	}

	writeJSON(w, map[string]any{"ok": true, "issue_url": issueURL})
}
