# KeyChat

Cryptography education meets Nostr private messaging.

KeyChat is a dual-face web application — an educational site about asymmetric cryptography compatible with Facebook's FreeBasics zero-rated platform, and a fully-compatible Nostr relay with encrypted direct messaging (NIP-17).

## How it works

KeyChat serves two faces from a single Go server:

- **FreeBasics face**: Pure HTML educational content (no JavaScript required, zero-rated). Users paste pre-encrypted blobs via HTML forms — the server never sees plaintext (BYOE model).
- **Direct-access face**: Full Progressive Web App with client-side Web Crypto for key generation, encryption/decryption, and clipboard-based blob transport.
- **Nostr client compatibility**: Any standard Nostr client (Damus, Amethyst, noscl) can connect via WebSocket and use the relay.

## Quick start

### Prerequisites

- Go 1.26+
- Node.js 22+
- pnpm (install via `npm install -g pnpm` or enable `corepack`)

### Build and run

```bash
# Clone the repo
git clone https://github.com/meyer-pidiache/keychat.git
cd keychat

# Install JS dependencies
make js-dev

# Build the JS frontend
make js-build

# Build the Go relay
make build

# Run
make run
```

Open http://localhost:8080 — the server auto-detects FreeBasics via the `X-IORG-FBS` header and serves the appropriate face.

### Docker (multi-environment)

KeyChat uses multiple Compose files for different environments. Docker is all you need — no Node.js or Go required.

```bash
# Development (hot reload with compose watch)
make docker-dev

# Production (optimized multi-stage build)
make docker-prod

# Build production image only
make docker-build
```

Or use Compose directly:

```bash
# Development — live code changes without rebuild
docker compose -f compose.yaml -f compose.dev.yaml up --watch

# Production — clean build, healthchecks, resource limits
docker compose -f compose.yaml -f compose.prod.yaml up -d
```

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                   KeyChat Relay                      │
│  Go server with:                                     │
│    • NIP-01 WebSocket relay                          │
│    • NIP-17 gift-wrapped encrypted DMs               │
│    • FreeBasics HTTP bridge (REST forms)             │
│    • SQLite storage with 24h TTL                     │
│    • Rate limiting (100 req/h/IP)                    │
│    • 5 educational pages (Go html/template)          │
├─────────────────────────────────────────────────────┤
│                   Web Client (PWA)                    │
│  TypeScript + Vite with:                             │
│    • NIP-44 ChaCha20-Poly1305 encryption             │
│    • secp256k1 ECDH key exchange                     │
│    • NIP-17 gift wrap/unwrap                         │
│    • Service Worker for offline operation            │
│    • Clipboard-based blob transport                  │
└─────────────────────────────────────────────────────┘
```

### Routes

| Path | Description |
|------|-------------|
| `GET /` | PWA (JS) or FreeBasics landing (HTML) |
| `GET /learn/:topic` | Educational pages (5 topics) |
| `GET/POST /fb/send` | FreeBasics send form |
| `GET /fb/receive` | FreeBasics receive form |
| `WS /ws` | Nostr NIP-01 WebSocket relay |
| `GET /api/events` | REST events API |
| `POST /api/events` | REST event submission |

## Tech stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.26, `net/http`, `coder/websocket` |
| Frontend | TypeScript 5.7, Vite 6 |
| Database | SQLite (`modernc.org/sqlite`, pure Go, no CGO) |
| Crypto (server) | secp256k1 (`decred/dcrd`), NIP-44 |
| Crypto (client) | `@noble/ciphers`, `@noble/curves`, `@noble/hashes` |
| Deployment | Docker, Docker Compose, Cloudflare Tunnel |

## License

PolyForm Noncommercial License 1.0.0. See [LICENSE](LICENSE).

You may use, modify, and distribute this software for noncommercial purposes. Commercial use requires explicit permission from the author.

Copyright Meyer Pidiache (https://github.com/meyer-pidiache)
