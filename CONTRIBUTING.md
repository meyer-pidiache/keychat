# Contributing

Thanks for your interest in KeyChat.

## Prerequisites

- **Go** 1.26 or later. Install via [golang.org/dl](https://golang.org/dl/).
- **Node.js** 22 or later. Install via [nodejs.org](https://nodejs.org/).
- **npm** (ships with Node.js).

## Setup

```bash
# Clone your fork
git clone https://github.com/<your-username>/keychat.git
cd keychat

# Install JS dependencies
make js-dev

# Build the frontend
make js-build

# Build the Go relay
make build

# Run the tests
make test
```

## Git hooks

This project ships custom git hooks under `.githooks/`:

| Hook | What it does |
|------|-------------|
| `pre-commit` | `go fmt ./...` + `go vet ./...` |
| `commit-msg` | Validates the commit message follows Conventional Commits |
| `pre-push` | `go test ./... -count=1` |

Enable them with:

```bash
git config core.hooksPath .githooks
```

## Running tests

```bash
# Go tests
go test ./... -v -count=1

# JS lint
make js-lint
```

## Commit conventions

This project uses [Conventional Commits](https://www.conventionalcommits.org/).

```
feat: add NIP-42 AUTH support
fix: handle missing created_at in filter
docs: update README with WebSocket examples
chore: bump coder/websocket to 1.8.16
test: add property tests for event validation
refactor: extract subscription manager
style: format Go imports
```

Prefixes: `feat:`, `fix:`, `docs:`, `test:`, `chore:`, `refactor:`, `style:`.

Keep commits focused on a single change.

## Pull request process

1. Fork the repository on GitHub.
2. Create a feature branch from `main`.
3. Make your changes and commit them.
4. Run `go test ./... -count=1` and `go vet ./...` — both must pass.
5. Push your branch and open a PR against `main`.
6. Describe what the PR does and why.

## Pull request checklist

- [ ] `go test ./... -count=1` passes
- [ ] `go vet ./...` passes
- [ ] `go fmt ./...` passes
- [ ] Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/)
- [ ] Changes are scoped to a single logical unit

## What to work on

Check the issue tracker for open bugs and feature requests. If you find something you'd like to work on, comment on the issue to let others know you're taking it.
