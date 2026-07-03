# KeyChat: Educational Crypto Site + Nostr Messaging Relay

## TL;DR

> **Quick Summary**: Build a dual-face web application — a FreeBasics-compatible educational site about asymmetric cryptography that also functions as a fully-compatible Nostr relay with encrypted messaging (NIP-17). Users access educational content zero-rated via FreeBasics, and can use the same relay for private messaging via direct browser access (with full JS) or via Nostr clients (Damus, Amethyst, etc.).
>
> **Deliverables**:
> - Custom Nostr relay in Go (NIP-01, NIP-17 compliant) with SQLite storage + TTL
> - 5 educational HTML pages about crypto + Nostr (FreeBasics-compatible, no JS)
> - PWA offline app with Web Crypto + @noble/ciphers for encrypted messaging (keygen, encrypt, decrypt, clipboard)
> - HTTP REST bridge for FreeBasics form submissions (BYOE — paste encrypted blobs)
> - PWA Service Worker for full offline operation + installability
> - FreeBasics proxy detection (X-IORG-FBS header) → bifurcates between HTML and JS versions
>
> **Estimated Effort**: Large
> **Parallel Execution**: YES — 5 waves
> **Critical Path**: Scaffolding → Nostr Core → NIP-17 → FreeBasics Bridge → Integration

---

## Context

### Original Request
User wanted to create a FreeBasics app for downloading content without data charges. After research, FreeBasics does not support file transfer, JS, or open proxies — it's a curated zero-rating platform for lightweight informational sites.

### Pivot
Pivoted to an **educational site about asymmetric cryptography** that disguises a **Nostr-based private messaging system**. The site serves dual purpose:
1. **FreeBasics face**: Pure HTML educational content (zero-rated, no JS needed)
2. **Direct-access face**: Full messaging app with client-side crypto (JS, Web Crypto API)
3. **Same Nostr relay backend**: Serves both faces and is compatible with any Nostr client

### Interview Summary
**Key Decisions**:
- **NIP-17**: Current Nostr encrypted DM standard (not deprecated NIP-04)
- **Full Nostr compatibility**: Any Nostr client can connect to the relay
- **BYOE on FreeBasics**: Users paste pre-encrypted blobs via HTML forms
- **Project name**: KeyChat
- **Tech stack**: Go (relay/server), Vanilla JS + @noble/ciphers (client), SQLite (storage)

**Research Findings**:
- FreeBasics strips JS, rewrites URLs, uses dual-certificate HTTPS (MITM proxy), limits images to <200KB, blocks file transfer
- FreeBasics detection via `X-IORG-FBS` HTTP header
- NIP-04 is deprecated; NIP-17 (NIP-44 + NIP-59 gift wraps) is the current standard
- Nostr-native protocol is WebSocket; FreeBasics needs an HTTP REST bridge
- Libraries exist: `khatru` (Go relay framework), `vertex-lab/nostr-sqlite` (storage)

### Metis Review
**Identified Gaps** (addressed):
- **NIP-04 deprecated**: Resolved → using NIP-17
- **WebSocket in FreeBasics**: Resolved → REST HTTP bridge for form submissions
- **BYOE vs server-side**: Resolved → BYOE (users bring their own encryption tool)
- **Nostr compatibility scope**: Resolved → full NIP-01 compliance, any client can connect
- **Project naming**: Resolved → KeyChat
- **Rate limiting**: Added → 100 events/hour/IP
- **TTL policy**: Added → 24 hours for events
- **Content size limits**: Added → 64KB max per event content
- **Edge cases**: Added tests for XSS, spam, colisión de pubkeys, created_at extremos

---

## Work Objectives

### Core Objective
Build a dual-face Nostr relay + educational site that teaches asymmetric cryptography through interactive content while providing functional NIP-17 encrypted messaging.

### Concrete Deliverables
- Custom Nostr relay: Go binary, NIP-01 + NIP-17 compliant, WebSocket + REST bridge
- Educational pages: 5 HTML pages rendered server-side via Go templates
- PWA offline app: Direct-access Progressive Web App with key generation, encryption, messaging (clipboard-based transport)
- PWA Service Worker: Full offline operation + installable manifest
- FreeBasics face: Full navigation, forms, and content without JavaScript (blob transport via copy/paste)
- SQLite storage with 24h TTL cleanup

### Definition of Done
- [ ] Nostr relay passes NIP-01 compliance tests (valid events, filters, EOSE, OK messages)
- [ ] NIP-17 messages can be sent and received between two PWA clients
- [ ] NIP-17 messages can be sent and received between PWA client and Nostr client (Damus/Amethyst)
- [ ] FreeBasics detection works → serves JS-free HTML when `X-IORG-FBS` is present
- [ ] Educational pages render correctly without JavaScript
- [ ] PWA offline app can generate keys, encrypt/decrypt NIP-17 messages, and copy blobs to clipboard
- [ ] PWA installs as standalone app via manifest + Service Worker
- [ ] Rate limiting blocks >100 events/hour/IP
- [ ] Events older than 24h are purged
- [ ] All 4 verification agents pass (F1-F4)

### Must Have
- Working Nostr relay (NIP-01 WebSocket with REQ, EVENT, CLOSE, EOSE, OK)
- NIP-17 message support (kind 1059 gift wraps, kind 14 rumors, kind 13 seals)
- NIP-44 encryption/decryption (ChaCha20-Poly1305 + HMAC-SHA256, secp256k1 ECDH)
- FreeBasics detection and bifurcation
- 5 educational pages (asymmetric crypto, Nostr, relay architecture, key generation, private messaging)
- HTML forms for message transport on FreeBasics (paste encrypted blobs → POST / receive GET)
- PWA utility app (offline) for client-side crypto: keygen, encrypt, decrypt, clipboard
- PWA Service Worker for full offline operation
- SQLite storage with 24h TTL
- Schnorr signature validation for incoming events
- Rate limiting (100 events/hour/IP)

### Must NOT Have (Guardrails)
- **No NIP-42 (AUTH)**: No authentication in v1
- **No NIP-05 (DNS verification)**: No DNS identity verification
- **No admin panel**: No relay management UI
- **No group chats**: Only one-to-one DMs
- **No event forwarding**: No propagation to other relays
- **No WebSockets on FreeBasics path**: REST only for FB users
- **No JS dependency for FreeBasics**: Core navigation must work without JS
- **No >5 educational pages**: Fixed scope
- **No CLI tool dependency**: Mobile users can't use CLI — PWA replaces it
- **No server-side private keys**: Keys NEVER leave the PWA
- **No mandatory WebSocket**: PWA can use HTTP REST for transport
- **No SVG images**: Not supported by FreeBasics proxy
- **No images >200KB**: Will be truncated by FreeBasics
- **No videos or audio**: Not supported

---

## Verification Strategy (MANDATORY)

> **ZERO HUMAN INTERVENTION** — ALL verification is agent-executed. No exceptions.

### Test Decision
- **Infrastructure exists**: Will be set up as part of Task 1
- **Automated tests**: Tests-after (Go tests for relay package, JS tests for crypto module)
- **Framework**: Go `testing` package + `websocat`/`wscat` for WS tests
- **Agent QA**: Playwright for browser UI (PWA + FreeBasics), Bash/curl for API, websocat for WebSocket

### QA Policy
Every task MUST include agent-executed QA scenarios. Evidence saved to `.sisyphus/evidence/task-{N}-{scenario-slug}.{ext}`.

- **Go relay**: Bash + curl (REST bridge), websocat (WebSocket)
- **PWA client**: Playwright (browser with JS enabled, test offline via devtools)
- **FreeBasics simulation**: Playwright with JS disabled, custom headers
- **PWA install**: Playwright (check manifest link, Service Worker registration, offline mode)

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Foundation — start in parallel):
├── T1: Project scaffolding (Go module, deps, directories, CI)
├── T2: Nostr data types + event validation library (Go)
├── T3: Go HTML template system + CSS design system
└── T4: JS project scaffolding + Nostr client library bootstrap

Wave 2 (Nostr Relay Core — depends on T2):
├── T5: NIP-01 WebSocket implementation (REQ, EVENT, CLOSE, EOSE, OK)
├── T6: SQLite storage with NIP-01 filter support
├── T7: Schnorr signature validation (secp256k1)
└── T8: NIP-44 encryption implementation (Go)

Wave 3 (Messaging + Bridge — depends on Wave 2):
├── T9: NIP-17 gift wrap/unwrap (Go)
├── T10: HTTP REST bridge for FreeBasics forms
├── T11: FreeBasics detection + bifurcation middleware
├── T12: TTL cleanup + rate limiting
├── T13: Educational pages content (5 HTML pages)
└── T14: JS NIP-44 + NIP-17 implementation (Web Crypto + @noble/ciphers)

Wave 4 (PWA + Transport — parallel with Wave 3):
├── T15: JS PWA crypto utility (NIP-44/17 + keygen + clipboard)
├── T16: PWA UI (keygen, contacts, encrypt/decrypt, clipboard, history)
├── T17: FreeBasics face UI (educational nav + message transport forms)
└── T18: PWA manifest + Service Worker (offline install)

Wave 5 (Integration + Hardening):
├── T19: Nostr interop testing (connect with real Nostr clients)
├── T20: FreeBasics proxy simulation + tests (PWA clipboard integration)
├── T21: Edge case hardening (XSS, spam, big payloads, invalid sigs)
└── T22: Deployment configuration + FreeBasics submission prep

Wave FINAL (4 parallel reviews, then user okay):
├── F1: Plan compliance audit (oracle)
├── F2: Code quality + security review
├── F3: Real manual QA (full scenario execution)
└── F4: Scope fidelity check
```

### Agent Dispatch Summary

- **Wave 1**: 4 tasks → visual-engineering (T3), quick (T1, T4), deep (T2)
- **Wave 2**: 4 tasks → deep (T5, T6, T7), unspecified-high (T8)
- **Wave 3**: 6 tasks → deep (T9, T10, T11, T12), visual-engineering (T13), unspecified-high (T14)
- **Wave 4**: 4 tasks → deep (T15, T18), visual-engineering (T16, T17)
- **Wave 5**: 4 tasks → deep (T19, T21), unspecified-high (T20, T22)
- **FINAL**: 4 review agents in parallel

---

## TODOs

- [ ] 1. **Project Scaffolding — Go module + JS project + CI**

  **What to do**:
  - Initialize Go module at `/` (relay binary)
  - Directory structure: `cmd/keychat/` (main binary), `internal/relay/` (core relay), `internal/crypto/` (NIP-44), `internal/storage/` (SQLite), `internal/templates/` (HTML), `internal/bridge/` (REST), `web/` (PWA client)
  - Initialize JS project in `web/` with package.json, esbuild/vite config
  - Add Go dependencies: `github.com/nbd-wtf/go-nostr`, `modernc.org/sqlite`, `github.com/puzpuzpuz/xsync/v3`
  - Add JS dependencies: `@noble/ciphers`, `@noble/curves`, `@noble/hashes`
  - Set up Go `testing` package test patterns
  - Create Makefile with targets: `build`, `test`, `run`, `lint`, `clean`
  - Set up `.github/workflows/ci.yml` with Go test + JS build
  - Add `.gitignore` (Go, Node, SQLite)
  - **IMPORTANT**: Study `khatru` library patterns before designing relay architecture — `khatru` is a Go relay framework that handles NIP-01 WebSocket plumbing. Use it as a foundation, not as a dependency (we'll build custom to learn, but reference its patterns).
  - **IMPORTANT**: Study `vertex-lab/nostr-sqlite` for SQLite schema patterns for Nostr events.

  **Must NOT do**:
  - Do not add NIP-42 or NIP-05 dependencies
  - Do not set up Docker yet (will be in T22)
  - Do not create admin UI or panels

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: [] (standard setup, no special skills needed)

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1
  - **Blocks**: T2, T3, T4
  - **Blocked By**: None

  **References**:
  - `https://github.com/nbd-wtf/go-nostr` — Go Nostr library for type reference
  - `https://github.com/fiatjaf/khatru` — Go relay framework patterns
  - `https://github.com/vertex-lab/nostr-sqlite` — SQLite storage patterns for Nostr

  **Acceptance Criteria**:

  **QA Scenarios**:
  ```
  Scenario: Go project compiles
    Tool: Bash
    Preconditions: Go 1.22+ installed, dependencies downloaded
    Steps:
      1. cd /repo && go build ./cmd/keychat/...
      2. Verify binary exists at cmd/keychat/keychat (or .exe on Windows)
    Expected Result: Binary compiles without errors
    Failure Indicators: go build returns non-zero exit code
    Evidence: .sisyphus/evidence/task-1-build.txt

  Scenario: JS project builds
    Tool: Bash
    Preconditions: Node 20+ installed
    Steps:
      1. cd /repo/web && npm install && npm run build
      2. Verify dist/ directory contains built files
    Expected Result: npm build completes with exit code 0
    Failure Indicators: npm build fails or dist/ is empty
    Evidence: .sisyphus/evidence/task-1-js-build.txt

  Scenario: Make targets work
    Tool: Bash
    Preconditions: All deps installed
    Steps:
      1. make test  # should run `go test ./...` (may have zero tests, that's OK)
      2. make lint  # should run go vet (may pass with no issues)
    Expected Result: make targets execute without errors
    Evidence: .sisyphus/evidence/task-1-make.txt
  ```

  **Commit**: YES
  - Message: `chore(project): scaffold Go module and JS project structure`
  - Files: `go.mod`, `Makefile`, `.gitignore`, `.github/workflows/ci.yml`, `web/package.json`, `web/vite.config.ts`

---

- [ ] 2. **Nostr Data Types + Event Validation Library (Go)**

  **What to do**:
  - Implement core Nostr data types: `Event`, `Filter`, `Subscription`, `OKResponse`, `EOSEResponse`, `ClosedResponse`
  - Implement `Event.Validate()`: checks `id` hash (SHA-256), `sig` (Schnorr), required fields, `created_at` bounds (not > now+1h, not < 2015 for sanity)
  - Implement `Filter.Match(event)`: match against `ids`, `authors`, `kinds`, `#p` (tag), `since`, `until`, `limit`
  - Implement event serialization/deserialization: `Event.ToJSON()`, `Event.FromJSON()`, `Filter.ToJSON()`, `Filter.FromJSON()`
  - Implement subscription ID management (string-based, validated alphanumeric + `-` + `_`)
  - Implement `OKResponse` builder: `NewOK(subID, eventID, true, "")` and `NewError(subID, eventID, msg)`
  - All types must support JSON marshal/unmarshal matching NIP-01 specification exactly
  - Schnorr validation: use `github.com/nbd-wtf/go-nostr`'s `nostr.Signature` or implement with `github.com/decred/dcrd/dcrec/secp256k1/v4`
  - Write unit tests for each method: `TestEventValidate`, `TestFilterMatch`, `TestSerialization`

  **Must NOT do**:
  - Do not implement WebSocket handling yet (that's T5)
  - Do not implement storage yet (that's T6)
  - Do not implement NIP-44 or NIP-17 yet
  - Keep it pure data types + validation, no I/O

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: [] (pure Go data structures, no special skills)

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1
  - **Blocks**: T5, T6, T7, T8, T9
  - **Blocked By**: T1

  **References**:
  - `https://github.com/nostr-protocol/nips/blob/master/01.md` — NIP-01 spec for event format, filters, and protocol messages
  - Study `Event` and `Filter` types in `github.com/nbd-wtf/go-nostr` for reference patterns
  - Schnorr validation: secp256k1 BIP-340

  **Acceptance Criteria**:

  **QA Scenarios**:
  ```
  Scenario: Valid event passes validation
    Tool: Bash
    Preconditions: Go types implemented
    Steps:
      1. go test ./internal/relay/... -run TestEventValidate -v
    Expected Result: Test passes (valid fixtures validate OK)
    Failure Indicators: Test fails or panics
    Evidence: .sisyphus/evidence/task-2-validate.txt

  Scenario: Invalid event rejected
    Tool: Bash
    Preconditions: Go types implemented
    Steps:
      1. go test ./internal/relay/... -run TestFilterMatch -v
    Expected Result: Test passes (known good/bad filters match correctly)
    Failure Indicators: Wrong match results
    Evidence: .sisyphus/evidence/task-2-filter.txt

  Scenario: JSON roundtrip
    Tool: Bash
    Preconditions: Go types implemented
    Steps:
      1. go test ./internal/relay/... -run TestSerialization -v
    Expected Result: Event → JSON → Event is identical
    Evidence: .sisyphus/evidence/task-2-json.txt
  ```

  **Commit**: YES
  - Message: `feat(relay): implement Nostr data types and event validation`
  - Files: `internal/relay/types.go`, `internal/relay/event.go`, `internal/relay/filter.go`, `internal/relay/messages.go`, `internal/relay/types_test.go`, `internal/relay/event_test.go`, `internal/relay/filter_test.go`

---

- [ ] 3. **Go HTML Template System + CSS Design System**

  **What to do**:
  - Set up Go `html/template` rendering in `internal/templates/`
  - Create base layout template: `base.html` with header, nav, content slot, footer
  - Create CSS design system in `web/static/css/`:
    - Mobile-first, responsive (feature phone friendly — min-width 240px)
    - No JS dependencies, no CSS frameworks (pure CSS)
    - Color scheme: high contrast (accessibility for feature phones)
    - Typography: system fonts only (no WOFF — FreeBasics blocks WOFF)
    - No SVG icons — use Unicode or HTML entities
    - Max page weight: <50KB per page (including all resources)
  - Create template helpers: `safeHTML`, `truncate`, `formatTime`
  - Implement template inheritance: pages extend `base.html`
  - Add CSS reset + base styles for forms, tables, navigation
  - All images must be PNG/GIF <200KB and use `<img>` with alt text
  - **CRITICAL**: No SVGs, no JS, no external resources, no webfonts

  **Must NOT do**:
  - Do not use any CSS framework (Tailwind, Bootstrap — adds weight)
  - Do not add JS files to the template directory
  - Do not use SVG (not supported by FreeBasics)
  - Do not use WOFF/custom fonts (blocked by FreeBasics)

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [] — pure HTML/CSS design, feature-phone UX

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1
  - **Blocks**: T13, T17
  - **Blocked By**: T1

  **References**:
  - FreeBasics design constraints: HTML only, no JS, no SVG, no WOFF, max 200KB images
  - Feature phone best practices: high contrast, large touch targets (44px min), readable font sizes (16px min body)

  **Acceptance Criteria**:

  **QA Scenarios**:
  ```
  Scenario: Template renders valid HTML
    Tool: Bash
    Preconditions: Go templates + CSS files exist
    Steps:
      1. go test ./internal/templates/... -run TestRender -v
      2. OR: go run ./cmd/keychat/... & (if server wired) and curl http://localhost:8080/
    Expected Result: Valid HTML5 output (no errors)
    Failure Indicators: Template panics or returns malformed HTML
    Evidence: .sisyphus/evidence/task-3-render.txt

  Scenario: No prohibited content
    Tool: Bash
    Preconditions: Template files exist
    Steps:
      1. grep -r '<svg' internal/templates/ web/static/css/ | wc -l
      2. grep -r '<script' internal/templates/ | wc -l
    Expected Result: 0 occurrences of svg or script in template/css dirs
    Failure Indicators: SVG or script tags found in templates
    Evidence: .sisyphus/evidence/task-3-no-prohibited.txt

  Scenario: Static files <200KB
    Tool: Bash
    Preconditions: Static files exist
    Steps:
      1. find web/static/ -type f -exec wc -c {} \; | awk '$1 > 204800 {print}'
    Expected Result: No files exceed 200KB
    Evidence: .sisyphus/evidence/task-3-sizes.txt
  ```

  **Commit**: YES
  - Message: `feat(site): add Go HTML templates and CSS design system`
  - Files: `internal/templates/base.html`, `internal/templates/helpers.go`, `internal/templates/helpers_test.go`, `web/static/css/reset.css`, `web/static/css/main.css`

---

- [ ] 4. **JS Project Scaffolding + Nostr Client Library Bootstrap**

  **What to do**:
  - Set up `web/` directory with Vite (or esbuild) for JS bundling
  - Create `web/src/` structure: `crypto/` (NIP-44/17), `nostr/` (WS client), `ui/` (components), `store/` (localStorage)
  - Add JS dependencies: `@noble/ciphers`, `@noble/curves`, `@noble/hashes`
  - Install `websocket-polyfill` for Node-compatible WS (dev testing)
  - Create `tsconfig.json` with strict mode
  - Create entry point: `web/src/main.ts` (empty, ready for T14-T16)
  - Create Vite config: outputs to `web/dist/`, served by Go relay in production
  - Add npm scripts: `dev`, `build`, `test`, `lint`
  - Add ESLint + Prettier config
  - Create `web/src/crypto/ecdh.ts` stub: `deriveSharedSecret(privateKey, publicKey)` using `@noble/curves/secp256k1`

  **Must NOT do**:
  - Do not use a JS framework (React, Vue, Svelte — keeps it lightweight)
  - Do not implement actual messaging yet (that's T14-T16)
  - Do not add unnecessary dependencies

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: [] — standard JS project setup

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1
  - **Blocks**: T14, T15, T16
  - **Blocked By**: T1

  **References**:
  - `@noble/curves/secp256k1` for secp256k1 operations (ECDH, Schnorr)
  - `@noble/ciphers` for ChaCha20-Poly1305 (NIP-44)
  - `@noble/hashes` for SHA-256, HMAC, HKDF (NIP-44 key derivation)
  - NIP-44 spec: `https://github.com/nostr-protocol/nips/blob/master/44.md`

  **Acceptance Criteria**:

  **QA Scenarios**:
  ```
  Scenario: JS project builds
    Tool: Bash
    Preconditions: Node modules installed
    Steps:
      1. cd /repo/web && npm run build
      2. ls dist/ | grep -c '\.js$'
    Expected Result: Build succeeds, dist/ has .js files
    Failure Indicators: Build errors or empty dist/
    Evidence: .sisyphus/evidence/task-4-js-build.txt

  Scenario: ECDH stub compiles
    Tool: Bash
    Preconditions: TypeScript config works
    Steps:
      1. cd /repo/web && npx tsc --noEmit
    Expected Result: TypeScript compiles with no errors
    Failure Indicators: tsc reports type errors
    Evidence: .sisyphus/evidence/task-4-tsc.txt
  ```

  **Commit**: YES
  - Message: `chore(js): bootstrap JS client project with deps`
  - Files: `web/package.json`, `web/tsconfig.json`, `web/vite.config.ts`, `web/src/main.ts`, `web/src/crypto/ecdh.ts`, `web/.eslintrc.json`

---

- [ ] 5. **NIP-01 WebSocket Implementation**

  **What to do**:
  - Implement WebSocket handler in `internal/relay/ws.go` using `github.com/gorilla/websocket` or `nhooyr.io/websocket`
  - Handle all NIP-01 client messages:
    - `["EVENT", <event JSON>]` → validate, store, respond `["OK", <id>, true, ""]` or `["OK", <id>, false, "<message>"]`
    - `["REQ", <subID>, <filter1>, <filter2>, ...]` → create subscription, query stored events, send back `["EVENT", <subID>, <event>]` for each match, then `["EOSE", <subID>]`
    - `["CLOSE", <subID>]` → remove subscription
  - Implement subscription manager: concurrent map of subID → filter set, with cleanup on disconnect
  - Implement concurrent client handling: goroutine per connection, channel-based writes
  - Reconnection handling: clean up subscriptions and partial data on disconnect
  - Handle unsupported message types with `["NOTICE", "unknown message type"]`
  - Write tests using `websocat` or Go `httptest.Server` + test WebSocket connections
  - Productionize: read timeouts, write timeouts, max message size (128KB), ping/pong keepalive

  **Must NOT do**:
  - Do not implement storage — storage comes in T6, wire it up as an interface
  - Do not implement AUTH (NIP-42)
  - Do not implement NIP-17 logic here (comes in T9)
  - Do not add HTTP-only features

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: [] — concurrent Go WebSocket server

  **Parallelization**:
  - **Can Run In Parallel**: YES (with T6, T7, T8)
  - **Parallel Group**: Wave 2
  - **Blocks**: T9, T10, T12
  - **Blocked By**: T2

  **References**:
  - NIP-01 spec: `https://github.com/nostr-protocol/nips/blob/master/01.md` — Client ↔ Relay protocol messages
  - `github.com/gorilla/websocket` — WebSocket library for Go
  - Study WebSocket patterns in `github.com/fiatjaf/khatru` for relay connection management

  **Acceptance Criteria**:

  **QA Scenarios**:
  ```
  Scenario: EVENT → OK response
    Tool: Bash + websocat
    Preconditions: Relay binary running on localhost:8080
    Steps:
      1. echo '["EVENT", {
           "id": "<valid event hash>",
           "pubkey": "<valid 64-char hex>",
           "created_at": 1700000000,
           "kind": 1,
           "tags": [],
           "content": "hello",
           "sig": "<valid schnorr sig>"
         }]' | websocat ws://localhost:8080
      2. Wait for response
    Expected Result: Response is ["OK", "<event-id>", true, ""]
    Failure Indicators: Response contains false, error, or no response
    Evidence: .sisyphus/evidence/task-5-event-ok.txt

  Scenario: REQ → EVENT + EOSE
    Tool: Bash + websocat
    Preconditions: At least one stored event
    Steps:
      1. echo '["REQ", "test", {"kinds": [1], "limit": 10}]' | websocat ws://localhost:8080
    Expected Result: ["EVENT", "test", {...}] followed by ["EOSE", "test"]
    Failure Indicators: No EOSE sent, or wrong subscription ID
    Evidence: .sisyphus/evidence/task-5-req-eose.txt

  Scenario: Invalid event rejected
    Tool: Bash + websocat
    Preconditions: Relay running
    Steps:
      1. echo '["EVENT", {"id": "bad", "kind": 1, ...}]' | websocat ws://localhost:8080
    Expected Result: ["OK", "<id>", false, "error: invalid event"]
    Failure Indicators: Event accepted with true
    Evidence: .sisyphus/evidence/task-5-reject.txt

  Scenario: CLOSE removes subscription
    Tool: Bash + websocat
    Preconditions: Relay running, subscription active
    Steps:
      1. Send REQ, then CLOSE for same subID
      2. No more events should arrive for that subID
    Expected Result: Subscriptions cleanly removed
    Evidence: .sisyphus/evidence/task-5-close.txt
  ```

  **Commit**: YES
  - Message: `feat(relay): implement NIP-01 WebSocket protocol`
  - Files: `internal/relay/ws.go`, `internal/relay/subscription.go`, `internal/relay/ws_test.go`, `internal/relay/subscription_test.go`

---

- [ ] 6. **SQLite Storage with NIP-01 Filter Support**

  **What to do**:
  - Implement `internal/storage/store.go` with interface: `IEventStore`
    - `SaveEvent(event) error`
    - `QueryEvents(filters []Filter) ([]Event, error)`
    - `DeleteOlderThan(t time.Time) (int64, error)`
    - `Close() error`
  - Implement SQLite backing with `modernc.org/sqlite` (pure Go, no CGO)
  - Schema:
    ```sql
    CREATE TABLE events (
      id TEXT PRIMARY KEY,
      pubkey TEXT NOT NULL,
      created_at INTEGER NOT NULL,
      kind INTEGER NOT NULL,
      tags TEXT NOT NULL,  -- JSON array
      content TEXT NOT NULL,
      sig TEXT NOT NULL
    );
    CREATE INDEX idx_events_pubkey ON events(pubkey);
    CREATE INDEX idx_events_kind ON events(kind);
    CREATE INDEX idx_events_created_at ON events(created_at);
    ```
  - Implement Tag indexing: `CREATE TABLE event_tags (event_id TEXT, key TEXT, value TEXT, ...)` for efficient `#p` and other tag filtering
  - Implement NIP-01 filter-to-SQL translation:
    - `ids` → `WHERE id IN (...)`
    - `authors` → `WHERE pubkey IN (...)`
    - `kinds` → `WHERE kind IN (...)`
    - `#p` → JOIN on `event_tags` WHERE key='p' AND value IN (...)
    - `since` → `WHERE created_at >= ?`
    - `until` → `WHERE created_at <= ?`
    - `limit` → `ORDER BY created_at DESC LIMIT ?`
  - Implement TTL-based deletion: `DeleteOlderThan()` for T24 cleanup (T12)
  - Use WAL mode for concurrent reads
  - Write tests: save 10 events, query with each filter type, verify results

  **Must NOT do**:
  - Do not use CGO SQLite drivers (must compile to single binary without C deps)
  - Do not implement caching layer (keep it simple)
  - Do not add migrations system (single schema, re-creatable)

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: [] — Go + SQLite performance

  **Parallelization**:
  - **Can Run In Parallel**: YES (with T5, T7, T8)
  - **Parallel Group**: Wave 2
  - **Blocks**: T9, T10, T12
  - **Blocked By**: T2

  **References**:
  - `modernc.org/sqlite` — Pure Go SQLite driver (no CGO)
  - `vertex-lab/nostr-sqlite` — Reference SQLite schema for Nostr events
  - NIP-01 filter semantics: `https://github.com/nostr-protocol/nips/blob/master/01.md#from-client-to-relay-sending-events-and-creating-subscriptions`

  **Acceptance Criteria**:

  **QA Scenarios**:
  ```
  Scenario: Save and retrieve event
    Tool: Bash
    Preconditions: Store implementation ready
    Steps:
      1. go test ./internal/storage/... -run TestSaveAndQuery -v
    Expected Result: Event saved, retrieved by ID
    Failure Indicators: Event not found or fields differ
    Evidence: .sisyphus/evidence/task-6-save-query.txt

  Scenario: Filter by kind returns correct events
    Tool: Bash
    Preconditions: 5 events of kind 1, 5 of kind 4 stored
    Steps:
      1. go test ./internal/storage/... -run TestFilterByKind -v
    Expected Result: Only kind 1 or kind 4 events returned
    Failure Indicators: Wrong kinds in results
    Evidence: .sisyphus/evidence/task-6-filter-kind.txt

  Scenario: Tag filter (#p) works
    Tool: Bash
    Preconditions: Events with p-tags stored
    Steps:
      1. go test ./internal/storage/... -run TestFilterByPTag -v
    Expected Result: Only events with matching p-tag returned
    Failure Indicators: Wrong events returned
    Evidence: .sisyphus/evidence/task-6-filter-ptag.txt

  Scenario: TTL delete works
    Tool: Bash
    Preconditions: Old and recent events stored
    Steps:
      1. go test ./internal/storage/... -run TestTTLDelete -v
    Expected Result: Only old events deleted, recent remain
    Evidence: .sisyphus/evidence/task-6-ttl.txt
  ```

  **Commit**: YES
  - Message: `feat(relay): add SQLite storage with NIP-01 filters`
  - Files: `internal/storage/store.go`, `internal/storage/schema.go`, `internal/storage/store_test.go`, `internal/storage/schema_test.go`

---

- [ ] 7. **Schnorr Signature Validation**

  **What to do**:
  - Implement `internal/crypto/schnorr.go` — event signature validation
  - Use `github.com/nbd-wtf/go-nostr`'s `nostr.Signature` type or implement via `github.com/decred/dcrd/dcrec/secp256k1/v4`
  - Implement `VerifyEventSignature(event) bool`:
    1. Recompute event ID: SHA-256 of serialized event fields (NIP-01: `[0, pubkey, created_at, kind, tags, content]`)
    2. Verify ID matches `event.ID`
    3. Verify Schnorr signature: `event.Sig` over `event.ID` with `event.Pubkey`
  - Key validation: validate pubkey is 32 bytes, hex-encoded (64 hex chars)
  - Signature validation: validate sig is 64 bytes, hex-encoded (128 hex chars)
  - Write tests: valid signature passes, invalid signature fails, malformed keys rejected
  - Use constant-time comparison for signature verification

  **Must NOT do**:
  - Do not implement key generation here (JS handles it for clients)
  - Do not implement NIP-44 or encryption here (T8)
  - Do not implement NIP-17 gift wrap logic here (T9)

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: [] — cryptography in Go

  **Parallelization**:
  - **Can Run In Parallel**: YES (with T5, T6, T8)
  - **Parallel Group**: Wave 2
  - **Blocks**: T9
  - **Blocked By**: T2

  **References**:
  - NIP-01 event ID computation: `https://github.com/nostr-protocol/nips/blob/master/01.md#events-and-signatures`
  - BIP-340 Schnorr signatures: `https://github.com/bitcoin/bips/blob/master/bip-0340.mediawiki`
  - `github.com/nbd-wtf/go-nostr` — `nostr.Signature.Verify()` reference

  **Acceptance Criteria**:

  **QA Scenarios**:
  ```
  Scenario: Valid signature passes
    Tool: Bash
    Preconditions: Test fixtures with known-valid event
    Steps:
      1. go test ./internal/crypto/... -run TestVerifyValidSig -v
    Expected Result: Test passes (true)
    Failure Indicators: Valid sig rejected
    Evidence: .sisyphus/evidence/task-7-valid-sig.txt

  Scenario: Invalid signature rejected
    Tool: Bash
    Preconditions: Test fixtures with tampered sig
    Steps:
      1. go test ./internal/crypto/... -run TestVerifyInvalidSig -v
    Expected Result: Test passes (false)
    Failure Indicators: Invalid sig accepted
    Evidence: .sisyphus/evidence/task-7-invalid-sig.txt

  Scenario: Malformed pubkey rejected
    Tool: Bash
    Preconditions: Test fixtures with bad hex
    Steps:
      1. go test ./internal/crypto/... -run TestVerifyBadKey -v
    Expected Result: Test passes (false)
    Failure Indicators: Bad pubkey accepted
    Evidence: .sisyphus/evidence/task-7-bad-key.txt
  ```

  **Commit**: YES
  - Message: `feat(relay): add Schnorr signature validation`
  - Files: `internal/crypto/schnorr.go`, `internal/crypto/schnorr_test.go`

---

- [ ] 8. **NIP-44 Encryption Implementation (Go)**

  **What to do**:
  - Implement `internal/crypto/nip44.go` following NIP-44 spec exactly
  - NIP-44 encryption scheme:
    1. ECDH: derive shared point from sender's secp256k1 private key and recipient's public key
    2. HKDF: derive 76-byte key material (32 key + 32 nonce + 12 iv) from shared point using `SHA-256`
    3. Encryption: ChaCha20-Poly1305 with derived key + iv
    4. Output: base64(version || nonce || ciphertext || mac)
  - Implement `Encrypt(plaintext string, privateKey, publicKey []byte) (string, error)`
  - Implement `Decrypt(ciphertext string, privateKey, publicKey []byte) (string, error)`
  - Use `golang.org/x/crypto/chacha20poly1305` for ChaCha20-Poly1305
  - Use `golang.org/x/crypto/hkdf` for key derivation
  - Use secp256k1 ECDH from `github.com/decred/dcrd/dcrec/secp256k1/v4`
  - Write tests: encrypt → decrypt roundtrip, wrong key fails, tampered ciphertext fails
  - Version prefix handling: check version byte, support future versions gracefully
  - Test with known test vectors from NIP-44 spec

  **Must NOT do**:
  - Do not implement key generation (T14 handles it in JS)
  - Do not implement NIP-17 gift wrapping here (T9)
  - Do not roll custom crypto — strictly follow NIP-44

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: [] — cryptographic implementation in Go

  **Parallelization**:
  - **Can Run In Parallel**: YES (with T5, T6, T7)
  - **Parallel Group**: Wave 2
  - **Blocks**: T9, T14 (JS NIP-44 mirrors this)
  - **Blocked By**: T2

  **References**:
  - NIP-44 spec: `https://github.com/nostr-protocol/nips/blob/master/44.md` — MUST follow exactly
  - `golang.org/x/crypto/chacha20poly1305` — ChaCha20-Poly1305 AEAD
  - `golang.org/x/crypto/hkdf` — HKDF key derivation
  - NIP-44 test vectors: `https://github.com/nostr-protocol/nips/blob/master/44.md#test-vectors`

  **Acceptance Criteria**:

  **QA Scenarios**:
  ```
  Scenario: Encrypt → Decrypt roundtrip
    Tool: Bash
    Preconditions: Key pair test fixtures
    Steps:
      1. go test ./internal/crypto/... -run TestNIP44Roundtrip -v
    Expected Result: plaintext == decrypted(encrypted(plaintext))
    Failure Indicators: Mismatch or error
    Evidence: .sisyphus/evidence/task-8-roundtrip.txt

  Scenario: Wrong key cannot decrypt
    Tool: Bash
    Preconditions: Two different key pairs
    Steps:
      1. go test ./internal/crypto/... -run TestNIP44WrongKey -v
    Expected Result: Decryption fails with error
    Failure Indicators: Decryption succeeds with wrong key
    Evidence: .sisyphus/evidence/task-8-wrong-key.txt

  Scenario: Tampered ciphertext rejected
    Tool: Bash
    Preconditions: Valid ciphertext
    Steps:
      1. go test ./internal/crypto/... -run TestNIP44Tampered -v
    Expected Result: Decryption fails (integrity check catches tampering)
    Failure Indicators: Tampered data decrypts successfully
    Evidence: .sisyphus/evidence/task-8-tamper.txt
  ```

  **Commit**: YES
  - Message: `feat(crypto): implement NIP-44 encryption in Go`
  - Files: `internal/crypto/nip44.go`, `internal/crypto/nip44_test.go`

---

- [ ] 9. **NIP-17 Gift Wrap/Unwrap Implementation (Go)**

  **What to do**:
  - Implement `internal/relay/nip17.go` following NIP-17 spec exactly
  - NIP-17 uses a triple wrapping scheme:
    1. **Rumor** (kind 14): The actual message content. Has `p` tag with recipient pubkey.
    2. **Seal** (kind 13): Rumor encrypted via NIP-44, wrapped with sender's pubkey tag.
    3. **Gift Wrap** (kind 1059): Seal encrypted via NIP-44 with a RANDOM key (not sender's key), tagged to recipient's pubkey.
  - Implement `CreateGiftWrap(rumorContent string, senderPrivKey, recipientPubKey []byte) (Event, error)`
    - Generate random ephemeral key for gift wrap encryption
    - Create rumor event (kind 14) with content + p-tag
    - Create seal (encrypt rumor with `NIP44.Encrypt` using sender→recipient shared secret)
    - Create gift wrap (encrypt seal with `NIP44.Encrypt` using ephemeral→recipient shared secret)
    - Gift wrap has kind 1059, tags: `["p", <recipient>]`
    - Sign gift wrap with ephemeral key
  - Implement `UnwrapGiftWrap(giftWrap Event, recipientPrivKey []byte) (*Event, error)`
    - Try to decrypt gift wrap content with recipient's key (try each `p` tag)
    - If decrypted, parse as seal event (kind 13)
    - Decrypt seal content with recipient's key
    - If decrypted, parse as rumor event (kind 14)
    - Return the rumor
  - Implement recipient detection: `GetGiftWrapRecipients(event) []string` — extract `p` tags from kind 1059 events
  - Tests: wrap → unwrap roundtrip, wrong key fails, gift wrap with multiple p-tags

  **Must NOT do**:
  - Do not implement AUTH (NIP-42)
  - Do not implement group chats or kind-less wraps
  - Do not implement NIP-59 directly (NIP-17 is built on NIP-59, but we implement the specific NIP-17 kind flow)

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: [] — complex protocol logic in Go

  **Parallelization**:
  - **Can Run In Parallel**: YES (with T10, T11, T12, T13)
  - **Parallel Group**: Wave 3
  - **Blocks**: T19 (interop testing)
  - **Blocked By**: T5, T6, T8

  **References**:
  - NIP-17 spec: `https://github.com/nostr-protocol/nips/blob/master/17.md` — MUST follow exactly
  - NIP-59 spec: `https://github.com/nostr-protocol/nips/blob/master/59.md` — Gift wrap foundation
  - NIP-44 (from T8): used for encryption/decryption

  **Acceptance Criteria**:

  **QA Scenarios**:
  ```
  Scenario: Gift wrap roundtrip
    Tool: Bash
    Preconditions: Key pair fixtures, NIP-44 from T8
    Steps:
      1. go test ./internal/relay/... -run TestNIP17Roundtrip -v
    Expected Result: rumor content matches after unwrap
    Failure Indicators: Unwrap returns different content or error
    Evidence: .sisyphus/evidence/task-9-roundtrip.txt

  Scenario: Wrong recipient cannot unwrap
    Tool: Bash
    Preconditions: Two different recipient key pairs
    Steps:
      1. go test ./internal/relay/... -run TestNIP17WrongKey -v
    Expected Result: Unwrap fails (wrong recipient can't decrypt)
    Failure Indicators: Wrong key successfully unwraps
    Evidence: .sisyphus/evidence/task-9-wrong-key.txt

  Scenario: Gift wrap stored as kind 1059
    Tool: Bash
    Preconditions: Gift wrap created via CreateGiftWrap
    Steps:
      1. Inspect the created event's kind
    Expected Result: Event kind == 1059, tags contain ["p", <recipient>]
    Failure Indicators: Wrong kind or missing p-tag
    Evidence: .sisyphus/evidence/task-9-kind.txt
  ```

  **Commit**: YES
  - Message: `feat(relay): implement NIP-17 gift wrap protocol`
  - Files: `internal/relay/nip17.go`, `internal/relay/nip17_test.go`

---

- [ ] 10. **HTTP REST Bridge for FreeBasics Forms**

  **What to do**:
  - Implement `internal/bridge/freebasics.go` — HTTP handlers for FreeBasics form submissions
  - Endpoints:
    - `GET /fb/send` → Render HTML form: fields for recipient pubkey (64 hex), encrypted content blob (textarea), optional kind selector
    - `POST /fb/send` → Accept form POST, create Nostr event internally:
      - Validate pubkey format (64 hex chars)
      - Validate content size (<64KB)
      - Validate content as base64 (BYOE — user brings own encrypted blob)
      - Create Nostr Event with kind 1059 (or whatever kind user specified)
      - Sign event internally with the relay's key OR accept user-provided event JSON
      - Store via `IEventStore`
      - Return 302 redirect to confirmation page or JSON response
    - `GET /fb/receive?pubkey=<hex>` → Show list of events tagged to this pubkey
      - Query storage for `#p` tag matching pubkey
      - Render HTML table: created_at, kind, content (truncated), link to view full
    - `GET /fb/view?id=<event-id>` → Show single event in detail (full encrypted blob)
    - `POST /fb/submit-event` → Accept full Nostr event JSON (for advanced users), validate, store
  - All endpoints serve pure HTML (no JS) — forms, links, tables
  - Include `X-IORG-FBS` header detection in middleware (see T11)
  - Support `Accept: application/json` for API access
  - Rate limiting: apply T12 rate limiter to POST endpoints
  - CSRF protection: generate + validate tokens on form pages

  **Must NOT do**:
  - Do NOT add WebSocket endpoints to the bridge
  - Do NOT add JS dependencies to the FreeBasics forms
  - Do NOT implement NIP-17 unwrap on the server (plaintext never touches server)
  - Do NOT add authentication/authorization

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: [] — HTTP server design in Go

  **Parallelization**:
  - **Can Run In Parallel**: YES (with T9, T11, T12, T13)
  - **Parallel Group**: Wave 3
  - **Blocks**: T17 (FreeBasics UI)
  - **Blocked By**: T5, T6, T2

  **References**:
  - FreeBasics Participation Guidelines — form-based interaction is supported
  - BYOE model: users encrypt locally with PWA (T15/T16) or JS crypto lib (T14), paste blob
  - Nostr event creation without WebSocket: events can be created server-side and stored directly

  **Acceptance Criteria**:

  **QA Scenarios**:
  ```
  Scenario: FreeBasics send form renders
    Tool: Bash
    Preconditions: Relay + bridge running
    Steps:
      1. curl http://localhost:8080/fb/send
      2. grep -c '<form' response && grep -c '<script' response
    Expected Result: HTML form exists, 0 script tags
    Failure Indicators: No form, or script tags present
    Evidence: .sisyphus/evidence/task-10-form-render.txt

  Scenario: POST encrypted blob creates event
    Tool: Bash
    Preconditions: Relay running
    Steps:
      1. curl -X POST http://localhost:8080/fb/send \
           -d 'pubkey=<64-hex>&content=<base64-blob>&kind=1059'
      2. Check response (302 or JSON OK)
      3. Query WebSocket or /fb/receive to verify event stored
    Expected Result: Event created and queryable
    Failure Indicators: 400/500 error, or event not found
    Evidence: .sisyphus/evidence/task-10-post-success.txt

  Scenario: Invalid pubkey rejected
    Tool: Bash
    Preconditions: Relay running
    Steps:
      1. curl -X POST http://localhost:8080/fb/send \
           -d 'pubkey=invalid&content=<blob>'
    Expected Result: 400 error with validation message
    Failure Indicators: 200 OK or event stored with bad pubkey
    Evidence: .sisyphus/evidence/task-10-validation.txt
  ```

  **Commit**: YES
  - Message: `feat(bridge): add HTTP REST bridge for FreeBasics`
  - Files: `internal/bridge/freebasics.go`, `internal/bridge/router.go`, `internal/bridge/templates.go`, `internal/bridge/freebasics_test.go`

---

- [ ] 11. **FreeBasics Detection + Bifurcation Middleware**

  **What to do**:
  - Implement `internal/bridge/detect.go` — middleware that detects FreeBasics requests
  - Detection via `X-IORG-FBS` header (primary) — value is always `true`
  - Detection via `Via` header (secondary) — contains `Internet.org` string
  - Detection via User-Agent — contains `InternetOrgApp` string (Android app)
  - On detection:
    - Set context value: `ctx.IsFreeBasics = true`
    - Strip `script` and `link[rel=import]` elements from HTML responses (belt-and-suspenders since templates should already be JS-free)
    - No WebSocket upgrades allowed on FreeBasics-detected connections
    - No SVG or WOFF responses
  - Route bifurcation:
    - Root `/` → if FreeBasics: serve educational landing page (T13); if direct: serve JS app (T16)
    - `/learn/*` → educational pages (same for both, but JS-free HTML for FB, enhanced for direct)
    - `/app/*` → JS messaging app (redirect FB users to `/fb/...`)
    - `/fb/*` → bridge endpoints (T10)
  - Add `X-Robots-Tag: noindex` to all pages (prevent search indexing of proxy-served content)
  - Logging: log detection status for analytics

  **Must NOT do**:
  - Do not modify or inspect request bodies (just headers)
  - Do not block non-FreeBasics traffic
  - Do not add more than ~50ms latency to detection (it's just header checks)

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: [] — Go HTTP middleware

  **Parallelization**:
  - **Can Run In Parallel**: YES (with T9, T10, T12, T13)
  - **Parallel Group**: Wave 3
  - **Blocks**: None (routing helpers)
  - **Blocked By**: T5 (needs HTTP server)

  **References**:
  - FreeBasics Technical Guidelines: `https://developers.facebook.com/docs/internet-org/platform-technical-guidelines/` — X-IORG-FBS header, Via header
  - `X-Forwarded-For` and `X-IORG-FBS-UIP` for original IP detection

  **Acceptance Criteria**:

  **QA Scenarios**:
  ```
  Scenario: FreeBasics header detected
    Tool: Bash
    Preconditions: Relay running
    Steps:
      1. curl -H "X-IORG-FBS: true" http://localhost:8080/
    Expected Result: FreeBasics-optimized response (educational, no JS)
    Failure Indicators: JS-enabled app returned
    Evidence: .sisyphus/evidence/task-11-detected.txt

  Scenario: Direct request gets JS app
    Tool: Bash
    Preconditions: Relay running, JS app built
    Steps:
      1. curl http://localhost:8080/
      2. Check for script tag or app HTML
    Expected Result: JS app served (with script tags)
    Failure Indicators: Educational page served to direct request
    Evidence: .sisyphus/evidence/task-11-direct.txt

  Scenario: Via header also works
    Tool: Bash
    Preconditions: Relay running
    Steps:
      1. curl -H "Via: 1.1 Internet.org" http://localhost:8080/
    Expected Result: FreeBasics mode active
    Failure Indicators: Via header not recognized
    Evidence: .sisyphus/evidence/task-11-via.txt
  ```

  **Commit**: YES
  - Message: `feat(freebasics): add proxy detection and bifurcation`
  - Files: `internal/bridge/detect.go`, `internal/bridge/detect_test.go`, `internal/bridge/router.go`

---

- [ ] 12. **TTL Cleanup + Rate Limiting**

  **What to do**:
  - Implement `internal/relay/cleanup.go` — TTL-based event cleanup
    - Background goroutine that runs every 30 minutes
    - Calls `IEventStore.DeleteOlderThan(time.Now().Add(-24 * time.Hour))`
    - Logs deletion count
    - Graceful shutdown via context cancellation
    - Make TTL configurable (default 24h, env var `EVENT_TTL`)
  - Implement `internal/relay/ratelimit.go` — per-IP rate limiting
    - Use `github.com/puzpuzpuz/xsync/v3` concurrent map for IP counters
    - Rate: 100 events/hour per IP (configurable via `RATE_LIMIT` env var)
    - Window: sliding window with 1-minute granularity
    - Apply to: EVENT messages (WebSocket), POST /fb/send, POST /fb/submit-event
    - Don't apply to: REQ, CLOSE, GET requests
    - Reject with `["OK", <id>, false, "rate-limited: too many events"]` (WebSocket)
    - Reject with HTTP 429 for REST endpoints
    - Cleanup: remove stale IP entries every 10 minutes
  - Implement event size limit: reject events with `content > 64KB` (configurable)
  - Implement max subscriptions per connection: limit to 32

  **Must NOT do**:
  - Do not block other relay operations
  - Do not use external rate limiting service (in-process only)
  - Do not log IP addresses long-term

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: [] — concurrent Go utilities

  **Parallelization**:
  - **Can Run In Parallel**: YES (with T9, T10, T11, T13)
  - **Parallel Group**: Wave 3
  - **Blocks**: T21 (hardening)
  - **Blocked By**: T5, T6

  **References**:
  - `github.com/puzpuzpuz/xsync/v3` — concurrent map for rate limiting counters
  - Go `time.Ticker` for periodic cleanup
  - Context-based graceful shutdown pattern

  **Acceptance Criteria**:

  **QA Scenarios**:
  ```
  Scenario: TTL cleanup removes old events
    Tool: Bash
    Preconditions: Events stored with created_at > 24h ago
    Steps:
      1. Insert events with old timestamps directly into SQLite
      2. Trigger cleanup (or wait for tick)
      3. Query for those events
    Expected Result: Old events no longer present
    Failure Indicators: Old events still queryable
    Evidence: .sisyphus/evidence/task-12-ttl.txt

  Scenario: Rate limit blocks excess events
    Tool: Bash
    Preconditions: Rate limit set to 100/hr for testing (or lower for test)
    Steps:
      1. Use test mode: override rate limit to 5/min
      2. Send 6 events rapidly from same IP
      3. 6th event should be rejected
    Expected Result: 6th event returns OK false
    Failure Indicators: All 6 accepted
    Evidence: .sisyphus/evidence/task-12-ratelimit.txt

  Scenario: Oversized content rejected
    Tool: Bash
    Preconditions: Content limit set (64KB)
    Steps:
      1. Send event with 65KB content
    Expected Result: Event rejected (OK false or 413)
    Failure Indicators: Event accepted
    Evidence: .sisyphus/evidence/task-12-size-limit.txt
  ```

  **Commit**: YES
  - Message: `feat(relay): add TTL cleanup and rate limiting`
  - Files: `internal/relay/cleanup.go`, `internal/relay/ratelimit.go`, `internal/relay/cleanup_test.go`, `internal/relay/ratelimit_test.go`

---

- [ ] 13. **Educational Pages Content (5 HTML Pages)**

  **What to do**:
  - Create 5 educational HTML pages as Go templates in `internal/templates/pages/`:
    1. **`learn/crypto.html`** — "¿Qué es la Criptografía Asimétrica?": Explain public/private keys, how they work, visual diagram using CSS/HTML (no SVG). Include analogy (lockbox with two keys).
    2. **`learn/nostr.html`** — "¿Qué es Nostr y cómo funciona?": Explain relays, events, pubkeys as identities. Show message flow: Client A → Relay → Client B. Diagram using pure CSS boxes + arrows (Unicode).
    3. **`learn/architecture.html`** — "Arquitectura de KeyChat": Show the dual-face system. Diagram: FreeBasics → Proxy → Relay → DB, and separately Browser → Relay → DB. Explain how the same relay serves both.
    4. **`learn/key-generation.html`** — "Generación de Claves": Explain how to generate a key pair. Show examples of public key format (hex). Interactive form that generates keys server-side (using Go crypto) when submitted via POST (no JS).
    5. **`learn/private-messaging.html`** — "Mensajería Privada con NIP-17": Explain gift wrap, seal, rumor. Show the triple-wrapping process with a visual diagram. Explain BYOE for FreeBasics users (encrypt with PWA app, paste blob via clipboard).
  - Each page extends `base.html` and fits within the CSS design system
  - Each page <50KB total (HTML + embedded CSS)
  - Navigation between pages (previous/next links at bottom)
  - Add server-side key generation on `learn/key-generation.html`: POST form → Go generates secp256k1 key pair → displays public key + hex-encoded private key (with warning to save)
  - All text content in English (default) with Spanish lang attribute (`lang="es"`)
  - Add breadcrumb navigation: `Home > Learn > Topic`
  - Responsive: works on 240px-480px screens (feature phone optimization)

  **Must NOT do**:
  - No JavaScript in these pages
  - No SVG images
  - No external resources (everything self-contained)
  - No tracking or analytics

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [] — educational HTML/CSS content design

  **Parallelization**:
  - **Can Run In Parallel**: YES (with T9, T10, T11, T12)
  - **Parallel Group**: Wave 3
  - **Blocks**: T17 (FreeBasics navigation)
  - **Blocked By**: T3 (template system)

  **References**:
  - FreeBasics guidelines: no JS, no SVG, no WOFF, images <200KB
  - Feature phone design: high contrast, 16px+ body text, 44px+ touch targets
  - Nostr protocol diagrams from NIP-01, NIP-17 specs

  **Acceptance Criteria**:

  **QA Scenarios**:
  ```
  Scenario: All 5 pages render without errors
    Tool: Bash
    Preconditions: Templates compiled into Go binary, relay running
    Steps:
      1. for page in crypto nostr architecture key-generation private-messaging; do
           curl http://localhost:8080/learn/$page | grep -c '<html'
         done
    Expected Result: Each page returns valid HTML (grep count >= 1)
    Failure Indicators: Any page returns 404, 500, or malformed HTML
    Evidence: .sisyphus/evidence/task-13-pages.txt

  Scenario: No prohibited content in any page
    Tool: Bash
    Preconditions: Pages exist
    Steps:
      1. grep -r '<script' internal/templates/pages/ | wc -l
      2. grep -r -i '<svg' internal/templates/pages/ | wc -l
    Expected Result: 0 script tags, 0 SVG tags
    Failure Indicators: Prohibited tags found
    Evidence: .sisyphus/evidence/task-13-no-prohibited.txt

  Scenario: Key generation form works
    Tool: Bash
    Preconditions: Relay running
    Steps:
      1. curl -X POST http://localhost:8080/learn/key-generation
    Expected Result: Returns HTML with generated public key displayed
    Failure Indicators: Error or no key in response
    Evidence: .sisyphus/evidence/task-13-keygen.txt

  Scenario: Navigation between pages works
    Tool: Bash
    Preconditions: Relay running
    Steps:
      1. Extract all href="/learn/" links from one page
      2. curl each link and verify 200
    Expected Result: All internal links return 200
    Failure Indicators: 404 or 500 on any link
    Evidence: .sisyphus/evidence/task-13-nav.txt
  ```

  **Commit**: YES
  - Message: `feat(site): add 5 educational HTML pages`
  - Files: `internal/templates/pages/crypto.html`, `internal/templates/pages/nostr.html`, `internal/templates/pages/architecture.html`, `internal/templates/pages/key_generation.html`, `internal/templates/pages/private_messaging.html`, `internal/templates/pages/learn_index.html`

---

- [ ] 14. **JS NIP-44 + NIP-17 Implementation**

  **What to do**:
  - Implement `web/src/crypto/nip44.ts` — NIP-44 encryption in JS:
    - `deriveSharedSecret(privateKey hex, publicKey hex) → Uint8Array` using `@noble/curves/secp256k1` ECDH
    - `hkdfExpand(key, info, length)` using `@noble/hashes/hkdf`
    - `encrypt(plaintext string, privateKey, publicKey) → base64 string`
    - `decrypt(ciphertext base64 string, privateKey, publicKey) → string`
    - Use `@noble/ciphers/chacha` for ChaCha20-Poly1305
    - Follow NIP-44 exactly: version byte prepended, HKDF key derivation, ChaCha20-Poly1305 AEAD
    - Roundtrip tests with known test vectors from NIP-44 spec
    - MUST produce same results as Go NIP-44 implementation (T8) — cross-test between Go and JS
  - Implement `web/src/crypto/nip17.ts` — NIP-17 gift wrapping:
    - `createGiftWrap(rumorContent, senderPrivKey, recipientPubKey) → NostrEvent`
    - `unwrapGiftWrap(giftWrap, recipientPrivKey) → rumorContent`
    - Generate ephemeral keys using `@noble/curves/secp256k1` `utils.randomPrivateKey()`
    - Event signing: implement Schnorr signing for secp256k1
  - Implement `web/src/crypto/keys.ts`:
    - `generateKeyPair() → {privKey, pubKey}` using `@noble/curves/secp256k1`
    - `privateKeyToPublicKey(privKey) → pubKey`
    - Key format: hex-encoded 32-byte private key, hex-encoded 32-byte public key (NIP-01 format)
  - Write comprehensive tests: encrypt/decrypt roundtrip, cross-implementation (Go vs JS), gift wrap roundtrip, wrong key rejection
  - Export all functions from `web/src/crypto/index.ts`

  **Must NOT do**:
  - Do not use any non-noble crypto library (noble is audited and maintained)
  - Do not implement custom crypto — strictly follow NIP-44 and NIP-17 specs
  - Do not implement WebSocket client here (T15)
  - Do not implement UI here (T16)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: [] — TypeScript cryptography implementation

  **Parallelization**:
  - **Can Run In Parallel**: YES (with T9, T10, T11, T12, T13)
  - **Parallel Group**: Wave 3
  - **Blocks**: T16 (messaging UI needs crypto)
  - **Blocked By**: T4 (JS scaffolding), references Go NIP-44 from T8

  **References**:
  - NIP-44 spec: `https://github.com/nostr-protocol/nips/blob/master/44.md` — MUST follow exactly
  - NIP-17 spec: `https://github.com/nostr-protocol/nips/blob/master/17.md`
  - NIP-01 Schnorr signing: BIP-340 with secp256k1
  - `@noble/curves` secp256k1: `getSharedSecret`, `sign`, `verify`
  - `@noble/ciphers/chacha`: `chacha20poly1305`
  - `@noble/hashes/hkdf`: `hkdf`, `expand`
  - Cross-test: Go T8 NIP-44 test vectors must produce same results

  **Acceptance Criteria**:

  **QA Scenarios**:
  ```
  Scenario: JS encrypt → decrypt roundtrip
    Tool: Bash + node
    Preconditions: NIP-44 implementation complete
    Steps:
      1. node -e "const {encrypt, decrypt} = require('./web/dist/crypto/nip44'); ..."
      OR
      2. npx vitest run --testPathPattern nip44
    Expected Result: plaintext === decrypt(encrypt(plaintext))
    Failure Indicators: Roundtrip fails
    Evidence: .sisyphus/evidence/task-14-js-roundtrip.txt

  Scenario: Cross-test: Go encrypt → JS decrypt
    Tool: Bash + node + Go
    Preconditions: Both implementations complete
    Steps:
      1. Go encrypts known plaintext with known keys → output base64
      2. node decrypts the base64 with same keys
    Expected Result: JS decryption matches Go encryption
    Failure Indicators: Cross-implementation mismatch
    Evidence: .sisyphus/evidence/task-14-cross.txt

  Scenario: Gift wrap roundtrip (JS → JS)
    Tool: Bash + node
    Preconditions: NIP-17 + NIP-44 implemented
    Steps:
      1. npx vitest run --testPathPattern nip17
    Expected Result: rumor content matches after wrap → unwrap
    Failure Indicators: Wrap/unwrap fails
    Evidence: .sisyphus/evidence/task-14-giftwrap.txt

  Scenario: Key generation produces valid keys
    Tool: Bash + node
    Preconditions: keys.ts implemented
    Steps:
      1. npx vitest run --testPathPattern keys
    Expected Result: Generated keys are valid (64 hex chars pubkey, 64 hex privkey)
    Failure Indicators: Invalid key format
    Evidence: .sisyphus/evidence/task-14-keys.txt
  ```

  **Commit**: YES
  - Message: `feat(crypto): implement NIP-44/17 in JS`
  - Files: `web/src/crypto/nip44.ts`, `web/src/crypto/nip17.ts`, `web/src/crypto/keys.ts`, `web/src/crypto/index.ts`, `web/src/crypto/nip44.test.ts`, `web/src/crypto/nip17.test.ts`, `web/src/crypto/keys.test.ts`

---

- [ ] 15. **JS PWA Crypto Utility (Offline Key Management + Clipboard + Blob Transport)**

  **What to do**:
  - Build on top of T14 (`web/src/crypto/`) to create the PWA utility layer in `web/src/pwa/`:
  - Implement `web/src/pwa/keystore.ts` — Persistent key management (100% client-side, offline):
    - `initKeyStore() → Promise<void>` — initialize from localStorage, check for existing keys
    - `generateAndSaveKeyPair() → {privKey, pubKey}` — generate via T14 `generateKeyPair()`, save to localStorage
    - `loadKeyPair() → {privKey, pubKey} | null` — load from localStorage
    - `exportPublicKey() → string` — get hex pubkey for sharing
    - `exportPrivateKey() → string` — get hex privkey (with warning about security)
    - `importKeyPair(privKeyHex, pubKeyHex) → boolean` — import existing keys
    - `hasKeys() → boolean` — check if keys exist
    - `clearKeys() → void` — remove keys (factory reset)
    - Keys stored ONLY in localStorage — NEVER sent to server
  - Implement `web/src/pwa/contacts.ts` — Contact management:
    - `addContact(name, pubKey) → Contact`
    - `removeContact(pubKey) → void`
    - `getContacts() → Contact[]`
    - `findContact(pubKey) → Contact | null`
    - Stored in localStorage as JSON array
  - Implement `web/src/pwa/message-blob.ts` — Blob format for FreeBasics clipboard transport:
    - `createOutboxBlob(encryptedMsg, recipientPubKey, senderPubKey) → string`
      - Format: JSON `{version, type, sender, recipient, ciphertext, created_at}`
      - This blob is what user copies to clipboard → pastes into FreeBasics form
    - `parseInboxBlob(json: string) → {sender, recipient, ciphertext, created_at}`
      - Parse blob received from FreeBasics forms (from relay response)
    - `copyToClipboard(text: string) → Promise<boolean>` — clipboard API wrapper
    - `readFromClipboard() → Promise<string | null>` — paste from clipboard
  - Implement `web/src/pwa/message-history.ts` — Local message history:
    - `saveMessage(msg)` — store sent/received message in localStorage
    - `getConversation(contactPubKey) → Message[]` — thread by contact
    - `getAllConversations() → {contact, messages}[]` — all threads
    - Messages stored locally as plaintext (already decrypted by user)
    - `exportConversation(contactPubKey) → string` — export as JSON (educational: show the crypto flow)
  - All modules exported from `web/src/pwa/index.ts`
  - All functions work 100% offline — NO WebSocket, NO fetch calls in this module
  - Clipboard API handles HTTPS requirement (PWA origin must be HTTPS or localhost)

  **Must NOT do**:
  - No WebSocket: this is the OFFLINE layer
  - No server communication: T16 handles HTTP transport for sending/receiving blobs
  - No auto-sync: user explicitly copies/pastes via clipboard
  - No keys in URL parameters or server requests
  - No Service Worker logic in this task (that's T18)

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: [] — TypeScript utility module, localStorage, clipboard API

  **Parallelization**:
  - **Can Run In Parallel**: YES (with T16, T17, T18)
  - **Parallel Group**: Wave 4
  - **Blocks**: T16 (PWA UI consumes this utility)
  - **Blocked By**: T4 (JS scaffolding), T14 (crypto primitives)

  **References**:
  - MDN Clipboard API: `https://developer.mozilla.org/en-US/docs/Web/API/Clipboard`
  - localStorage patterns: `https://developer.mozilla.org/en-US/docs/Web/API/Window/localStorage`
  - T14 crypto primitives: `web/src/crypto/nip44.ts`, `web/src/crypto/nip17.ts`, `web/src/crypto/keys.ts`
  - Blob format: JSON with fields `version`, `type`, `sender`, `recipient`, `ciphertext`, `created_at`

  **Acceptance Criteria**:

  **QA Scenarios**:
  ```
  Scenario: Generate keys and persist in localStorage
    Tool: Playwright
    Preconditions: PWA loaded at http://localhost:5173 (or https equivalent for clipboard)
    Steps:
      1. Open browser devtools → Application → Local Storage
      2. Call keystore.generateAndSaveKeyPair()
      3. Verify localStorage has keychat_privkey and keychat_pubkey entries
      4. Reload page, call keystore.loadKeyPair() — returns same keys
    Expected Result: Keys persist across page reload
    Failure Indicators: Keys lost after reload, or localStorage not populated
    Evidence: .sisyphus/evidence/task-15-keygen-persist.png

  Scenario: Encrypt message, copy blob to clipboard, parse it back
    Tool: Playwright
    Preconditions: Two keypairs generated (Alice, Bob), T14 crypto available
    Steps:
      1. Alice encrypts message "hello Bob" to Bob's pubkey
      2. createOutboxBlob(encrypted, bobPub, alicePub) → blob string
      3. copyToClipboard(blob) → true
      4. readFromClipboard() → blob string
      5. parseInboxBlob(blob) → {sender: alicePub, recipient: bobPub, ciphertext: ...}
    Expected Result: Clipboard roundtrip preserves blob exactly
    Failure Indicators: Clipboard copy/paste fails, blob doesn't parse
    Evidence: .sisyphus/evidence/task-15-clipboard.txt

  Scenario: Contact CRUD works
    Tool: Playwright
    Preconditions: PWA loaded
    Steps:
      1. addContact("Alice", "abc...123")
      2. getContacts() → [{name: "Alice", pubKey: "abc...123"}]
      3. findContact("abc...123") → {name: "Alice", ...}
      4. removeContact("abc...123")
      5. getContacts() → []
    Expected Result: Contact lifecycle works correctly
    Failure Indicators: Contact not found, or persists after removal
    Evidence: .sisyphus/evidence/task-15-contacts.txt

  Scenario: Message history persists conversations
    Tool: Playwright
    Preconditions: Contact "Bob" exists
    Steps:
      1. saveMessage({contact: bobPub, content: "hi", direction: "sent", timestamp: Date.now()})
      2. getConversation(bobPub) → [message]
      3. saveMessage({contact: bobPub, content: "hello back", direction: "received", ...})
      4. getConversation(bobPub) → [2 messages, ordered by timestamp]
    Expected Result: Messages stored and retrieved by conversation
    Failure Indicators: Wrong order, missing messages, wrong contact grouping
    Evidence: .sisyphus/evidence/task-15-history.txt
  ```

  **Commit**: YES
  - Message: `feat(pwa): implement offline crypto utility (keystore, contacts, clipboard, history)`
  - Files: `web/src/pwa/keystore.ts`, `web/src/pwa/contacts.ts`, `web/src/pwa/message-blob.ts`, `web/src/pwa/message-history.ts`, `web/src/pwa/index.ts`

---

- [ ] 16. **PWA UI (Key Management, Compose/Receive via Clipboard, History)**

  **What to do**:
  - Build `web/src/ui/` — Vanilla TypeScript UI components (no framework), consuming T15's PWA utility:
    - **Key Management Screen** (`#keys`): 
      - Uses T15 keystore: generate new key pair, display pubkey (copy button), private key (with security warning + copy button), import existing key
      - Shows QR code for pubkey (optional, educational)
      - "Factory reset" button with confirmation dialog
    - **Contacts Screen** (`#contacts`): 
      - Uses T15 contacts: add contact by pubkey+name, list contacts, remove contacts
      - Pubkey validation (64 hex chars), nickname editing
    - **Compose Screen** (`#compose`): 
      - Select contact from contacts list
      - Type message, click "Encrypt & Copy Blob"
      - Shows encrypted blob preview (educational: "This is what the ciphertext looks like")
      - Auto-copies blob to clipboard
      - Shows STEP-BY-STEP transport instructions:
        ```
        Step 1: ✅ Message encrypted
        Step 2: 📋 Blob copied to clipboard
        Step 3: 🌐 Open FreeBasics (or switch to browser tab with relay)
        Step 4: 📝 Paste blob into the Send form
        Step 5: ✅ Click Submit — relay stores your message
        ```
      - "Copy blob again" button, "Show raw JSON" toggle (educational)
    - **Inbox Screen** (`#inbox`): 
      - Paste blob from clipboard (paste button + textarea fallback)
      - Decrypt pasted blob → display plaintext message
      - Shows all previously decrypted messages per contact
      - "Paste & Decrypt" workflow instructions
    - **Message History** (`#history`): 
      - Uses T15 message-history: thread view per contact
      - Shows send/receive timestamps, contact name, decrypted content
      - Export conversation as JSON (educational: show the full crypto flow)
  - **Transport Screen** (`#transport`): 
    - Visual explanation of the clipboard-based workflow
    - Diagram showing: PWA (encrypt) → Clipboard → FreeBasics Form → Relay → FreeBasics Form → Clipboard → PWA (decrypt)
    - Links to FreeBasics `/fb/send` and `/fb/receive` for the actual transport
  - UI design:
    - Mobile-first responsive (works on 320px screens)
    - Clean, minimal — focus on educational value
    - Status badges: 🔒 encrypted, 📋 in clipboard, 📤 submitted, 📥 received
    - "Show technical details" toggle: raw Nostr event JSON, blob structure, encryption params
  - All crypto operations use T15 utility (which wraps T14) — ensure async where needed
  - Error handling: clipboard access denied (show fallback textarea), decryption failure (wrong key? corrupted blob?), import validation
  - Help tooltips: "What's a pubkey?", "What's a gift wrap?", "Why use clipboard?", "How FreeBasics works"
  - Entry point: `web/src/main.ts` renders the SPA, routing via hash (#keys, #contacts, #compose, #inbox, #history, #transport)
  - The PWA UI works 100% offline — all data is in localStorage, all crypto in-browser

  **Must NOT do**:
  - No React/Vue/Svelte — pure TypeScript + DOM API
  - No CSS framework — use the CSS design system from T3
  - No routing library — hash-based routing is sufficient
  - No state management library — localStorage + in-memory Map
  - No WebSocket: transport is clipboard-based, NOT real-time
  - No server-side keys: all crypto is client-side

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [] — Vanilla TypeScript SPA, clipboard API, offline-first design

  **Parallelization**:
  - **Can Run In Parallel**: YES (with T15, T17, T18)
  - **Parallel Group**: Wave 4
  - **Blocks**: T19 (integration testing)
  - **Blocked By**: T14 (crypto), T15 (PWA crypto utility), T4 (scaffolding)

  **References**:
  - T15 PWA utility: `web/src/pwa/keystore.ts`, `web/src/pwa/contacts.ts`, `web/src/pwa/message-blob.ts`, `web/src/pwa/message-history.ts`
  - T14 crypto primitives: `web/src/crypto/nip44.ts`, `web/src/crypto/nip17.ts`
  - Vanilla TypeScript SPA patterns (no framework needed)
  - MDN Clipboard API: `https://developer.mozilla.org/en-US/docs/Web/API/Clipboard`

  **Acceptance Criteria**:

  **QA Scenarios**:
  ```
  Scenario: Full create-encrypt-copy workflow
    Tool: Playwright
    Preconditions: PWA at http://localhost:5173 (HTTPS for clipboard), no existing keys
    Steps:
      1. Navigate to http://localhost:5173/#keys
      2. Click "Generate New Key Pair"
      3. Verify pubkey displayed (64 hex chars)
      4. Navigate to #contacts, add "Bob" with a valid 64-char hex pubkey
      5. Navigate to #compose, select Bob, type "hello Bob", click "Encrypt & Copy Blob"
      6. Verify blob preview shown (JSON with ciphertext field)
      7. Verify clipboard contains the blob string
    Expected Result: Key generated, contact saved, message encrypted and blob copied
    Failure Indicators: No key generated, blob not shown, clipboard empty
    Evidence: .sisyphus/evidence/task-16-compose-copy.png

  Scenario: Paste blob from clipboard and decrypt
    Tool: Playwright
    Preconditions: Alice has keys, Bob's contact saved, an encrypted blob for Bob is in clipboard
    Steps:
      1. Navigate to #inbox
      2. Click "Paste from Clipboard"
      3. Verify blob JSON is parsed correctly (sender, ciphertext fields shown)
      4. Click "Decrypt" — message appears as plaintext
    Expected Result: Blob pasted and decrypted successfully
    Failure Indicators: Clipboard read fails, decryption error, wrong plaintext
    Evidence: .sisyphus/evidence/task-16-paste-decrypt.txt

  Scenario: Fallback textarea for clipboard workaround
    Tool: Playwright
    Preconditions: PWA loaded, clipboard access limited (simulate restricted context)
    Steps:
      1. Navigate to #inbox
      2. Verify a textarea is shown as fallback
      3. Paste a known blob JSON into textarea
      4. Click "Decrypt"
    Expected Result: Decryption works via textarea fallback
    Failure Indicators: Textarea not shown, decryption fails pasted content
    Evidence: .sisyphus/evidence/task-16-textarea-fallback.png

  Scenario: Message history persists across sessions
    Tool: Playwright
    Preconditions: Keys exist, at least one decrypted message
    Steps:
      1. Navigate to #history
      2. Verify conversation with Bob appears with timestamp
      3. Click "Export" — verify JSON download contains messages array
      4. Reload page, navigate to #history
    Expected Result: History persists (same conversations visible after reload)
    Failure Indicators: History empty after reload, export fails
    Evidence: .sisyphus/evidence/task-16-history-persist.png
  ```

  **Commit**: YES
  - Message: `feat(pwa): add PWA UI (keys, contacts, compose, inbox, history, transport)`
  - Files: `web/src/main.ts`, `web/src/ui/app.ts`, `web/src/ui/keygen.ts`, `web/src/ui/contacts.ts`, `web/src/ui/compose.ts`, `web/src/ui/inbox.ts`, `web/src/ui/message-history.ts`, `web/src/ui/transport.ts`, `web/src/styles/app.css`

---

- [ ] 17. **FreeBasics Face UI (Educational Navigation + BYOE Form)**

  **What to do**:
  - Build the FreeBasics navigation experience (pure HTML, server-rendered via Go templates):
    - Landing page (`/` with FreeBasics detection): Welcome to KeyChat, brief explanation, links to /learn/ and /fb/
    - `/fb/` index: FreeBasics portal — "Learn about Crypto" (→ /learn/), "Send a Message" (→ /fb/send), "Check Messages" (→ /fb/receive), "Install PWA" (→ /fb/pwa)
    - `/fb/send` → Enhanced HTML form for BYOE submission (from T10 bridge)
    - `/fb/receive` → Enhanced display of received events with pubkey input
    - `/fb/pwa` → Instructions for installing the PWA (once via WiFi/data, then works offline), step-by-step guide for the clipboard-based workflow
    - Navigation: breadcrumbs, sidebar links (all HTML/CSS, no JS)
  - Add progress/wizard for BYOE flow: Step 1 → Install PWA (once) → Step 2 → Encrypt message in PWA → Step 3 → Copy blob → Step 4 → Paste blob here → Step 5 → Submit
  - Add FAQ section: "Why can't I decrypt here?", "How do I install the PWA?", "What's a public key?"
  - All pages must work with JS disabled (test with Playwright JS disabled)
  - Ensure responsive design works on 240px widths (feature phones)

  **Must NOT do**:
  - No JavaScript
  - No redirects to JS app
  - No external dependencies
  - No SVGs or WOFF fonts

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [] — mobile UX, feature phone design

  **Parallelization**:
  - **Can Run In Parallel**: YES (with T15, T16, T18)
  - **Parallel Group**: Wave 4
  - **Blocks**: T20 (FreeBasics testing)
  - **Blocked By**: T10 (bridge), T11 (detection), T13 (educational pages)

  **References**:
  - FreeBasics Participation Guidelines: form-based interaction, no JS, images <200KB
  - Feature phone UX: large targets (44px+), readable fonts (16px+), linear layout
  - BYOE workflow documentation

  **Acceptance Criteria**:

  **QA Scenarios**:
  ```
  Scenario: FreeBasics portal renders without JS
    Tool: Playwright (JS disabled)
    Preconditions: Relay running
    Steps:
      1. Open browser with JS disabled, navigate to http://localhost:8080/
      2. With FreeBasics header: curl -H "X-IORG-FBS: true" http://localhost:8080/fb/
    Expected Result: Portal renders with all links working, no JS errors
    Failure Indicators: Page blank, links broken, or JS required
    Evidence: .sisyphus/evidence/task-17-portal.png

  Scenario: BYOE wizard guides user through steps
    Tool: Playwright (JS disabled)
    Preconditions: Relay running
    Steps:
      1. Navigate to /fb/send
      2. Follow wizard steps
      3. Submit encrypted blob
    Expected Result: Wizard completes, event stored
    Failure Indicators: Wizard broken or submission fails
    Evidence: .sisyphus/evidence/task-17-wizard.txt

  Scenario: All FreeBasics pages accessible via navigation
    Tool: Playwright (JS disabled)
    Preconditions: Relay running
    Steps:
      1. Start at /fb/
      2. Click every link, verify each page loads (200 OK, valid HTML)
    Expected Result: All pages accessible, no dead links
    Failure Indicators: 404, 500, or broken navigation
    Evidence: .sisyphus/evidence/task-17-nav.txt
  ```

  **Commit**: YES
  - Message: `feat(ui): add FreeBasics educational face UI`
  - Files: `internal/templates/pages/fb_portal.html`, `internal/templates/pages/fb_send.html`, `internal/templates/pages/fb_receive.html`, `internal/templates/pages/fb_tool.html`, `internal/templates/pages/fb_faq.html`, `internal/templates/pages/fb_wizard.html`

---

- [ ] 18. **PWA Manifest + Service Worker (Offline Install)**

  **What to do**:
  - Create `web/public/manifest.json` — Web App Manifest for PWA installability:
    ```json
    {
      "name": "KeyChat — Crypto Messenger",
      "short_name": "KeyChat",
      "description": "Learn cryptography and send encrypted messages via Nostr",
      "start_url": "/app/",
      "display": "standalone",
      "background_color": "#ffffff",
      "theme_color": "#1a1a2e",
      "icons": [
        { "src": "/icons/icon-192.png", "sizes": "192x192", "type": "image/png" },
        { "src": "/icons/icon-512.png", "sizes": "512x512", "type": "image/png" }
      ],
      "categories": ["education", "social", "utilities"],
      "scope": "/app/"
    }
    ```
  - Create `web/public/icons/` with placeholder PNG icons (192x192 and 512x512):
    - Simple lock+key icon or text-based icon (keep PNGs <10KB each for FreeBasics compatibility)
    - Use a simple PNG generator or base64-encoded minimal PNG
  - Create `web/sw.ts` — Service Worker for offline operation:
    - **Install event**: Pre-cache ALL app assets:
      - `index.html`, `manifest.json`, all CSS, all JS bundles (from Vite build)
      - App shell caching strategy: cache-first for app shell (everything needed to run offline)
    - **Activate event**: Clear old caches, take control of all clients immediately
    - **Fetch event**: 
      - App assets (JS/CSS/HTML): cache-first (serve from cache, fallback to network)
      - API calls (`/api/`, `/fb/`): network-only (never cache — messages are ephemeral)
      - Icons: cache-first
    - **Message event**: Listen for "skipWaiting" message from UI → call `skipWaiting()`
    - **Error handling**: If fetch fails and not in cache, return offline fallback page
  - Create `web/src/pwa/register-sw.ts` — Service Worker registration:
    - `registerServiceWorker() → Promise<boolean>`
    - Register SW with scope `/app/`
    - Handle registration errors gracefully (some browsers block SW)
    - Show "Install KeyChat" prompt when `beforeinstallprompt` event fires
    - Add "Install App" button that triggers deferred prompt
    - Track install state in localStorage (`keychat_installed`)
  - Integrate into Vite config:
    - Copy `manifest.json` and `web/public/icons/` to build output
    - Ensure SW is compiled as separate entry point (not bundled with app)
    - Configure Vite to inject `<link rel="manifest" href="/manifest.json">` and SW registration script into `index.html`
  - Test offline mode:
    - Load PWA once (caches all assets)
    - Disconnect network (DevTools → Network → Offline)
    - Verify PUA loads and works fully (keygen, encrypt, decrypt, clipboard, history)
    - Verify trying to use transport features shows "Go online to send" message gracefully

  **Must NOT do**:
  - No complex caching strategies — keep it simple: cache-first for app, network-only for API
  - No push notifications (not supported in FreeBasics, not needed for v1)
  - No background sync (complex, not needed for clipboard-based workflow)
  - No IndexedDB in SW (localStorage is sufficient, SW doesn't need it)
  - No non-PWA install methods (no native app wrapper)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: [] — PWA, Service Worker, Vite config

  **Parallelization**:
  - **Can Run In Parallel**: YES (with T15, T16, T17)
  - **Parallel Group**: Wave 4
  - **Blocks**: T19 (integration testing)
  - **Blocked By**: T4 (JS scaffolding), T16 (PWA UI exists to cache)

  **References**:
  - MDN Service Worker API: `https://developer.mozilla.org/en-US/docs/Web/API/Service_Worker_API`
  - MDN Web App Manifest: `https://developer.mozilla.org/en-US/docs/Web/Manifest`
  - Vite PWA guide: `https://vitejs.dev/guide/features.html`
  - Workbox patterns (for reference only — we're implementing manually): `https://developer.chrome.com/docs/workbox/`

  **Acceptance Criteria**:

  **QA Scenarios**:
  ```
  Scenario: Manifest loads and validates
    Tool: Playwright
    Preconditions: Vite dev server running
    Steps:
      1. Open http://localhost:5173/app/
      2. Check <link rel="manifest" href="/manifest.json"> exists in <head>
      3. Fetch /manifest.json, validate JSON structure
      4. Check browser DevTools → Application → Manifest shows valid PWA manifest
    Expected Result: Manifest valid with all required fields, icons exist and load
    Failure Indicators: Manifest missing, invalid JSON, icons 404
    Evidence: .sisyphus/evidence/task-18-manifest.txt

  Scenario: Service Worker registers and caches assets
    Tool: Playwright
    Preconditions: Vite production build completed (npx vite build)
    Steps:
      1. Serve built files with npx serve dist/
      2. Open browser, check console for "Service Worker registered"
      3. DevTools → Application → Service Workers → verify "activated" status
      4. DevTools → Cache → Cache Storage → verify app assets cached
    Expected Result: SW registers and activates, assets cached
    Failure Indicators: SW not registered, not activated, cache empty
    Evidence: .sisyphus/evidence/task-18-sw-activate.png

  Scenario: App works fully offline after first load
    Tool: Playwright
    Preconditions: Production build served, SW activated, assets cached
    Steps:
      1. Load app once (caches everything) — generate a key pair, save a contact
      2. DevTools → Network → toggle "Offline"
      3. Reload page
      4. Verify PWA loads without errors (no network requests failing)
      5. Verify keygen, contacts, compose screens work (localStorage-based)
      6. Verify clipboard operations work
    Expected Result: PWA fully functional offline
    Failure Indicators: White screen, console errors, features broken
    Evidence: .sisyphus/evidence/task-18-offline.png

  Scenario: API calls are NOT cached (network-only)
    Tool: Playwright
    Preconditions: SW active, relay running on localhost:8080
    Steps:
      1. Load app online
      2. Make a POST to /api/events (via fetch from devtools console)
      3. Go offline
      4. Try same POST again
    Expected Result: Offline API call fails with network error (not served from cache)
    Failure Indicators: API call succeeds offline (wrong caching strategy)
    Evidence: .sisyphus/evidence/task-18-api-no-cache.txt
  ```

  **Commit**: YES
  - Message: `feat(pwa): add manifest, Service Worker, and offline caching`
  - Files: `web/public/manifest.json`, `web/public/icons/icon-192.png`, `web/public/icons/icon-512.png`, `web/sw.ts`, `web/src/pwa/register-sw.ts`

---

- [ ] 19. **Nostr Interop Testing**

  **What to do**:
  - Test relay compatibility with REAL Nostr clients:
    - **WebSocket compliance**: Verify all NIP-01 messages (EVENT, REQ, CLOSE, EOSE, OK)
    - **noscl (CLI)**: `noscl relay wss://localhost:8080`, publish, subscribe
    - **nostr-tool**: Automated testing of event publishing and subscription
  - Automated tests via `websocat`:
    - Send kind 1059 from external client, verify relay stores it
    - Subscribe with complex filters (multiple kinds, authors, tags)
    - Verify EOSE sent correctly
    - Verify rate limiting applies to WebSocket
    - Verify TTL purges external events
  - Fix any compatibility issues found

  **Must NOT do**:
  - Do not modify Nostr spec — fix relay to be spec-compliant

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: [] — Nostr protocol testing

  **Parallelization**:
  - **Can Run In Parallel**: YES (with T20, T21, T22)
  - **Parallel Group**: Wave 5
  - **Blocks**: Final verification
  - **Blocked By**: T5, T6, T9, T16

  **References**:
  - NIP-01: `https://github.com/nostr-protocol/nips/blob/master/01.md`
  - `websocat`: CLI WebSocket client for scripting

  **Acceptance Criteria**:

  **QA Scenarios**:
  ```
  Scenario: External event published via websocat stored
    Tool: Bash + websocat
    Steps:
      1. Send valid Nostr event via WebSocket
      2. Query for it
    Expected Result: Published event appears in query results
    Evidence: .sisyphus/evidence/task-19-external-event.txt

  Scenario: Complex filter from external client
    Tool: Bash + websocat
    Steps:
      1. Subscribe with filter: kinds [1,4], #p with specific pubkey, limit 5
    Expected Result: Correct events returned, then EOSE
    Evidence: .sisyphus/evidence/task-19-filters.txt
  ```

  **Commit**: YES
  - Message: `test(interop): add Nostr client compatibility tests`
  - Files: `test/interop/websocket_test.sh`, `test/interop/README.md`

---

- [ ] 20. **FreeBasics Proxy Simulation + Tests**

  **What to do**:
  - Create `test/freebasics/proxy.go` — local proxy simulator that mimics FreeBasics:
    - Strips `<script>` tags from HTML
    - Strips `<svg>` elements
    - Strips `onclick`, `onload`, etc. inline JS handlers
    - Limits images >200KB (replace with placeholder)
    - Blocks WebSocket upgrade requests
  - Test scenarios:
    1. Educational pages through simulator → no JS, no SVG
    2. Navigation: all internal links work through simulated proxy
    3. Form submission through simulator → event stored correctly
  - Use Playwright with JS disabled
  - Verify page weights <100KB, images <200KB

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`

  **Parallelization**:
  - **Can Run In Parallel**: YES (with T19, T21, T22)
  - **Parallel Group**: Wave 5
  - **Blocked By**: T11, T13, T17

  **Acceptance Criteria**:

  **QA Scenarios**:
  ```
  Scenario: All educational pages pass proxy simulation
    Tool: Bash + Go test
    Steps:
      1. go test ./test/freebasics/... -run TestPagesThroughProxy
    Expected Result: All pages render without JS, no SVGs
    Evidence: .sisyphus/evidence/task-20-proxy-pages.txt

  Scenario: Zero JS in proxied responses
    Tool: Playwright (JS disabled)
    Steps:
      1. Navigate to / with X-IORG-FBS header
    Expected Result: Zero JS found in source
    Evidence: .sisyphus/evidence/task-20-no-js.txt
  ```

  **Commit**: YES
  - Message: `test(freebasics): add proxy simulation tests`
  - Files: `test/freebasics/proxy.go`, `test/freebasics/freebasics_test.go`

---

- [ ] 21. **Edge Case Hardening**

  **What to do**:
  - **XSS Prevention**: CSP headers, sanitize event content display, validate hex inputs
  - **Spam Prevention**: Same-event dedup, same-content rate limit
  - **Input Validation**: pubkey (64 hex), sig (128 hex), content max 64KB, created_at bounds, max 100 tags/event, max 100 items per filter array
  - **Error Handling**: Graceful shutdown (drain WS, close DB), panic recovery middleware, write timeouts
  - **Security Headers**: `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`, `Permissions-Policy`

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: [] — security hardening

  **Parallelization**:
  - **Can Run In Parallel**: YES (with T19, T20, T22)
  - **Parallel Group**: Wave 5
  - **Blocked By**: T10, T11, T12

  **Acceptance Criteria**:

  **QA Scenarios**:
  ```
  Scenario: XSS attempt sanitized
    Tool: Bash
    Steps:
      1. POST event with content: "<script>alert('xss')</script>hello"
      2. View on /fb/view/
    Expected Result: Script displayed as text, not executed
    Evidence: .sisyphus/evidence/task-21-xss.txt

  Scenario: Invalid hex pubkey rejected
    Tool: Bash
    Steps:
      1. POST to /fb/send with pubkey="not-hex"
    Expected Result: 400 error
    Evidence: .sisyphus/evidence/task-21-validation.txt
  ```

  **Commit**: YES
  - Message: `fix(security): harden edge cases and input validation`
  - Files: `internal/bridge/security.go`, `internal/relay/validation.go`

---

- [ ] 22. **Deployment Configuration + FreeBasics Submission Prep**

  **What to do**:
  - Docker multi-stage build (Go → scratch)
  - `docker-compose.yml` with relay
  - Env vars documented: `PORT`, `DB_PATH`, `EVENT_TTL`, `RATE_LIMIT`
  - Health endpoint at `/health`
  - FreeBasics submission doc: `docs/freebasics-submission.md` with description (max 35 chars), technical checklist, screenshots
  - Config validation on startup (fail fast if missing vars)

  **Recommended Agent Profile**:
  - **Category**: `quick`

  **Parallelization**:
  - **Can Run In Parallel**: YES (with T19, T20, T21)
  - **Parallel Group**: Wave 5
  - **Blocked By**: T5-T17 (functioning relay)

  **Acceptance Criteria**:

  **QA Scenarios**:
  ```
  Scenario: Docker image builds
    Tool: Bash
    Steps:
      1. docker build -t keychat:latest .
    Expected Result: Image builds successfully
    Evidence: .sisyphus/evidence/task-22-docker-build.txt

  Scenario: Docker container starts and serves
    Tool: Bash
    Steps:
      1. docker run -d -p 8080:8080 keychat:latest
      2. curl http://localhost:8080/health
    Expected Result: Health endpoint returns 200
    Evidence: .sisyphus/evidence/task-22-docker-start.txt
  ```

  **Commit**: YES
  - Message: `chore(deploy): add deployment config and FB submission prep`
  - Files: `Dockerfile`, `docker-compose.yml`, `.env.example`, `docs/freebasics-submission.md`

---

## Final Verification Wave

> 4 review agents run in PARALLEL. ALL must APPROVE. Present consolidated results to user and get explicit "okay" before completing.

- [ ] F1. **Plan Compliance Audit** — `oracle`
  Read plan end-to-end. For each "Must Have": verify implementation exists. For each "Must NOT Have": search for forbidden patterns. Check evidence files.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [ ] F2. **Code Quality + Security Review** — `unspecified-high`
  Run `go vet ./...` + `golangci-lint` + `go test ./...`. Review for hardcoded keys, console.log, commented-out code. Security: constant-time comparison for keys, CSP headers.
  Output: `Build [PASS/FAIL] | Lint [PASS/FAIL] | Tests [N pass/N fail] | Security [CLEAN/N issues] | VERDICT`

- [ ] F3. **Real Manual QA** — `unspecified-high`
  Execute EVERY QA scenario from EVERY task. Test integration: (1) PWA → clipboard → FB form → relay → FB response → clipboard → PWA, (2) Nostr client → relay → PWA, (3) PWA → relay → Nostr client.
  Save to `.sisyphus/evidence/final-qa/`.
  Output: `Scenarios [N/N pass] | Integration [N/N] | Edge Cases [N tested] | VERDICT`

- [ ] F4. **Scope Fidelity Check** — `deep`
  For each task: read "What to do", read actual diff. Verify 1:1 — everything built (no missing), nothing beyond spec (no creep). Check "Must NOT do" compliance.
  Output: `Tasks [N/N compliant] | Contamination [CLEAN/N issues] | VERDICT`

## Commit Strategy

- **T1**: `chore(project): scaffold Go module and JS project structure`
- **T2**: `feat(relay): implement Nostr data types and event validation`
- **T3**: `feat(site): add Go HTML templates and CSS design system`
- **T4**: `chore(js): bootstrap JS client project with deps`
- **T5**: `feat(relay): implement NIP-01 WebSocket protocol`
- **T6**: `feat(relay): add SQLite storage with NIP-01 filters`
- **T7**: `feat(relay): add Schnorr signature validation`
- **T8**: `feat(crypto): implement NIP-44 encryption in Go`
- **T9**: `feat(relay): implement NIP-17 gift wrap protocol`
- **T10**: `feat(bridge): add HTTP REST bridge for FreeBasics`
- **T11**: `feat(freebasics): add proxy detection and bifurcation`
- **T12**: `feat(relay): add TTL cleanup and rate limiting`
- **T13**: `feat(site): add 5 educational HTML pages`
- **T14**: `feat(crypto): implement NIP-44/17 in JS`
- **T15**: `feat(pwa): implement offline crypto utility (keystore, contacts, clipboard, history)`
- **T16**: `feat(pwa): add PWA UI (keys, contacts, compose, inbox, history)`
- **T17**: `feat(ui): add FreeBasics educational face UI`
- **T18**: `feat(pwa): add manifest, Service Worker, and offline caching`
- **T19**: `test(interop): add Nostr client compatibility tests`
- **T20**: `test(freebasics): add proxy simulation tests`
- **T21**: `fix(security): harden edge cases and input validation`
- **T22**: `chore(deploy): add deployment config and FB submission prep`

---

## Success Criteria

### Verification Commands
```bash
# Relay WebSocket basics
echo '["REQ", "test", {"kinds": [4], "limit": 1}]' | websocat ws://localhost:8080
# Expected: EVENT messages followed by EOSE

# REST bridge
curl -X POST http://localhost:8080/fb/send -d 'content=<encrypted-blob>&pubkey=<hex>&kind=1059'
# Expected: 302 redirect or JSON OK

# FreeBasics detection
curl -H "X-IORG-FBS: true" http://localhost:8080/ | grep -c '<script'
# Expected: 0 (no JS when FreeBasics header present)

# Educational page
curl http://localhost:8080/learn/crypto | grep -c '<html'
# Expected: 1 (valid HTML)

# PWA offline + clipboard roundtrip
# Load PWA, generate keys, encrypt message, copy blob, paste and decrypt
# (tested via Playwright QA scenarios in T16)

# TTL test
# Wait 24h+1m, then query for old event
# Expected: empty result
```

### Final Checklist
- [ ] All "Must Have" present
- [ ] All "Must NOT Have" absent
- [ ] All tests pass
- [ ] Nostr interop verified (Damus or Amethyst)
- [ ] FreeBasics proxy simulation passes
- [ ] User explicitly approves F1-F4 results
