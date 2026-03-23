package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

// GET /api/brands
// GET /api/brands/{brand_id}
// GET /api/brands/{brand_id}/{model}
func apiBrands(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// parse path: /api/brands, /api/brands/dahua, /api/brands/dahua/IPC-HDW1220S
	path := strings.TrimPrefix(r.URL.Path, "/api/brands")
	path = strings.TrimPrefix(path, "/")
	parts := strings.SplitN(path, "/", 2)

	switch {
	case path == "":
		// all brands
		brands, err := queryBrands()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, brands)

	case len(parts) == 1:
		// models for brand
		brandID := parts[0]
		models, err := queryModels(brandID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if models == nil {
			http.Error(w, "brand not found", http.StatusNotFound)
			return
		}
		writeJSON(w, models)

	case len(parts) == 2:
		// streams for model
		brandID := parts[0]
		model := parts[1]
		streams, err := queryStreams(brandID, model)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if streams == nil {
			http.Error(w, "model not found", http.StatusNotFound)
			return
		}
		writeJSON(w, streams)
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
