package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/opc/keychat/internal/bridge"
	"github.com/opc/keychat/internal/middleware"
	"github.com/opc/keychat/internal/relay"
	"github.com/opc/keychat/internal/storage"
	"github.com/opc/keychat/internal/templates"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "keychat.db"
	}
	webDir := os.Getenv("WEB_DIR")
	if webDir == "" {
		webDir = "web"
	}

	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		log.Fatalf("storage error: %v", err)
	}
	defer store.Close()

	rateLim := relay.NewRateLimiter(100, 1*time.Hour)
	rl := relay.NewRelay(store, rateLim)

	tc := relay.NewTTLCleanup(store, 24*time.Hour)
	tc.Start()

	fbBridge := bridge.NewBridge(rateLim, "", store)

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/ws", rl.HandleWebSocket)
	mux.HandleFunc("/ws/", rl.HandleWebSocket)

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(webDir+"/static"))))
	mux.Handle("/dist/", http.StripPrefix("/dist/", http.FileServer(http.Dir(webDir+"/dist"))))

	mux.HandleFunc("/manifest.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, webDir+"/dist/manifest.json")
	})
	mux.HandleFunc("/sw.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, webDir+"/dist/sw.js")
	})

	fbBridge.RegisterRoutes(mux)

	mux.HandleFunc("/learn/crypto", templates.ServePage("crypto"))
	mux.HandleFunc("/learn/nostr", templates.ServePage("nostr"))
	mux.HandleFunc("/learn/keys", templates.ServePage("keys"))
	mux.HandleFunc("/learn/messaging", templates.ServePage("messaging"))
	mux.HandleFunc("/learn/architecture", templates.ServePage("architecture"))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		if bridge.IsFreeBasics(r) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(`<!doctype html><html lang="es"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"><link rel="icon" href="/static/icons/keychat-32.png" type="image/png"><link rel="icon" href="/static/icons/keychat.svg" type="image/svg+xml"><title>KeyChat</title><link rel="stylesheet" href="/static/css/reset.css"><link rel="stylesheet" href="/static/css/main.css"></head><body>
<a href="#main-content" class="skip-link">Saltar al contenido</a>
<header class="site-header"><nav class="nav" aria-label="Principal"><a href="/" class="nav-logo">KeyChat</a><ul class="nav-links"><li><a href="/learn/crypto">Criptografía</a></li><li><a href="/learn/nostr">Nostr</a></li><li><a href="/learn/keys">Claves</a></li><li><a href="/learn/messaging">Mensajería</a></li><li><a href="/learn/architecture">Arquitectura</a></li><li><a href="/fb/send">Enviar</a></li><li><a href="/fb/receive">Recibir</a></li></ul></nav></header>
<main id="main-content" class="main-content"><h1>KeyChat</h1>
<p>Plataforma educativa de criptografía asimétrica y relay Nostr con mensajería privada.</p>
<div class="action-row"><a href="/learn/crypto" class="btn">Aprender Criptografía</a><a href="/learn/nostr" class="btn">Aprender Nostr</a><a href="/fb/send" class="btn btn-secondary">Enviar Mensaje</a><a href="/fb/receive" class="btn btn-secondary">Recibir Mensajes</a></div>
<h2>Páginas Educativas</h2>
<ul><li><a href="/learn/crypto">Criptografía Asimétrica</a></li><li><a href="/learn/nostr">Protocolo Nostr</a></li><li><a href="/learn/keys">Generación de Claves</a></li><li><a href="/learn/messaging">Mensajería Privada NIP-17</a></li><li><a href="/learn/architecture">Arquitectura de KeyChat</a></li></ul></main>
<footer class="site-footer"><p>KeyChat — Proyecto educativo de código abierto</p></footer></body></html>`))
			return
		}
		http.ServeFile(w, r, webDir+"/dist/index.html")
	})

	handler := middleware.Chain(
		mux,
		middleware.Recovery,
		middleware.RequestLog,
		middleware.SecurityHeaders,
		middleware.MaxBodySize(2<<20),
		bridge.DetectFreeBasics,
	)

	server := &http.Server{
		Addr:           ":" + port,
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		log.Printf("KeyChat starting on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")
	tc.Stop()
}
