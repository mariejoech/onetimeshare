# OnetimeShare

OnetimeShare is a small Go web app for sharing one-time secrets (burn-after-reading). It uses Fiber for routing, Django-style templates for views, and SQLite for storage.

**Tech stack:** Go, Fiber, SQLite, HTML templates

## Features

- Create short-lived secrets that can be viewed once and then burned
- Optional password protection for secrets
- Simple HTML views for creating and viewing secrets

## Repo structure

- [main.go](main.go) — application entry
- [database/schema.sql](database/schema.sql) — SQL schema
- [database/repository](database/repository) — DB connection and models
- [src](src) — route setup and app wiring
- [views](views) — HTML templates

## Requirements

- Go 1.20+ (or compatible)
- git

This project uses the following Go modules (automatically managed via `go.mod`):

- `github.com/gofiber/fiber/v2`
- `github.com/gofiber/template/django/v3`
- `github.com/joho/godotenv`
- `modernc.org/sqlite`

## Environment

Create a `.env` file in the project root with the following variables:

- `DB_FILE` — path to the SQLite database file (e.g. `data/onetimeshare.db`)

Example `.env`:

```
DB_FILE=data/onetimeshare.db
```

The app will read `.env` at startup using `godotenv`.

## Initialize database

You can create the SQLite file and apply the schema manually. From the repo root:

```bash
mkdir -p data
sqlite3 data/onetimeshare.db < database/schema.sql
```

If you don't run the SQL manually, the app's DB code will ensure the `secrets` table exists when it first connects.

## Run locally

Start the app with:

```bash
go run main.go
```

By default the server listens on port `3002` (see [main.go](main.go)). Open `http://localhost:3002` to use the UI.

## Build

```bash
go build -o onetimeshare .
./onetimeshare
```

## Tests

There are no automated tests included in this repo. Add tests under a `*_test.go` convention and run `go test ./...`.

## Contributing

Contributions are welcome. Open an issue or submit a pull request. Keep changes small and focused.

## Notes

- The app uses SQLite for simplicity. For production use, consider migrating to PostgreSQL and updating the DB layer accordingly.
- Templates are in the `views/` folder; feel free to style or internationalize them.

---
