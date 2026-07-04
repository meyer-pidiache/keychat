import { generateKeyPair, privateKeyToHex, publicKeyToHex, hexToPrivateKey, getPublicKey } from '../crypto/keys'
import { hexToBytes } from '@noble/hashes/utils'

const STORAGE_KEY_PRIV = 'keychat_privkey'
const STORAGE_KEY_PUB = 'keychat_pubkey'

export interface KeyPair {
  privateKey: Uint8Array
  publicKey: Uint8Array
}

export function hasKeys(): boolean {
  return !!localStorage.getItem(STORAGE_KEY_PRIV)
}

export function generateAndSaveKeyPair(): KeyPair {
  const pair = generateKeyPair()
  localStorage.setItem(STORAGE_KEY_PRIV, privateKeyToHex(pair.privateKey))
  localStorage.setItem(STORAGE_KEY_PUB, publicKeyToHex(pair.publicKey))
  return pair
}

export function loadKeyPair(): KeyPair | null {
  const privHex = localStorage.getItem(STORAGE_KEY_PRIV)
  const pubHex = localStorage.getItem(STORAGE_KEY_PUB)
  if (!privHex || !pubHex) return null
  return {
    privateKey: hexToPrivateKey(privHex),
    publicKey: hexToBytes(pubHex),
  }
}

export function exportPublicKey(): string {
  const pub = localStorage.getItem(STORAGE_KEY_PUB)
  return pub || ''
}

export function exportPrivateKey(): string {
  const priv = localStorage.getItem(STORAGE_KEY_PRIV)
  return priv || ''
}

export function importKeyPair(privKeyHex: string, pubKeyHex: string): boolean {
  try {
    const priv = hexToPrivateKey(privKeyHex)
    const computedPub = getPublicKey(priv)
    const computedHex = publicKeyToHex(computedPub)
    if (pubKeyHex !== computedHex) return false
    localStorage.setItem(STORAGE_KEY_PRIV, privKeyHex)
    localStorage.setItem(STORAGE_KEY_PUB, pubKeyHex)
    return true
  } catch {
    return false
  }
}

export function importPrivateKey(privKeyHex: string): boolean {
  try {
    const priv = hexToPrivateKey(privKeyHex)
    const pub = getPublicKey(priv)
    localStorage.setItem(STORAGE_KEY_PRIV, privKeyHex)
    localStorage.setItem(STORAGE_KEY_PUB, publicKeyToHex(pub))
    return true
  } catch {
    return false
  }
}

export function clearKeys(): void {
  localStorage.removeItem(STORAGE_KEY_PRIV)
  localStorage.removeItem(STORAGE_KEY_PUB)
}
