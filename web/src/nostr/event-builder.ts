import type { NostrEvent } from './relay-client'
import { hexToBytes, bytesToHex } from '@noble/hashes/utils'

export function createEvent(pubkey: string, kind: number, tags: string[][], content: string): NostrEvent {
  return {
    id: '',
    pubkey,
    created_at: Math.floor(Date.now() / 1000),
    kind,
    tags,
    content,
    sig: '',
  }
}

export function serializeEvent(event: NostrEvent): string {
  return JSON.stringify([0, event.pubkey, event.created_at, event.kind, event.tags, event.content])
}

export async function computeEventId(event: NostrEvent): Promise<string> {
  const serialized = serializeEvent(event)
  const hash = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(serialized))
  return Array.from(new Uint8Array(hash))
    .map((b) => b.toString(16).padStart(2, '0'))
    .join('')
}

export async function signEvent(event: NostrEvent, privateKeyHex: string): Promise<NostrEvent> {
  const id = await computeEventId(event)
  event.id = id

  const keyBytes = hexToBytes(privateKeyHex)
  const idBytes = hexToBytes(id)

  const sigBytes = await schnorrSign(keyBytes, idBytes)
  event.sig = bytesToHex(sigBytes)
  return event
}

async function schnorrSign(privateKey: Uint8Array, message: Uint8Array): Promise<Uint8Array> {
  const { schnorr } = await import('@noble/curves/secp256k1')
  return schnorr.sign(message, privateKey)
}
