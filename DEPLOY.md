# Despliegue de KeyChat

## Requisitos del servidor
- Linux (aarch64 o amd64)
- 512 MB RAM mínimo, 256 MB disco libre
- Docker + Docker Compose (recomendado) o Go 1.26+ y Node 22+

## Opción A: Docker (recomendado)

```bash
# Construir y ejecutar
docker compose build
docker compose up -d

# Verificar
curl http://localhost:8080/
```

**Configuración vía variables de entorno:**
- `PORT` (default: 8080)
- `DB_PATH` (default: /data/keychat.db)

Los datos persisten en el volumen Docker `keychat-data`.

## Opción B: Compilación directa

```bash
# 1. Compilar frontend
cd web && npm ci && npm run build && cd ..

# 2. Compilar backend
go build -o keychat ./cmd/keychat/...

# 3. Ejecutar
DB_PATH=./data/keychat.db ./keychat
```

## Cloudflare Tunnel (recomendado — sin Nginx ni certbot)

Cloudflare Tunnel expone tu servidor local a Internet sin necesidad de puertos abiertos, IP pública, ni certificados SSL manuales. Cloudflare maneja el HTTPS automáticamente.

### 1. Requisitos en Cloudflare Dashboard
- Dominio apuntando a Cloudflare (nameservers de Cloudflare)
- Cuenta de Cloudflare (plan gratis es suficiente)

### 2. Instalar cloudflared en el servidor

```bash
# Descargar cloudflared (aarch64 para Oracle Linux)
curl -LO https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-arm64.rpm
sudo dnf install -y ./cloudflared-linux-arm64.rpm

# Verificar
cloudflared version
```

### 3. Autenticar y crear el túnel

```bash
# Autenticar (abre un navegador para autorizar)
cloudflared tunnel login

# Crear el túnel
cloudflared tunnel create keychat

# Esto crea un archivo en ~/.cloudflared/<tunnel-id>.json
# Y te da un subdominio como <tunnel-id>.cfargotunnel.com
```

### 4. Configurar el túnel

Crea `~/.cloudflared/config.yml`:

```yaml
tunnel: <tunnel-id>
credentials-file: /root/.cloudflared/<tunnel-id>.json

ingress:
  - hostname: tudominio.com
    service: http://localhost:8080
  - service: http_status:404
```

### 5. DNS — apuntar dominio al túnel

```bash
# Crea un registro CNAME en Cloudflare
cloudflared tunnel route dns keychat tudominio.com
```

### 6. Iniciar el túnel (como servicio)

```bash
# Instalar como servicio del sistema
cloudflared service install

# Iniciar
sudo systemctl start cloudflared
sudo systemctl enable cloudflared

# Ver estado
sudo systemctl status cloudflared
```

¡Listo! Cloudflare maneja SSL, DDoS protection, y caching. El túnel funciona aunque el servidor esté detrás de NAT/CGNAT (como Oracle Cloud).

### Nota sobre WebSocket
Cloudflare Tunnel soporta WebSocket sin configuración adicional. El relay Nostr en `/ws` funcionará automáticamente.

## FreeBasics (Internet.org) — Publicación

### Requisitos
1. **Dominio propio** ✅
2. **HTTPS habilitado** (Cloudflare lo provee automáticamente)
3. **Cuenta de Meta Developer** registrada
4. **Solicitud de inclusión** en el programa FreeBasics

### Pasos para publicar en FreeBasics

1. **Verificar que funciona**
   ```bash
   # Local: simular tráfico FreeBasics
   curl http://localhost:8080/?__fb=1
   
   # Público: el header X-IORG-FBS lo agrega la infra de Meta
   curl -H "X-IORG-FBS: true" https://tudominio.com/
   ```

2. **Registrar cuenta de desarrollador en Meta**
   - Ve a https://developers.facebook.com/
   - Crea una app

3. **Solicitar acceso a FreeBasics**
   - En tu app de Meta Developers:
     - Products → FreeBasics by Facebook
     - URL del sitio: `https://tudominio.com/`
     - URL de Política de Privacidad: `https://tudominio.com/`
   - El sitio debe funcionar sin JavaScript (ya está implementado vía detección FreeBasics)

4. **Enviar para revisión**
   - Completa el formulario de solicitud
   - La revisión toma de días a semanas
   - Recibirás correo cuando sea aprobado

### URLs públicas una vez desplegado

| Ruta | Descripción |
|------|-------------|
| `/` | Página principal (PWA con JS si acceso directo, HTML simple si FreeBasics) |
| `/fb/send` | Formulario para enviar mensajes cifrados |
| `/fb/receive` | Formulario para recibir mensajes |
| `/fb/confirm` | Confirmación de envío |
| `/learn/crypto` | Página educativa |
| `/learn/nostr` | Página educativa |
| `/learn/keys` | Página educativa |
| `/learn/messaging` | Página educativa |
| `/learn/architecture` | Página educativa |
| `/ws` | WebSocket relay Nostr (NIP-01) |
| `/manifest.json` | Manifiesto PWA |
| `/sw.js` | Service Worker |

## Verificación post-despliegue

```bash
# Prueba rápida
bash test/integration/relay_test.sh https://tudominio.com

# O manualmente
curl -I https://tudominio.com/
curl https://tudominio.com/fb/send
curl https://tudominio.com/learn/crypto
```
