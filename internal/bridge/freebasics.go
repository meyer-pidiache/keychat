package bridge

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type EventStore interface {
	SaveEventJSON(eventJSON string) error
	QueryEventsByPTag(pubkey string) ([]map[string]any, error)
	QueryEventByID(id string) (map[string]any, error)
}

type FreeBasicsBridge struct {
	rateLim RateLimiter
	baseURL string
	store   EventStore
}

type RateLimiter interface {
	Allow(key string) bool
}

func NewBridge(rateLim RateLimiter, baseURL string, store EventStore) *FreeBasicsBridge {
	return &FreeBasicsBridge{
		rateLim: rateLim,
		baseURL: baseURL,
		store:   store,
	}
}

func (fb *FreeBasicsBridge) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/fb/send", fb.handleSendForm)
	mux.HandleFunc("/fb/receive", fb.handleReceive)
	mux.HandleFunc("/fb/view", fb.handleViewEvent)
	mux.HandleFunc("/fb/submit-event", fb.handleSubmitEvent)
	mux.HandleFunc("/fb/confirm", fb.handleConfirm)

	mux.HandleFunc("/api/events", fb.handleAPIEvents)
}

func (fb *FreeBasicsBridge) handleSendForm(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		fb.renderForm(w, r)
	case "POST":
		fb.handlePostSend(w, r)
	default:
		http.Error(w, "method not allowed", 405)
	}
}

func (fb *FreeBasicsBridge) renderForm(w http.ResponseWriter, r *http.Request) {
	html := `<!doctype html><html lang="es"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"><link rel="icon" href="/static/icons/keychat-32.png" type="image/png"><link rel="icon" href="/static/icons/keychat.svg" type="image/svg+xml"><title>Enviar Mensaje — KeyChat</title><link rel="stylesheet" href="/static/css/reset.css"><link rel="stylesheet" href="/static/css/main.css"></head><body>
<a href="#main-content" class="skip-link">Saltar al contenido</a>
<header class="site-header"><nav class="nav" aria-label="Principal"><a href="/" class="nav-logo">KeyChat</a><ul class="nav-links"><li><a href="/fb/send">Enviar</a></li><li><a href="/fb/receive">Recibir</a></li></ul></nav></header>
<main id="main-content" class="main-content"><h1>Enviar Mensaje Cifrado</h1>
<p>Pega el blob cifrado desde la aplicación KeyChat en tu dispositivo.</p>
<form method="POST" action="/fb/send">
<label for="pubkey">Clave pública del destinatario (64 caracteres hex):</label>
<input type="text" id="pubkey" name="pubkey" pattern="[0-9a-fA-F]{64}" maxlength="64" required placeholder="Destinatario (64 hex)">
<label for="content">Blob cifrado (base64):</label>
<textarea id="content" name="content" rows="8" required placeholder="Pega aquí el blob cifrado..."></textarea>
<label for="kind">Tipo de evento:</label>
<select id="kind" name="kind"><option value="1059">Mensaje cifrado (1059)</option><option value="1">Nota (1)</option></select>
<button type="submit" class="btn">Enviar mensaje</button>
</form>
<p><a href="/fb/receive">Recibir mensajes</a></p></main>
<footer class="site-footer"><p>KeyChat — Proyecto educativo</p></footer></body></html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func (fb *FreeBasicsBridge) handlePostSend(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", 400)
		return
	}
	pubkey := strings.TrimSpace(r.FormValue("pubkey"))
	content := strings.TrimSpace(r.FormValue("content"))
	kindStr := r.FormValue("kind")

	if _, err := hex.DecodeString(pubkey); err != nil || len(pubkey) != 64 {
		http.Error(w, "pubkey inválido: debe ser 64 caracteres hex", 400)
		return
	}
	if len(content) > 65536 {
		http.Error(w, "contenido demasiado grande: máximo 64KB", 413)
		return
	}
	kind, err := strconv.Atoi(kindStr)
	if err != nil || kind < 1 {
		kind = 1059
	}

	ip := r.RemoteAddr
	if fb.rateLim != nil && !fb.rateLim.Allow(ip) {
		http.Error(w, "demasiadas solicitudes", 429)
		return
	}

	log.Printf("fb/send: pubkey=%s kind=%d content_len=%d", pubkey[:8], kind, len(content))

	event := map[string]any{
		"pubkey":     pubkey,
		"created_at": time.Now().Unix(),
		"kind":       kind,
		"content":    content,
		"tags":       []any{[]any{"p", pubkey}},
		"id":         "",
		"sig":        "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
	}

	eventJSON, _ := json.Marshal(event)
	h := sha256.Sum256(eventJSON)
	id := hex.EncodeToString(h[:])
	event["id"] = id

	eventJSON, _ = json.Marshal(event)

	if fb.store != nil {
		if err := fb.store.SaveEventJSON(string(eventJSON)); err != nil {
			log.Printf("fb/send: store error: %v", err)
			http.Error(w, "error al guardar el evento", 500)
			return
		}
	}

	http.Redirect(w, r, "/fb/confirm", 302)
}

func (fb *FreeBasicsBridge) handleReceive(w http.ResponseWriter, r *http.Request) {
	pubkey := r.URL.Query().Get("pubkey")

	html := fmt.Sprintf(`<!doctype html><html lang="es"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"><link rel="icon" href="/static/icons/keychat-32.png" type="image/png"><link rel="icon" href="/static/icons/keychat.svg" type="image/svg+xml"><title>Recibir Mensajes — KeyChat</title><link rel="stylesheet" href="/static/css/reset.css"><link rel="stylesheet" href="/static/css/main.css"></head><body>
<a href="#main-content" class="skip-link">Saltar al contenido</a>
<header class="site-header"><nav class="nav" aria-label="Principal"><a href="/" class="nav-logo">KeyChat</a><ul class="nav-links"><li><a href="/fb/send">Enviar</a></li><li><a href="/fb/receive">Recibir</a></li></ul></nav></header>
<main id="main-content" class="main-content"><h1>Recibir Mensajes</h1>
<form method="GET" action="/fb/receive"><label for="pubkey">Tu clave pública (64 hex):</label><input type="text" id="pubkey" name="pubkey" pattern="[0-9a-fA-F]{64}" maxlength="64" required placeholder="Tu clave pública" value="%s"><button type="submit" class="btn">Buscar mensajes</button></form>`, pubkey)

	if pubkey != "" {
		html += fmt.Sprintf(`<p>Mostrando mensajes para: <code>%s</code></p>`, pubkey[:16]+"...")

		if fb.store != nil {
			events, err := fb.store.QueryEventsByPTag(pubkey)
			if err != nil {
				log.Printf("fb/receive: query error: %v", err)
			} else if len(events) == 0 {
				html += `<p><em>No hay mensajes disponibles</em></p>`
			} else {
				html += `<ul class="event-list">`
				for _, ev := range events {
					evID, _ := ev["id"].(string)
					evKind, _ := ev["kind"].(int64)
					evTime, _ := ev["created_at"].(int64)
					content, _ := ev["content"].(string)
					trunc := content
					if len(trunc) > 60 {
						trunc = trunc[:60] + "..."
					}
					html += fmt.Sprintf(`<li><a href="/fb/view?id=%s">[kind %d] %s</a> <span class="event-time">%s</span></li>`,
						evID, evKind, trunc, time.Unix(evTime, 0).Format("2006-01-02 15:04"))
				}
				html += `</ul>`
			}
		} else {
			html += `<p><em>No hay mensajes disponibles</em></p>`
		}
	}

	html += `<p><a href="/fb/send">Enviar mensaje</a></p></main>
<footer class="site-footer"><p>KeyChat — Proyecto educativo</p></footer></body></html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func (fb *FreeBasicsBridge) handleViewEvent(w http.ResponseWriter, r *http.Request) {
	eventID := r.URL.Query().Get("id")

	var content, pubkey, eventKind string
	var createdAt string

	if fb.store != nil {
		ev, err := fb.store.QueryEventByID(eventID)
		if err == nil && ev != nil {
			content, _ = ev["content"].(string)
			pk, _ := ev["pubkey"].(string)
			kind, _ := ev["kind"].(int64)
			ct, _ := ev["created_at"].(int64)
			pubkey = pk
			eventKind = strconv.FormatInt(kind, 10)
			createdAt = time.Unix(ct, 0).Format("2006-01-02 15:04:05")
		}
	}

	if pubkey == "" {
		content = "Evento no encontrado"
	}

	html := fmt.Sprintf(`<!doctype html><html lang="es"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"><link rel="icon" href="/static/icons/keychat-32.png" type="image/png"><link rel="icon" href="/static/icons/keychat.svg" type="image/svg+xml"><title>Evento — KeyChat</title><link rel="stylesheet" href="/static/css/reset.css"><link rel="stylesheet" href="/static/css/main.css"></head><body>
<a href="#main-content" class="skip-link">Saltar al contenido</a>
<header class="site-header"><nav class="nav" aria-label="Principal"><a href="/" class="nav-logo">KeyChat</a><ul class="nav-links"><li><a href="/fb/send">Enviar</a></li><li><a href="/fb/receive">Recibir</a></li></ul></nav></header>
<main id="main-content" class="main-content"><h1>Evento</h1>
<p><strong>ID:</strong> <code>%s</code></p>
<p><strong>Pubkey:</strong> <code>%s</code></p>
<p><strong>Kind:</strong> %s</p>
<p><strong>Fecha:</strong> %s</p>
<pre style="white-space:pre-wrap;word-break:break-all;">%s</pre>
<p><a href="/fb/receive">Volver</a></p></main>
<footer class="site-footer"><p>KeyChat — Proyecto educativo</p></footer></body></html>`, eventID, pubkey, eventKind, createdAt, content)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func (fb *FreeBasicsBridge) handleSubmitEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "use POST", 405)
		return
	}
	http.Error(w, "endpoint en desarrollo", 501)
}

func (fb *FreeBasicsBridge) handleConfirm(w http.ResponseWriter, r *http.Request) {
	html := `<!doctype html><html lang="es"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"><link rel="icon" href="/static/icons/keychat-32.png" type="image/png"><link rel="icon" href="/static/icons/keychat.svg" type="image/svg+xml"><title>Confirmación — KeyChat</title><link rel="stylesheet" href="/static/css/reset.css"><link rel="stylesheet" href="/static/css/main.css"></head><body>
<a href="#main-content" class="skip-link">Saltar al contenido</a>
<header class="site-header"><nav class="nav" aria-label="Principal"><a href="/" class="nav-logo">KeyChat</a><ul class="nav-links"><li><a href="/fb/send">Enviar</a></li><li><a href="/fb/receive">Recibir</a></li></ul></nav></header>
<main id="main-content" class="main-content"><h1>Mensaje Enviado</h1>
<p>Tu mensaje cifrado ha sido enviado al relay. El destinatario puede verificarlo usando /fb/receive.</p>
<p><a href="/fb/send" class="btn">Enviar otro</a> <a href="/fb/receive" class="btn btn-secondary">Recibir mensajes</a></p></main>
<footer class="site-footer"><p>KeyChat — Proyecto educativo</p></footer></body></html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func (fb *FreeBasicsBridge) handleAPIEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "message": "API endpoint (demo)"})
}
