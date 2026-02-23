# Monolith

> A local-first personal documentation indexer. Search across your local files, Google Drive, Google Docs, and Notion from a single interface — read-only, fast, and entirely on your machine.

---

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Configuration](#configuration)
- [Usage](#usage)
  - [Running the App](#running-the-app)
  - [CLI Search](#cli-search)
  - [Web UI](#web-ui)
- [Source Setup](#source-setup)
  - [Local Filesystem](#local-filesystem)
  - [Google Drive & Docs](#google-drive--docs)
  - [Notion](#notion)
- [Development](#development)
  - [Project Structure](#project-structure)
  - [Running Tests](#running-tests)
  - [Building from Source](#building-from-source)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

Monolith is a self-hosted, read-only document search tool written in Go. It ingests documents from multiple sources, indexes them locally using an embedded full-text search engine, and exposes a simple web UI and REST API for querying. No data ever leaves your machine.

**Core principles:**
- **Local-first** — all indexing and search happens on-device
- **Read-only** — Monolith never writes to, modifies, or deletes source documents
- **Unified** — one search box across all your sources
- **Simple** — single binary, one config file, no external services required

---

## Features

- Full-text search with BM25 ranking and snippet highlighting
- Incremental (delta) sync — only re-indexes changed documents
- Source filtering by origin (local, gdrive, gdocs, notion)
- OS-native document opening (read-only)
- Embedded web UI served from the binary
- Per-source connectivity status and sync metrics
- Configurable sync interval and watched directories

---

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                        Monolith                          │
│                                                         │
│  ┌─────────────┐    ┌──────────────┐    ┌────────────┐  │
│  │  Ingestion  │    │   Indexing   │    │   Search   │  │
│  │   Layer     │───▶│    Layer     │───▶│    API     │  │
│  └─────────────┘    └──────────────┘    └────────────┘  │
│       │                    │                  │          │
│  ┌────┴────┐          ┌────┴────┐       ┌────┴────┐     │
│  │Sources  │          │ Bleve   │       │  HTTP   │     │
│  │- Local  │          │  FTS    │       │ /search │     │
│  │- GDrive │          ├─────────┤       └────┬────┘     │
│  │- GDocs  │          │ SQLite  │            │          │
│  │- Notion │          │Metadata │       ┌────┴────┐     │
│  └─────────┘          └─────────┘       │  Web UI │     │
│                                         └─────────┘     │
└─────────────────────────────────────────────────────────┘
```

---

## Getting Started

### Prerequisites

- Go 1.21 or later
- A Google Cloud project with the Drive and Docs APIs enabled (for Google sources)
- A Notion integration token (for Notion source)

### Installation

**From source:**

```bash
git clone https://github.com/yourname/monolith.git
cd monolith
go build -o monolith ./cmd/monolith
```

**Using `go install`:**

```bash
go install github.com/yourname/monolith/cmd/monolith@latest
```

### Configuration

Monolith reads from `~/.monolith/config.yaml` by default. Copy the example config and edit it:

```bash
cp config.example.yaml ~/.monolith/config.yaml
```

```yaml
# ~/.monolith/config.yaml

port: 7474
index_path: ~/.monolith/index
sync_interval: 15m

local:
  paths:
    - ~/Documents
    - ~/Notes

google:
  credentials_file: ~/.monolith/google_credentials.json
  token_file: ~/.monolith/google_token.json

notion:
  api_key: secret_xxxxxxxxxxxxxxxxxxxx
```

| Key | Default | Description |
|---|---|---|
| `port` | `7474` | Port for the web UI and API |
| `index_path` | `~/.monolith/index` | Directory for Bleve index and SQLite metadata |
| `sync_interval` | `15m` | How often to run background sync |
| `local.paths` | `[]` | Directories to walk recursively |
| `google.credentials_file` | — | Path to Google OAuth2 credentials JSON |
| `notion.api_key` | — | Notion internal integration secret |

---

## Usage

### Running the App

```bash
monolith
# Monolith running on http://localhost:7474
```

On first run, Monolith will perform a full sync of all configured sources. Subsequent runs use delta sync to only process changed documents.

**Flags:**

```
--config   Path to config file (default: ~/.monolith/config.yaml)
--port     Override the port
--sync     Run a one-time sync and exit
--reindex  Wipe and rebuild the index from scratch
```

### CLI Search

```bash
# Search from the terminal
monolith search "quarterly review"

# Filter by source
monolith search "standup notes" --source=notion

# Open the top result
monolith search "architecture diagram" --open
```

### Web UI

Open `http://localhost:7474` in your browser after starting Monolith.

- Type a query and press `Enter` or click **Search**
- Results show source badge, title, last modified date, and a snippet
- Click any result to open the document in its native viewer or browser
- Click **↻ Sync now** to trigger a manual sync

### REST API

| Endpoint | Method | Description |
|---|---|---|
| `/search?q=<query>` | `GET` | Full-text search, returns JSON results |
| `/search?q=<query>&source=notion` | `GET` | Filter results by source |
| `/open?path=<path>` | `GET` | Open a document via OS handler |
| `/sync` | `POST` | Trigger a manual sync |
| `/status` | `GET` | Source connectivity and sync stats |
| `/metrics` | `GET` | Ingestion and search performance metrics |

**Example response:**

```json
{
  "query": "quarterly review",
  "count": 3,
  "results": [
    {
      "id": "notion:abc123",
      "title": "Q3 Review Notes",
      "source": "notion",
      "path": "https://notion.so/...",
      "snippet": "...discussed the <mark>quarterly review</mark> agenda with the team...",
      "last_modified": "2025-01-15T10:30:00Z",
      "score": 1.842
    }
  ]
}
```

---

## Source Setup

### Local Filesystem

Add one or more directory paths under `local.paths` in your config. Monolith will recursively walk each directory and index files with the following extensions:

`.md` `.txt` `.pdf` `.html` `.rst` `.org`

No additional setup required.

### Google Drive & Docs

1. Go to [Google Cloud Console](https://console.cloud.google.com/) and create a new project.
2. Enable the **Google Drive API** and **Google Docs API**.
3. Create OAuth 2.0 credentials (Desktop App type) and download the JSON file.
4. Set `google.credentials_file` in your config to the path of that JSON file.
5. On first run, Monolith will open a browser window for you to authorize access. The token is cached locally for future runs.

Monolith requests read-only scopes only: `drive.readonly` and `documents.readonly`.

### Notion

1. Go to [notion.so/my-integrations](https://www.notion.so/my-integrations) and create a new internal integration.
2. Copy the **Internal Integration Secret**.
3. Set `notion.api_key` in your config to that secret.
4. In Notion, open each workspace or top-level page you want indexed, click **···** → **Add connections**, and add your integration.

---

## Development

### Project Structure

```
monolith/
├── cmd/monolith/         # Entry point
├── internal/
│   ├── config/           # Config loading and validation
│   ├── ingestion/        # Source orchestrator + Source interface
│   │   ├── local/        # Filesystem source
│   │   ├── gdrive/       # Google Drive + Docs source
│   │   └── notion/       # Notion source
│   ├── index/            # Bleve FTS + SQLite metadata
│   ├── search/           # Query logic and result assembly
│   ├── sync/             # Delta sync and scheduling
│   └── server/           # HTTP server, handlers, embedded UI
├── pkg/document/         # Shared Document type
├── config.example.yaml
├── go.mod
└── go.sum
```

### Running Tests

```bash
# All tests
go test ./...

# With coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# A specific package
go test ./internal/ingestion/local/...
```

### Building from Source

```bash
# Development build
go build -o monolith ./cmd/monolith

# Production build (smaller binary, no debug info)
go build -ldflags="-s -w" -o monolith ./cmd/monolith

# Cross-compile for Linux
GOOS=linux GOARCH=amd64 go build -o monolith-linux ./cmd/monolith
```

---

## Roadmap

- [x] Local filesystem ingestion
- [x] Google Drive + Docs ingestion
- [x] Notion ingestion
- [x] Bleve full-text search
- [x] Embedded web UI
- [x] Delta sync
- [ ] PDF text extraction
- [ ] Obsidian vault support
- [ ] Confluence source
- [ ] Semantic / vector search (optional embedding backend)
- [ ] Browser extension for quick search
- [ ] macOS menu bar app

---

## Contributing

Contributions are welcome. Please open an issue before submitting a pull request for significant changes.

```bash
# Fork and clone
git clone https://github.com/yourname/monolith.git

# Create a feature branch
git checkout -b feat/my-new-source

# Run tests before submitting
go test ./...
go vet ./...
```

Please follow standard Go conventions (`gofmt`, idiomatic error handling, table-driven tests).

---

## License

MIT License. See [LICENSE](LICENSE) for details.
