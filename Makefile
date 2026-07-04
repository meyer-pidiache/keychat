.PHONY: all build test lint run clean js-build js-dev js-lint docker-build docker-run

.PHONY: all build test lint run clean js-build js-dev js-lint docker-dev docker-prod docker-build docker-run

all: build

# Go relay
build:
	go build -o cmd/keychat/keychat ./cmd/keychat/...

test:
	go test ./... -v -count=1

lint:
	go vet ./...

run:
	go run ./cmd/keychat/...

clean:
	rm -f cmd/keychat/keychat
	rm -rf web/dist/

# JS client
js-build:
	cd web && pnpm run build

js-dev:
	cd web && pnpm run dev

js-lint:
	cd web && pnpm run lint

# Docker
docker-dev:
	docker compose -f compose.yaml -f compose.dev.yaml up --watch

docker-prod:
	docker compose -f compose.yaml -f compose.prod.yaml up -d

docker-build:
	docker compose -f compose.yaml -f compose.prod.yaml build

docker-run:
	docker compose -f compose.yaml -f compose.prod.yaml up -d
