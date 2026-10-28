# SonLotus — Makefile (Phases 1–4).
# PATH requirement: go and tailwindcss must be on PATH. In this VM
# they live at ~/local/go/bin/go and ~/local/bin/tailwindcss, so run
# make with: PATH=$HOME/local/go/bin:$HOME/local/bin:$PATH make <target>
# The Dockerfile does not depend on those host binaries: it downloads
# the pinned Tailwind CLI in the build stage.
GO ?= go
TAILWINDCSS ?= tailwindcss

.PHONY: dev build test vet fmt css seed

dev:
	$(GO) run ./cmd/server

build: css
	mkdir -p bin
	$(GO) build -o bin/server ./cmd/server
	$(GO) build -o bin/migrator ./cmd/migrator

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

fmt:
	$(GO) fmt ./...

# Build the Tailwind stylesheet from the committed input file.
# The generated web/static/css/app.css is a build artifact: never hand-edit it.
css:
	$(TAILWINDCSS) -c tailwind.config.js -i web/static/css/input.css -o web/static/css/app.css --minify

# make seed runs migrator --seed-only (shared seed function, PLAN §3).
seed:
	$(GO) run ./cmd/migrator --seed-only
