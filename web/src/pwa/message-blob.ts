export interface OutboxBlob {
  version: number
  type: 'nip44'
  sender: string
  recipient: string
  ciphertext: string
  created_at: number
}

export interface ParsedBlob {
  sender: string
  recipient: string
  ciphertext: string
  created_at: number
}

export function createOutboxBlob(
  ciphertext: string,
  senderPubKey: string,
  recipientPubKey: string,
): string {
  const blob: OutboxBlob = {
    version: 1,
    type: 'nip44',
    sender: senderPubKey,
    recipient: recipientPubKey,
    ciphertext,
    created_at: Math.floor(Date.now() / 1000),
  }
  return JSON.stringify(blob)
}

export function parseInboxBlob(json: string): ParsedBlob | null {
  try {
    const blob = JSON.parse(json) as OutboxBlob
    if (!blob.version || !blob.ciphertext || !blob.sender || !blob.recipient) return null
    return {
      sender: blob.sender,
      recipient: blob.recipient,
      ciphertext: blob.ciphertext,
      created_at: blob.created_at || 0,
    }
  } catch {
    return null
  }
}

export async function copyToClipboard(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text)
    return true
  } catch {
    return false
  }
}

export async function readFromClipboard(): Promise<string | null> {
  try {
    return await navigator.clipboard.readText()
  } catch {
    return null
  }
}
