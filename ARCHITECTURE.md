# KeyChat Architecture

> Sistema educativo de criptografía asimétrica + relay Nostr con mensajería privada.
> Doble cara: FreeBasics (zero-rated, sin JS) y acceso directo (full app).

---

## Tabla de Contenidos

1. [Vista General](#1-vista-general)
2. [Flujo FreeBasics: Contenido Educativo](#2-flujo-freebasics-contenido-educativo)
3. [Flujo FreeBasics: BYOE (Bring Your Own Encryption)](#3-flujo-freebasics-byoe-bring-your-own-encryption)
4. [Flujo Directo: Mensajería con JS](#4-flujo-directo-mensajería-con-js)
5. [Flujo Nostr: Cliente Externo](#5-flujo-nostr-cliente-externo)
6. [Flujo Interno: Gift Wrap NIP-17](#6-flujo-interno-gift-wrap-nip-17)
7. [Arquitectura del Relay](#7-arquitectura-del-relay)
8. [Modelo de Datos](#8-modelo-de-datos)

---

## 1. Vista General

```mermaid
graph TB
    subgraph "Internet"
        FB[FreeBasics Proxy<br/>internet.org / 0.freebasics.com]
        NC[Nostr Clients<br/>Damus, Amethyst, noscl]
    end

    subgraph "Dispositivo del Usuario"
        FB_APP[Navegador<br/>Sin JS]
        PWA_APP[PWA Offline App<br/>Keygen + Encrypt/Decrypt + Clipboard]
    end

    subgraph "Servidor KeyChat"
        DETECT[FreeBasics Detection<br/>X-IORG-FBS header]
        EDU[Educational Pages<br/>Go html/template]
        REST[REST Bridge<br/>Forms → Events]
        WS[WebSocket Relay<br/>NIP-01 + NIP-17]
        CRYPTO[NIP-44 Encryption<br/>ChaCha20-Poly1305]
        DB[(SQLite<br/>TTL: 24h)]
        RL[Rate Limiter<br/>100/h/IP]
    end

    FB_APP -->|"X-IORG-FBS: true"| FB
    FB -->|"JS Stripped"| DETECT
    DETECT -->|"Serve HTML"| EDU
    FB_APP -->|"POST encrypted blob"| FB
    FB --> REST
    REST --> RL
    RL --> DB

    PWA_APP -->|"WebSocket (online)"| WS
    PWA_APP -->|"Clipboard copy/paste"| FB_APP
    WS --> RL
    WS --> CRYPTO
    WS --> DB

    NC -->|"NIP-01 WebSocket"| WS

    EDU -.->|"Read-only"| DB
```

---

## 2. Flujo FreeBasics: Contenido Educativo

El usuario accede al contenido educativo **sin gastar datos**. El proxy de Internet.org quita el JS, y el relay detecta el header `X-IORG-FBS` para servir HTML puro.

```mermaid
sequenceDiagram
    participant U as User (Feature Phone)
    participant FB as FreeBasics Proxy (internet.org)
    participant R as KeyChat Relay
    participant DB as SQLite

    Note over U,FB: Zero-rated data

    U->>FB: GET http://keychat.example.com/
    FB->>R: GET / (X-IORG-FBS: true)
    R->>R: Detect X-IORG-FBS header
    R->>R: Select JS-free template
    R->>FB: HTML: Landing page (no JS, no SVG)
    FB->>U: Render HTML + strip any JS/SVG

    U->>FB: Click "Learn Asymmetric Crypto"
    FB->>R: GET /learn/crypto
    R->>FB: HTML: Educational page
    R->>DB: Log view (optional)
    FB->>U: Show crypto explanation + diagrams

    U->>FB: Click "Generate Keys"
    FB->>R: GET /learn/key-generation
    R->>FB: HTML form (no JS)
    FB->>U: Show key generation form

    U->>FB: Submit form (POST)
    FB->>R: POST /learn/key-generation
    R->>R: Generate secp256k1 key pair (Go crypto)
    R->>FB: HTML with pubkey + warning about privkey
    FB->>U: Display generated keys

    U->>FB: Click "How Nostr Works"
    FB->>R: GET /learn/nostr
    R->>FB: HTML: Nostr explanation + diagrams
    FB->>U: Show relay/client architecture
```

---

## 3. Flujo FreeBasics: BYOE (Bring Your Own Encryption)

El usuario tiene su clave privada **solo en su dispositivo móvil**, dentro de la PWA offline. La PWA cifra el mensaje (sin conexión a internet), y el usuario copia el blob cifrado al clipboard. Luego cambia al navegador FreeBasics (o al mismo navegador sin PWA) y pega el blob en un form. El servidor **nunca ve el texto plano ni la clave privada**.

```mermaid
sequenceDiagram
    participant PWA_A as Alice's PWA (Offline)
    participant A as Alice (Browser - FreeBasics)
    participant FB as FreeBasics Proxy
    participant R as KeyChat Relay
    participant DB as SQLite
    participant B as Bob (Browser - FreeBasics)
    participant PWA_B as Bob's PWA (Offline)

    Note over PWA_A: Alice has Bob's pubkey

    PWA_A->>PWA_A: Encrypt message with NIP-44
    Note right of PWA_A: ECDH shared secret<br/>ChaCha20-Poly1305<br/>Output: base64 blob
    PWA_A->>PWA_A: Create outbox blob (JSON wrapper)
    PWA_A->>PWA_A: Copy blob to clipboard
    Note right of PWA_A: "Message encrypted!<br/>Now paste into FreeBasics"

    A->>FB: GET /fb/send
    FB->>R: GET /fb/send
    R->>FB: HTML form (pubkey field + content textarea + paste instructions)
    FB->>A: Show BYOE form

    A->>A: Paste blob from clipboard into form
    A->>FB: POST /fb/send<br/>pubkey=<Bob's hex><br/>content=<blob from clipboard>
    FB->>R: POST /fb/send (X-IORG-FBS: true)
    R->>R: Validate pubkey + blob format
    R->>R: Create Nostr Event (kind 1059)<br/>Sign with ephemeral key
    R->>DB: INSERT event
    R->>FB: 302 Redirect to /fb/confirm
    FB->>A: "Message sent!"

    Note over B: Later, Bob checks his messages

    B->>FB: GET /fb/receive?pubkey=<Bob's hex>
    FB->>R: GET /fb/receive (X-IORG-FBS: true)
    R->>DB: SELECT events WHERE #p = Bob's pubkey
    R->>FB: HTML table of encrypted messages
    FB->>B: Show list: created_at, kind, truncated blob

    B->>FB: Click "View message"
    FB->>R: GET /fb/view?id=<event-id>
    R->>DB: SELECT event by id
    R->>FB: HTML: Full encrypted blob displayed
    FB->>B: Show encrypted blob

    B->>B: Select all → Copy to clipboard
    B->>PWA_B: Switch to PWA, paste blob
    PWA_B->>PWA_B: Parse blob, decrypt with NIP-44
    Note right of PWA_B: ECDH shared secret<br/>ChaCha20-Poly1305<br/>Plaintext recovered
    PWA_B->>PWA_B: Display: "Hi Bob! This is a secret message."
    PWA_B->>PWA_B: Save to local message history
```

---

## 4. Flujo Directo: PWA con REST Bridge

Cuando se accede **sin el proxy de FreeBasics**, el relay sirve la PWA offline-first. Toda la criptografía ocurre **client-side** (offline), y el envío/recepción de mensajes usa el REST bridge HTTP cuando hay conexión. El PWA **no usa WebSocket directo** — usa los mismos endpoints HTTP que FreeBasics, pero con JS para automatizar el copy/paste.

```mermaid
sequenceDiagram
    participant A as Alice's PWA
    participant R as KeyChat Relay (REST Bridge)
    participant DB as SQLite
    participant B as Bob's PWA

    Note over A: Alice loads the PWA (works offline)

    A->>R: GET / (no X-IORG-FBS)
    R->>A: PWA (JS enabled, offline-first)

    A->>A: Generate secp256k1 key pair (offline)
    Note right of A: @noble/curves/secp256k1
    A->>A: Store keys in localStorage

    A->>A: Import Bob's pubkey (contact)

    Note over A: Alice encrypts a message (OFFLINE)

    A->>A: Create rumor (kind 14): "Hello Bob!"
    A->>A: Create seal: Encrypt rumor with NIP-44
    A->>A: Generate ephemeral key
    A->>A: Create gift wrap: Encrypt seal with NIP-44
    A->>A: Sign gift wrap with ephemeral key
    Note right of A: Entirely offline:<br/>Web Crypto + @noble/ciphers

    Note over A: Alice sends (ONLINE - REST bridge)

    A->>R: POST /api/events<br/>{kind: 1059, content: "<gift wrap>", tags: [["p", "<Bob's pubkey>"]]}
    R->>R: Validate + rate limit
    R->>DB: INSERT event
    R->>A: {"id": "<event-id>", "success": true}

    Note over B: Bob checks his inbox

    B->>R: GET /api/events?pubkey=<Bob's hex>&kind=1059
    R->>DB: SELECT events WHERE #p = Bob's pubkey
    R->>B: {"events": [{kind: 1059, content: "<gift wrap>", ...}]}

    B->>B: Decrypt gift wrap offline<br/>→ seal → rumor: "Hello Bob!"
    B->>B: Display and save to local history

    Note over B: Bob replies

    B->>R: POST /api/events<br/>{kind: 1059, content: "<gift wrap to Alice>"}
    R->>DB: INSERT event
    R->>B: {"id": "<event-id>", "success": true}

    A->>R: GET /api/events?pubkey=<Alice's hex>&kind=1059
    R->>A: {"events": [{kind: 1059, ...}]}
    A->>A: Unwrap offline → display reply
```

---

## 5. Flujo Nostr: Cliente Externo

Un cliente Nostr existente (Damus, Amethyst, noscl) puede conectarse al relay KeyChat como a cualquier relay Nostr estándar.

```mermaid
sequenceDiagram
    participant C as Nostr Client (noscl/Damus)
    participant R as KeyChat Relay
    participant DB as SQLite

    C->>R: WebSocket connect
    R->>C: WebSocket open

    Note over C: Client queries for messages

    C->>R: ["REQ", "sync", {"kinds": [1059], "#p": ["<hex>"], "limit": 20}]

    R->>R: Parse filter: kinds=1059, #p=<hex>, limit=20
    R->>DB: SELECT FROM events JOIN tags<br/>WHERE kind=1059 AND tag_key='p' AND tag_val=<hex><br/>ORDER BY created_at DESC LIMIT 20
    R->>C: ["EVENT", "sync", {event data 1}]
    R->>C: ["EVENT", "sync", {event data 2}]
    R->>C: ...
    R->>C: ["EOSE", "sync"]

    Note over C: Client publishes a message

    C->>R: ["EVENT", {kind: 1, content: "hello", ...}]
    R->>R: Validate event ID hash (SHA-256)
    R->>R: Verify Schnorr signature (secp256k1)
    R->>R: Rate limit check
    R->>DB: INSERT event
    R->>C: ["OK", <id>, true, ""]

    C->>R: ["CLOSE", "sync"]
    R->>R: Remove subscription
```

---

## 6. Flujo Interno: Gift Wrap NIP-17

Cómo se construye y desarma un mensaje cifrado NIP-17 (triple wrapping).

### Construcción (Sender)

```mermaid
sequenceDiagram
    participant S as Sender
    participant NS as NIP-44 Crypto
    participant R as Relay

    S->>S: Rumor (kind 14)
    Note right of S: {"content": "Hello!", "kind": 14, "tags": [["p", "<recipient>"]]}

    S->>NS: Encrypt(rumor_json, sender_privkey, recipient_pubkey)
    NS->>NS: ECDH: shared_secret = secp256k1(sender_priv, recipient_pub)
    NS->>NS: HKDF: key_material(76 bytes) = expand(shared_secret)
    NS->>NS: ChaCha20-Poly1305: ciphertext = encrypt(rumor_json)
    NS-->>S: base64(ciphertext)

    S->>S: Seal (kind 13)
    Note right of S: {"content": "<encrypted_rumor>", "kind": 13, "tags": [["p", "<sender>"]]}

    S->>S: Generate ephemeral key pair (random)

    S->>NS: Encrypt(seal_json, ephemeral_privkey, recipient_pubkey)
    NS-->>S: base64(ciphertext)

    S->>S: Gift Wrap (kind 1059)
    Note right of S: {"content": "<encrypted_seal>", "kind": 1059, "tags": [["p", "<recipient>"]], "pubkey": "<ephemeral_pubkey>", "sig": "<schnorr_sig>"}

    S->>R: ["EVENT", {kind: 1059, tags: [["p", "<recipient>"]], ...}]

    Note over S,R: Relay sees ONLY kind 1059<br/>Can't read content (encrypted)<br/>Can't identify sender (ephemeral key)<br/>Only knows recipient (p-tag)
```

### Destrucción (Recipient)

```mermaid
sequenceDiagram
    participant R as Relay
    participant Re as Recipient
    participant NS as NIP-44 Crypto

    Re->>R: ["REQ", {...}]
    R-->>Re: ["EVENT", "sub", {kind: 1059, ...}]

    Re->>Re: Gift wrap received

    Re->>NS: Decrypt(content, recipient_privkey, ephemeral_pubkey)
    NS->>NS: ECDH: shared_secret = secp256k1(recipient_priv, ephemeral_pub)
    NS->>NS: HKDF: key_material(76 bytes)
    NS->>NS: ChaCha20-Poly1305: decrypt
    NS-->>Re: seal (kind 13)

    Re->>Re: Extract sender pubkey from seal tags

    Re->>NS: Decrypt(seal.content, recipient_privkey, sender_pubkey)
    NS->>NS: ECDH: shared_secret = secp256k1(recipient_priv, sender_pub)
    NS-->>Re: rumor (kind 14)

    Re->>Re: Rumor.content = "Hello!"
    Note right of Re: Original plaintext message
```

---

## 7. Arquitectura del Relay

```
┌──────────────────────────────────────────────────────────────┐
│                    cmd/keychat (main)                         │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────────────────────────────────────────────┐    │
│  │              HTTP Router (net/http)                    │    │
│  │                                                       │    │
│  │  ┌──────────────┐  ┌──────────────┐  ┌────────────┐  │    │
│  │  │ /learn/*     │  │ /fb/*        │  │ /app/*     │  │    │
│  │  │ Educational  │  │ REST Bridge  │  │ SPA        │  │    │
│  │  │ Pages (HTML) │  │ (HTML+JSON)  │  │ (JS App)   │  │    │
│  │  └──────────────┘  └──────────────┘  └────────────┘  │    │
│  │                                                       │    │
│  │  ┌──────────────────────────────────────────────────┐ │    │
│  │  │  FreeBasics Detection Middleware                  │ │    │
│  │  │  - X-IORG-FBS check                              │ │    │
│  │  │  - Via header check                              │ │    │
│  │  │  - Route bifurcation                             │ │    │
│  │  └──────────────────────────────────────────────────┘ │    │
│  │                                                       │    │
│  │  ┌──────────────────────────────────────────────────┐ │    │
│  │  │  WebSocket Handler (/ws)                          │ │    │
│  │  │  - NIP-01: REQ, EVENT, CLOSE, EOSE, OK           │ │    │
│  │  │  - Connection manager (goroutine per conn)        │ │    │
│  │  │  - Subscription manager                           │ │    │
│  │  │  - Ping/pong keepalive                            │ │    │
│  │  └──────────────────────────────────────────────────┘ │    │
│  └──────────────────────────────────────────────────────┘    │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐    │
│  │  internal/crypto                                      │    │
│  │  - NIP-44: ChaCha20-Poly1305 + HKDF + ECDH            │    │
│  │  - Schnorr: secp256k1 signature validate              │    │
│  └──────────────────────────────────────────────────────┘    │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐    │
│  │  internal/relay                                       │    │
│  │  - NIP-01: Event, Filter, Subscription types          │    │
│  │  - NIP-17: GiftWrap, Seal, Rumor                      │    │
│  │  - RateLimiter (xsync map, 100/h/IP)                  │    │
│  │  - TTL cleanup (24h, 30min ticker)                    │    │
│  └──────────────────────────────────────────────────────┘    │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐    │
│  │  internal/storage                                     │    │
│  │  - SQLite (modernc.org/sqlite, no CGO)                │    │
│  │  - events table                                       │    │
│  │  - event_tags table (for #p, #e lookups)             │    │
│  │  - NIP-01 filter → SQL translation                   │    │
│  │  - TTL-based deletion                                 │    │
│  │  - WAL mode (concurrent reads)                        │    │
│  └──────────────────────────────────────────────────────┘    │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐    │
│  │  internal/bridge                                      │    │
│  │  - FreeBasics form handlers (GET/POST)                 │    │
│  │  - BYOE event creation (REST → event → store)         │    │
│  │  - CSRF tokens for forms                              │    │
│  └──────────────────────────────────────────────────────┘    │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐    │
│  │  internal/templates                                   │    │
│  │  - base.html (Go html/template)                      │    │
│  │  - pages/ (5 educational pages)                      │    │
│  │  - fb/ (FreeBasics portal pages)                     │    │
│  │  - CSS design system (no framework, <50KB)           │    │
│  └──────────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────┐
│                    web/ (PWA Client)                          │
├──────────────────────────────────────────────────────────────┤
│  - src/crypto/: NIP-44 + NIP-17 + key generation            │
│  - src/pwa/: keystore, contacts, clipboard blobs, history    │
│  - src/ui/: key management, compose, inbox, contacts UI     │
│  - sw.ts: Service Worker (offline caching)                   │
│  - public/manifest.json: PWA install manifest               │
└──────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────┐
│                    web/ (JS SPA)                              │
├──────────────────────────────────────────────────────────────┤
│  src/                                                         │
│  ├── crypto/                                                  │
│  │   ├── nip44.ts  (ChaCha20-Poly1305, ECDH, HKDF)          │
│  │   ├── nip17.ts  (Gift wrap/unwrap)                        │
│  │   └── keys.ts   (Key generation, storage)                  │
│  ├── nostr/                                                   │
│  │   ├── client.ts (WebSocket client, reconnection)          │
│  │   ├── event.ts  (Event creation, signing)                  │
│  │   └── filter.ts (Filter helpers)                           │
│  ├── store/                                                   │
│  │   └── keys.ts   (localStorage key management)              │
│  └── ui/                                                      │
│      ├── app.ts       (SPA router)                            │
│      ├── keygen.ts    (Key generation screen)                 │
│      ├── contacts.ts  (Contact management)                    │
│      ├── compose.ts   (Message composition)                   │
│      └── inbox.ts     (Message inbox)                         │
└──────────────────────────────────────────────────────────────┘
```

### Dependencia entre Paquetes

```
cmd/keychat
    ├── internal/relay    (types, nip01, nip17, ratelimit, cleanup)
    │   ├── internal/crypto (nip44, schnorr)
    │   └── internal/storage (sqlite)
    ├── internal/bridge    (freebasics http handlers)
    │   └── internal/relay
    └── internal/templates (go html/templates)
```

---

## 8. Modelo de Datos

### SQLite Schema

```sql
-- Core events table
CREATE TABLE events (
    id         TEXT PRIMARY KEY,      -- NIP-01 event ID (SHA-256 hash, hex)
    pubkey     TEXT NOT NULL,         -- Author pubkey (64 hex chars)
    created_at INTEGER NOT NULL,      -- Unix timestamp
    kind       INTEGER NOT NULL,      -- Nostr kind (1, 4, 14, 1059, etc.)
    content    TEXT NOT NULL,         -- Event content (encrypted for NIP-17)
    sig        TEXT NOT NULL,         -- Schnorr signature (128 hex chars)
    created    TEXT NOT NULL DEFAULT (datetime('now'))  -- When stored
);

-- Tag index for efficient #p, #e lookups
CREATE TABLE event_tags (
    id       INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    key      TEXT NOT NULL,           -- Tag key (e.g. 'p', 'e', 't')
    value    TEXT NOT NULL            -- Tag value (e.g. pubkey hex)
);

-- Performance indexes
CREATE INDEX idx_events_pubkey     ON events(pubkey);
CREATE INDEX idx_events_kind       ON events(kind);
CREATE INDEX idx_events_created_at ON events(created_at);
CREATE INDEX idx_tags_key_value    ON event_tags(key, value);
CREATE INDEX idx_tags_event_id     ON event_tags(event_id);
```

### NIP-01 Filter → SQL Translation

| Filter | SQL |
|--------|------|
| `{"ids": ["a", "b"]}` | `WHERE id IN ('a', 'b')` |
| `{"authors": ["x"]}` | `WHERE pubkey IN ('x')` |
| `{"kinds": [1, 4]}` | `WHERE kind IN (1, 4)` |
| `{"#p": ["hex"]}` | `JOIN event_tags ON event_id = id WHERE key='p' AND value IN ('hex')` |
| `{"since": 100, "until": 200}` | `WHERE created_at >= 100 AND created_at <= 200` |
| `{"limit": 10}` | `ORDER BY created_at DESC LIMIT 10` |

### Estructura de Eventos Nostr

**Rumor (kind 14) — texto plano interno**
```json
{
  "id": "<sha256>",
  "pubkey": "<sender_hex>",
  "created_at": 1700000000,
  "kind": 14,
  "tags": [["p", "<recipient_hex>"]],
  "content": "Hello Bob! This is the plaintext message.",
  "sig": "<schnorr_sig>"
}
```

**Seal (kind 13) — rumor cifrado**
```json
{
  "id": "<sha256>",
  "pubkey": "<sender_hex>",
  "created_at": 1700000001,
  "kind": 13,
  "tags": [["p", "<sender_hex>"]],
  "content": "<NIP-44 encrypted rumor JSON>",
  "sig": "<schnorr_sig>"
}
```

**Gift Wrap (kind 1059) — seal cifrado (lo que ve el relay)**
```json
{
  "id": "<sha256>",
  "pubkey": "<ephemeral_hex>",     ← NO es el sender real
  "created_at": 1700000002,
  "kind": 1059,
  "tags": [["p", "<recipient_hex>"]],
  "content": "<NIP-44 encrypted seal JSON>",
  "sig": "<schnorr_sig firmado con ephemeral key>"
}
```

---

## Leyenda

```
─── → Request/Response síncrono
─‧─ → WebSocket (bidireccional, persistente)
-→  → Llamada interna/función
═══ → Límite de subsistema
```

---

*KeyChat v1 — Arquitectura documentada con Mermaid.js*
