package main

import (
	"database/sql"
	"fmt"
)

var db *sql.DB

func openDB(path string) error {
	var err error
	db, err = sql.Open("sqlite3", path+"?mode=ro&_journal_mode=OFF")
	if err != nil {
		return fmt.Errorf("db: open: %w", err)
	}

	db.Exec("PRAGMA cache_size = -512")

	return db.Ping()
}

type Brand struct {
	BrandID string `json:"brand_id"`
	Brand   string `json:"brand"`
}

type Model struct {
	Model string `json:"model"`
}

type Stream struct {
	URL      string `json:"url"`
	Protocol string `json:"protocol"`
	Port     int    `json:"port"`
	Notes    string `json:"notes,omitempty"`
}

type Stats struct {
	Brands  int `json:"brands"`
	Streams int `json:"streams"`
	Models  int `json:"models"`
}

func queryBrands() ([]Brand, error) {
	rows, err := db.Query("SELECT brand_id, brand FROM brands ORDER BY brand")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var brands []Brand
	for rows.Next() {
		var b Brand
		if err := rows.Scan(&b.BrandID, &b.Brand); err != nil {
			return nil, err
		}
		brands = append(brands, b)
	}
	return brands, rows.Err()
}

func queryModels(brandID string) ([]Model, error) {
	rows, err := db.Query(`
		SELECT DISTINCT sm.model
		FROM stream_models sm
		JOIN streams s ON s.id = sm.stream_id
		WHERE s.brand_id = ? AND sm.model != '*'
		ORDER BY sm.model`, brandID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []Model
	for rows.Next() {
		var m Model
		if err := rows.Scan(&m.Model); err != nil {
			return nil, err
		}
		models = append(models, m)
	}
	return models, rows.Err()
}

func queryStreams(brandID, model string) ([]Stream, error) {
	rows, err := db.Query(`
		SELECT DISTINCT s.url, s.protocol, s.port, COALESCE(s.notes, '')
		FROM streams s
		JOIN stream_models sm ON sm.stream_id = s.id
		WHERE s.brand_id = ? AND (sm.model = ? OR sm.model = '*')
		ORDER BY s.protocol, s.port, s.url`, brandID, model)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var streams []Stream
	for rows.Next() {
		var st Stream
		if err := rows.Scan(&st.URL, &st.Protocol, &st.Port, &st.Notes); err != nil {
			return nil, err
		}
		streams = append(streams, st)
	}
	return streams, rows.Err()
}

type SearchResult struct {
	BrandID string `json:"brand_id"`
	Brand   string `json:"brand"`
	Model   string `json:"model"`
}

func querySearch(q string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := db.Query(`
		SELECT DISTINCT s.brand_id, b.brand, sm.model
		FROM stream_models sm
		JOIN streams s ON s.id = sm.stream_id
		JOIN brands b ON b.brand_id = s.brand_id
		WHERE sm.model LIKE ? AND sm.model != '*'
		ORDER BY b.brand, sm.model
		LIMIT ?`, "%"+q+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.BrandID, &r.Brand, &r.Model); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func queryStats() (Stats, error) {
	var s Stats
	err := db.QueryRow("SELECT COALESCE(value, '0') FROM meta WHERE key = 'brands'").Scan(&s.Brands)
	if err != nil {
		return s, err
	}
	err = db.QueryRow("SELECT COALESCE(value, '0') FROM meta WHERE key = 'streams'").Scan(&s.Streams)
	if err != nil {
		return s, err
	}
	err = db.QueryRow("SELECT COUNT(DISTINCT model) FROM stream_models WHERE model != '*'").Scan(&s.Models)
	return s, err
}
