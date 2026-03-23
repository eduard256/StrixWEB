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
	githubToken = env("GITHUB_TOKEN", "")
	githubRepo = env("GITHUB_REPO", "eduard256/StrixCamDB")

	if err := openDB(dbPath); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/api/brands/", apiBrands)
	http.HandleFunc("/api/brands", apiBrands)
	http.HandleFunc("/api/search", apiSearch)
	http.HandleFunc("/api/stats", apiStats)
	http.HandleFunc("/api/contribute", apiContribute)

	log.Printf("listening on %s", listen)
	log.Fatal(http.ListenAndServe(listen, nil))
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
