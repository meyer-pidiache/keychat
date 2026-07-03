.PHONY: all build test run lint clean js-build js-dev

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
	cd web && npm run build

js-dev:
	cd web && npm run dev

js-test:
	cd web && npm test

js-lint:
	cd web && npm run lint
