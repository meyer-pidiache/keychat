## Project context

KeyChat is a dual-face Go + TypeScript application: a Nostr relay (Go) with encrypted DMs and an educational cryptography site compatible with Facebook's FreeBasics (no JS, no SVG). The PWA client lives in `/web`.

## Available skills

### Web (autoskills — `web/.agents/skills/`)

| Skill | Description |
|-------|-------------|
| `frontend-design` | Distinctive, production-grade frontend UI/UX design |
| `accessibility` | WCAG 2.2 compliance, screen reader, keyboard navigation |
| `seo` | Meta tags, structured data, sitemaps, search visibility |
| `vite` | Build tool config, plugins, SSR, Rolldown migration |
| `typescript-advanced-types` | Strict TypeScript patterns, generics, type safety |
| `nodejs-backend-patterns` | Express/Fastify middleware, error handling, API design |
| `nodejs-best-practices` | Node.js architecture decisions, async patterns, security |

### Go (TODO — future)

Once installed, Go skills will live in `web/.agents/skills/go-*` or at the repo root as `internal/.agents/skills/`. Common candidates:
- `go-project-layout` — Go project structure, naming, conventions
- `go-net-http` — idiomatic `net/http` handlers, middleware, testing
- `go-sqlite` — SQLite patterns with `modernc.org/sqlite`
- `go-crypto` — secp256k1, NIP-44, ECDH implementations

## How skills load

Skills in `web/.agents/skills/` are loaded automatically by OpenCode when the task matches the skill's `description` frontmatter. Each `SKILL.md` follows the [Agent Skills spec](https://agentskills.io): frontmatter metadata + progressive disclosure instructions.

To add a new skill, create a directory under `web/.agents/skills/<name>/` with a `SKILL.md` containing frontmatter (`name`, `description`) and markdown body. Optionally add `references/` and `scripts/` subdirectories.

## Build & verify

```bash
# Go
go fmt ./...
go vet ./...
go build ./...
go test ./... -count=1

# Web (PWA)
cd web && pnpm install && pnpm run build
```

## Conventions

- Conventional commits: `feat(scope):`, `fix(scope):`, `refactor(scope):`, `chore(scope):`
- PRs link an approved issue
- FreeBasics: no JS, no SVG. Serve pure HTML when `X-IORG-FBS` header detected.
- TypeScript: no `@ts-ignore`/`@ts-nocheck`. Use proper types from `@noble/*` libraries.
