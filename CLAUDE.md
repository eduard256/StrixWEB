# StrixWEB

Website and API backend for Strix project.

## Structure

```
api/          -- Go backend, self-contained module
web/          -- Frontend (TBD)
```

## API

Single Go binary. Reads `cameras.db` (SQLite) from StrixCamDB releases. Creates GitHub issues for user contributions.

### Endpoints

```
GET  /api/brands                   -- all brands
GET  /api/brands/{brand_id}        -- all models of a brand
GET  /api/brands/{brand_id}/{model} -- all streams for a model
GET  /api/search?q=DS-2CD&limit=50 -- search models by name
GET  /api/stats                    -- brands/streams/models count
POST /api/contribute               -- submit camera data (creates GitHub issue)
```

### Contribute request format

```json
{
  "brand": "Dahua",
  "url": "/live",
  "protocol": "rtsp",
  "port": 554,
  "model": "IPC-HDW1220S",
  "mac_prefix": "3C:EF:8C",
  "comment": "Works on firmware v2.800"
}
```

Required: `brand`, `url`, `protocol`, `port`. Optional: `model`, `mac_prefix`, `comment`.

Creates issue in StrixCamDB with label `contribution`.

### Security

- Rate limit: 60 req/min GET, 5 req/min POST per IP
- Body size limit: 1KB for POST
- Field length limits: brand/model 200, url 500, comment 1000
- CORS configurable via `CORS_ORIGINS` env
- Real IP from `X-Forwarded-For` (Traefik)
- SQLite in read-only mode, prepared statements

### Environment variables

```
LISTEN=:8080                      -- listen address
DB_PATH=./cameras.db              -- path to SQLite database
GITHUB_TOKEN=ghp_xxx              -- fine-grained token (issues only)
GITHUB_REPO=eduard256/StrixCamDB  -- target repo for issues
CORS_ORIGINS=*                    -- comma-separated allowed origins
```

## Docker Hub

Image: `eduard256/strixweb-api`

### Build and push

```bash
cd api
docker build -t eduard256/strixweb-api:latest .
docker push eduard256/strixweb-api:latest
```

Dockerfile downloads `cameras.db` from StrixCamDB GitHub Releases at build time. Updating the database = rebuild the image.

### Run on VPS

```bash
docker pull eduard256/strixweb-api:latest
docker run -d --name strixweb-api \
  -e GITHUB_TOKEN=ghp_xxx \
  -e CORS_ORIGINS=https://getstrix.com \
  -p 8080:8080 \
  eduard256/strixweb-api:latest
```

## Dependencies

- `github.com/mattn/go-sqlite3` -- SQLite driver (CGo)
- No frameworks, no routers. Standard `net/http`.

## Code style

Follow AlexxIT code style. No frameworks, no DI, no abstractions. Minimum code, maximum function.
