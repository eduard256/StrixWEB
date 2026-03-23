package main

import (
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var (
	githubToken string
	githubRepo  string
)

func main() {
	listen := env("LISTEN", ":8080")
	dbPath := env("DB_PATH", "./cameras.db")
	corsOrigins := env("CORS_ORIGINS", "*")
	githubToken = env("GITHUB_TOKEN", "")
	githubRepo = env("GITHUB_REPO", "eduard256/StrixCamDB")

	if err := openDB(dbPath); err != nil {
		log.Fatal(err)
	}

	// read-only endpoints: rate limit + CORS
	http.HandleFunc("/api/brands/", cors(rateLimit(apiBrands), corsOrigins))
	http.HandleFunc("/api/brands", cors(rateLimit(apiBrands), corsOrigins))
	http.HandleFunc("/api/search", cors(rateLimit(apiSearch), corsOrigins))
	http.HandleFunc("/api/stats", cors(rateLimit(apiStats), corsOrigins))

	// write endpoint: rate limit + CORS + body size limit (1KB)
	http.HandleFunc("/api/contribute", cors(rateLimit(limitBody(apiContribute, 1024)), corsOrigins))

	log.Printf("listening on %s", listen)
	log.Fatal(http.ListenAndServe(listen, nil))
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
