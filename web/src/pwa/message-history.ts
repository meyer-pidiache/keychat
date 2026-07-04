const STORAGE_KEY = 'keychat_messages'

export interface Message {
  id: string
  contactPubKey: string
  content: string
  direction: 'sent' | 'received'
  timestamp: number
}

export function saveMessage(msg: Omit<Message, 'id'>): Message {
  const messages = getAllMessages()
  const full: Message = { ...msg, id: crypto.randomUUID() }
  messages.push(full)
  localStorage.setItem(STORAGE_KEY, JSON.stringify(messages))
  return full
}

export function getAllMessages(): Message[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return []
    return JSON.parse(raw) as Message[]
  } catch {
    return []
  }
}

export function getConversation(contactPubKey: string): Message[] {
  return getAllMessages()
    .filter((m) => m.contactPubKey === contactPubKey)
    .sort((a, b) => a.timestamp - b.timestamp)
}

export function getAllConversations(): Map<string, Message[]> {
  const map = new Map<string, Message[]>()
  for (const msg of getAllMessages().sort((a, b) => b.timestamp - a.timestamp)) {
    const existing = map.get(msg.contactPubKey)
    if (existing) {
      existing.push(msg)
    } else {
      map.set(msg.contactPubKey, [msg])
    }
  }
  map.forEach((msgs) => msgs.sort((a, b) => a.timestamp - b.timestamp))
  return map
}

export function exportConversation(contactPubKey: string): string {
  const msgs = getConversation(contactPubKey)
  return JSON.stringify({ contact: contactPubKey, messages: msgs, exported_at: Date.now() }, null, 2)
}
