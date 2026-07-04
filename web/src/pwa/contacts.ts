import { validatePubkey } from '../crypto/keys'

const STORAGE_KEY = 'keychat_contacts'

export interface Contact {
  name: string
  pubkey: string
  createdAt: number
}

export function getContacts(): Contact[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return []
    return JSON.parse(raw) as Contact[]
  } catch {
    return []
  }
}

export function saveContacts(contacts: Contact[]): void {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(contacts))
}

export function addContact(name: string, pubkey: string): Contact | null {
  if (!name.trim()) return null
  if (!validatePubkey(pubkey)) return null
  const contacts = getContacts()
  if (contacts.some((c) => c.pubkey === pubkey)) return null
  const contact: Contact = { name: name.trim(), pubkey, createdAt: Date.now() }
  contacts.push(contact)
  saveContacts(contacts)
  return contact
}

export function removeContact(pubkey: string): void {
  const contacts = getContacts().filter((c) => c.pubkey !== pubkey)
  saveContacts(contacts)
}

export function findContact(pubkey: string): Contact | null {
  return getContacts().find((c) => c.pubkey === pubkey) || null
}

export function updateContactName(pubkey: string, name: string): boolean {
  if (!name.trim()) return false
  const contacts = getContacts()
  const idx = contacts.findIndex((c) => c.pubkey === pubkey)
  if (idx === -1) return false
  contacts[idx].name = name.trim()
  saveContacts(contacts)
  return true
}
