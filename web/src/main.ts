import '../static/css/reset.css'
import '../static/css/main.css'
import { generateAndSaveKeyPair, hasKeys, loadKeyPair, exportPublicKey, exportPrivateKey, clearKeys, importPrivateKey } from './pwa/keystore'
import { getContacts, addContact, removeContact, findContact } from './pwa/contacts'
import { encrypt, decrypt } from './crypto/nip44'
import { hexToPublicKey, publicKeyToHex, validatePubkey } from './crypto/keys'
import { createOutboxBlob, parseInboxBlob, copyToClipboard } from './pwa/message-blob'
import { saveMessage, getConversation, getAllConversations, exportConversation } from './pwa/message-history'
import { RelayClient, NostrEvent } from './nostr/relay-client'
import { createEvent, signEvent } from './nostr/event-builder'

function $(sel: string): HTMLElement | null {
  return document.querySelector(sel)
}

function $$(sel: string): NodeListOf<HTMLElement> {
  return document.querySelectorAll(sel)
}

function showView(id: string) {
  $$('.view').forEach((el) => el.classList.add('hidden'))
  const view = document.getElementById('view-' + id)
  if (view) view.classList.remove('hidden')
  $$('.nav-link').forEach((el) => el.classList.remove('active'))
  const link = document.querySelector(`.nav-link[data-view="${id}"]`)
  if (link) link.classList.add('active')
}

function updatePubkeyBadge() {
  const badge = document.getElementById('pubkey-badge')
  if (!badge) return
  const pub = exportPublicKey()
  if (pub) {
    badge.textContent = `Clave: ${pub.slice(0, 16)}...`
    badge.classList.remove('hidden')
  } else {
    badge.classList.add('hidden')
  }
}

function registerSW() {
  if ('serviceWorker' in navigator) {
    navigator.serviceWorker.register('/dist/sw.js').catch(() => {})
  }
}

const relayClient = new RelayClient(getRelayURL())

function getRelayURL(): string {
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${proto}//${location.host}/ws`
}

function updateRelayStatus() {
  const badge = document.getElementById('pubkey-badge')
  if (!badge) return
  const connected = relayClient.isConnected()
  const statusEl = document.getElementById('relay-status')
  if (statusEl) {
    statusEl.textContent = connected ? '🟢 Relay' : '🔴 Relay'
    statusEl.className = connected ? 'relay-status connected' : 'relay-status disconnected'
  }
}

function initRelay() {
  relayClient.onConnectionChange((connected) => {
    updateRelayStatus()
    if (connected) {
      subscribeInbox()
    }
  })
  relayClient.connect()
}

function subscribeInbox() {
  const pubkey = exportPublicKey()
  if (!pubkey) return
  relayClient.subscribe('inbox', { kinds: [1059], '#p': [pubkey], limit: 20 }, (event) => {
    const list = document.getElementById('inbox-messages')
    if (list) {
      list.insertAdjacentHTML('afterbegin', `<div class="conversation-item">
        <strong>Nuevo mensaje</strong>
        <span class="conversation-preview">Evento de ${event.pubkey.slice(0, 16)}</span>
        <span class="conversation-time">${new Date(event.created_at * 1000).toLocaleDateString()}</span>
      </div>`)
    }
  })
}

async function publishEvent(event: NostrEvent): Promise<boolean> {
  try {
    relayClient.publish(event)
    return true
  } catch {
    return false
  }
}

function navigate() {
  const hash = location.hash.slice(1) || 'home'
  showView(hash)
  if (hash === 'contacts') renderContacts()
  if (hash === 'history') renderHistory()
  if (hash === 'inbox') renderInbox()
}

function init() {
  if (hasKeys()) {
    updatePubkeyBadge()
  }
  registerSW()
  initRelay()
  window.addEventListener('hashchange', navigate)
  navigate()
}

document.addEventListener('DOMContentLoaded', init)

$$('.nav-link').forEach((el) => {
  el.addEventListener('click', (e) => {
    e.preventDefault()
    const view = (e.currentTarget as HTMLElement).dataset.view
    if (view) location.hash = view
  })
})

$('#btn-generate')?.addEventListener('click', () => {
  generateAndSaveKeyPair()
  const privHex = exportPrivateKey()
  const pubHex = exportPublicKey()
  const privEl = document.getElementById('privkey-display')
  const pubEl = document.getElementById('pubkey-display')
  if (privEl) privEl.textContent = privHex
  if (pubEl) pubEl.textContent = pubHex
  updatePubkeyBadge()
  showView('keys')
})

$('#btn-import-key')?.addEventListener('click', () => {
  const hex = ($('#import-hex') as HTMLInputElement)?.value.trim()
  if (!hex) { alert('Pega una clave privada hex (64 caracteres)'); return }
  if (importPrivateKey(hex)) {
    const privEl = document.getElementById('privkey-display')
    const pubEl = document.getElementById('pubkey-display')
    if (privEl) privEl.textContent = exportPrivateKey()
    if (pubEl) pubEl.textContent = exportPublicKey()
    updatePubkeyBadge()
    showView('keys')
  } else {
    alert('Clave privada inválida')
  }
})

$('#btn-clear-keys')?.addEventListener('click', () => {
  if (confirm('¿Eliminar todas las claves? Esta acción no se puede deshacer.')) {
    clearKeys()
    updatePubkeyBadge()
    location.hash = 'home'
  }
})

$('#btn-copy-privkey')?.addEventListener('click', () => {
  copyToClipboard(exportPrivateKey())
})

$('#btn-copy-pubkey')?.addEventListener('click', () => {
  copyToClipboard(exportPublicKey())
})

function renderContacts() {
  const list = document.getElementById('contacts-list')
  if (!list) return
  const contacts = getContacts()
  if (contacts.length === 0) {
    list.innerHTML = '<p class="empty-state">No hay contactos. Agrega uno arriba.</p>'
    return
  }
  list.innerHTML = contacts.map((c) => `
    <div class="contact-item">
      <div class="contact-info">
        <strong>${escHtml(c.name)}</strong>
        <code>${c.pubkey.slice(0, 16)}...</code>
      </div>
      <button class="btn btn-small btn-danger" data-remove="${c.pubkey}">Eliminar</button>
    </div>
  `).join('')
  list.querySelectorAll('[data-remove]').forEach((btn) => {
    btn.addEventListener('click', () => {
      const pk = (btn as HTMLElement).dataset.remove || ''
      removeContact(pk)
      renderContacts()
    })
  })
}

$('#btn-add-contact')?.addEventListener('click', () => {
  const name = ($('#contact-name') as HTMLInputElement)?.value.trim()
  const pubkey = ($('#contact-pubkey') as HTMLInputElement)?.value.trim()
  if (!name) { alert('Escribe un nombre'); return }
  if (!validatePubkey(pubkey)) { alert('Clave pública inválida (64 hex)'); return }
  const result = addContact(name, pubkey)
  if (!result) { alert('El contacto ya existe o datos inválidos'); return }
  ;($('#contact-name') as HTMLInputElement).value = ''
  ;($('#contact-pubkey') as HTMLInputElement).value = ''
  renderContacts()
})

function renderComposeContactList() {
  const sel = document.getElementById('compose-contact') as HTMLSelectElement | null
  if (!sel) return
  const current = sel.value
  sel.innerHTML = '<option value="">Seleccionar contacto...</option>'
  for (const c of getContacts()) {
    sel.innerHTML += `<option value="${c.pubkey}">${escHtml(c.name)}</option>`
  }
  sel.value = current
}

document.addEventListener('DOMContentLoaded', () => {
  const observer = new MutationObserver(() => {
    if (!document.getElementById('view-compose')?.classList.contains('hidden')) {
      renderComposeContactList()
    }
    if (!document.getElementById('view-contacts')?.classList.contains('hidden')) {
      renderContacts()
    }
  })
  document.querySelectorAll('.view').forEach((el) => {
    observer.observe(el, { attributes: true, attributeFilter: ['class'] })
  })
})

$('#btn-compose-encrypt')?.addEventListener('click', async () => {
  if (!hasKeys()) { alert('Genera una clave primero'); return }
  const sel = $('#compose-contact') as HTMLSelectElement | null
  const recipient = sel?.value || ''
  const message = ($('#compose-message') as HTMLTextAreaElement)?.value
  if (!recipient) { alert('Selecciona un contacto'); return }
  if (!message) { alert('Escribe un mensaje'); return }

  const keys = loadKeyPair()
  if (!keys) { alert('No hay claves'); return }
  try {
    const recipientBytes = hexToPublicKey(recipient)
    console.log('[encrypt] privateKey len:', keys.privateKey.length)
    console.log('[encrypt] recipientBytes len:', recipientBytes.length, 'first byte:', recipientBytes[0])
    const ciphertext = encrypt(message, keys.privateKey, recipientBytes)
    const blob = createOutboxBlob(ciphertext, publicKeyToHex(keys.publicKey), recipient)
    saveMessage({ contactPubKey: recipient, content: message, direction: 'sent', timestamp: Date.now() })
    const resultEl = document.getElementById('encrypt-result')
    const blobEl = document.getElementById('encrypt-blob')
    if (resultEl) resultEl.classList.remove('hidden')
    if (blobEl) blobEl.textContent = blob
    copyToClipboard(blob)

    if (relayClient.isConnected()) {
      const pubkeyHex = publicKeyToHex(keys.publicKey)
      const privkeyHex = exportPrivateKey()
      const event = createEvent(pubkeyHex, 1059, [['p', recipient]], ciphertext)
      const signed = await signEvent(event, privkeyHex)
      await publishEvent(signed)
    }
  } catch (err) {
    alert(`Error al cifrar: ${err}`)
  }
})

$('#btn-copy-blob')?.addEventListener('click', () => {
  const blobEl = document.getElementById('encrypt-blob')
  if (blobEl?.textContent) copyToClipboard(blobEl.textContent)
})

function renderInbox() {
  const list = document.getElementById('inbox-messages')
  if (!list) return
  const msgs = getAllConversations()
  if (msgs.size === 0) {
    list.innerHTML = '<p class="empty-state">No hay mensajes descifrados.</p>'
    return
  }
  let html = ''
  msgs.forEach((msgs, pubkey) => {
    const contact = findContact(pubkey)
    const name = contact ? contact.name : pubkey.slice(0, 16)
    const last = msgs[msgs.length - 1]
    html += `<div class="conversation-item" data-pubkey="${pubkey}">
      <strong>${escHtml(name)}</strong>
      <span class="conversation-preview">${escHtml(last.content.slice(0, 60))}</span>
      <span class="conversation-time">${new Date(last.timestamp).toLocaleDateString()}</span>
    </div>`
  })
  list.innerHTML = html
  list.querySelectorAll('.conversation-item').forEach((el) => {
    el.addEventListener('click', () => {
      const pk = (el as HTMLElement).dataset.pubkey || ''
      renderConversationDetail(pk)
    })
  })
}

function renderConversationDetail(pubkey: string) {
  const detail = document.getElementById('inbox-detail')
  if (!detail) return
  const msgs = getConversation(pubkey)
  const contact = findContact(pubkey)
  const name = contact ? contact.name : pubkey.slice(0, 16)
  detail.innerHTML = `<h3>${escHtml(name)}</h3>
    <div class="conversation-thread">${msgs.map((m) => `
      <div class="message-bubble ${m.direction}">
        <div class="message-content">${escHtml(m.content)}</div>
        <div class="message-time">${new Date(m.timestamp).toLocaleString()}</div>
      </div>`).join('')}
    </div>
    <button id="btn-export-conv" class="btn btn-small" data-pubkey="${pubkey}">Exportar conversación</button>
    <button id="btn-back-inbox" class="btn btn-small btn-secondary">Volver</button>`
  detail.classList.remove('hidden')
  document.getElementById('inbox-list')?.classList.add('hidden')
  document.getElementById('btn-export-conv')?.addEventListener('click', () => {
    const pk = document.getElementById('btn-export-conv')?.dataset?.pubkey || ''
    const json = exportConversation(pk)
    const blob = new Blob([json], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url; a.download = `keychat-conversation-${pk.slice(0, 8)}.json`; a.click()
    URL.revokeObjectURL(url)
  })
  document.getElementById('btn-back-inbox')?.addEventListener('click', () => {
    detail.classList.add('hidden')
    document.getElementById('inbox-list')?.classList.remove('hidden')
  })
}

$('#btn-relay-connect')?.addEventListener('click', () => {
  relayClient.connect()
})

$('#btn-relay-disconnect')?.addEventListener('click', () => {
  relayClient.disconnect()
})

$('#btn-decrypt')?.addEventListener('click', () => {
  if (!hasKeys()) { alert('Genera una clave primero'); return }
  const blobText = ($('#decrypt-input') as HTMLTextAreaElement)?.value.trim()
  if (!blobText) { alert('Pega el blob cifrado'); return }
  const parsed = parseInboxBlob(blobText)
  if (!parsed) { alert('Blob inválido'); return }

  const keys = loadKeyPair()
  if (!keys) { alert('No hay claves'); return }
  try {
    const senderBytes = hexToPublicKey(parsed.sender)
    const plaintext = decrypt(parsed.ciphertext, keys.privateKey, senderBytes)
    saveMessage({ contactPubKey: parsed.sender, content: plaintext, direction: 'received', timestamp: Date.now() })
    const resultEl = document.getElementById('decrypt-result')
    const textEl = document.getElementById('decrypt-text')
    if (resultEl) resultEl.classList.remove('hidden')
    if (textEl) textEl.textContent = plaintext
  } catch (err) {
    alert(`Error al descifrar: ${err}`)
  }
})

function renderHistory() {
  const container = document.getElementById('history-container')
  if (!container) return
  const convs = getAllConversations()
  if (convs.size === 0) {
    container.innerHTML = '<p class="empty-state">No hay historial de mensajes.</p>'
    return
  }
  let html = ''
  convs.forEach((msgs, pubkey) => {
    const contact = findContact(pubkey)
    const name = contact ? contact.name : pubkey.slice(0, 16)
    html += `<div class="history-group">
      <h3>${escHtml(name)} <code>${pubkey.slice(0, 16)}...</code></h3>
      <div class="history-messages">${msgs.map((m) => `
        <div class="history-msg ${m.direction}">
          <span class="history-direction">${m.direction === 'sent' ? '→' : '←'}</span>
          <span class="history-content">${escHtml(m.content)}</span>
          <span class="history-time">${new Date(m.timestamp).toLocaleString()}</span>
        </div>`).join('')}
      </div>
      <button class="btn btn-small btn-secondary btn-export-history" data-pubkey="${pubkey}">Exportar JSON</button>
    </div>`
  })
  container.innerHTML = html
  container.querySelectorAll('.btn-export-history').forEach((btn) => {
    btn.addEventListener('click', () => {
      const pk = (btn as HTMLElement).dataset.pubkey || ''
      const json = exportConversation(pk)
      const blob = new Blob([json], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url; a.download = `keychat-conversation-${pk.slice(0, 8)}.json`; a.click()
      URL.revokeObjectURL(url)
    })
  })
}

function escHtml(s: string): string {
  const d = document.createElement('div')
  d.textContent = s
  return d.innerHTML
}
