import { secp256k1 } from '@noble/curves/secp256k1'
import { sha256 } from '@noble/hashes/sha256'
import { hmac } from '@noble/hashes/hmac'
import { hkdf } from '@noble/hashes/hkdf'
import { chacha20poly1305 } from '@noble/ciphers/chacha'
import { bytesToHex, hexToBytes } from '@noble/hashes/utils'
import { toCompressedPubKey } from './keys'

const NIP44_VERSION = 0x02
const NONCE_SIZE = 32
const KEY_SIZE = 32
const IV_SIZE = 12
const KEY_MATERIAL_SIZE = KEY_SIZE + NONCE_SIZE + IV_SIZE

export function deriveSharedSecret(
  privateKey: Uint8Array,
  publicKey: Uint8Array,
): Uint8Array {
  const pk = publicKey.length === 32 ? toCompressedPubKey(publicKey) : publicKey
  const ss = secp256k1.getSharedSecret(privateKey, pk)
  return ss.subarray(1, 33)
}

function deriveKeyMaterial(sharedSecret: Uint8Array): Uint8Array {
  const salt = new TextEncoder().encode('nip44-v2')
  const info = new TextEncoder().encode('nip44-key-material')
  return hkdf(sha256, sharedSecret, salt, info, KEY_MATERIAL_SIZE)
}

export function encrypt(
  plaintext: string,
  privateKey: Uint8Array,
  publicKey: Uint8Array,
): string {
  const sharedSecret = deriveSharedSecret(privateKey, publicKey)
  const keyMaterial = deriveKeyMaterial(sharedSecret)

  const encKey = keyMaterial.subarray(0, KEY_SIZE)
  const nonce = keyMaterial.subarray(KEY_SIZE, KEY_SIZE + NONCE_SIZE)
  const iv = keyMaterial.subarray(KEY_SIZE + NONCE_SIZE)

  const cipher = chacha20poly1305(encKey, iv)
  const ciphertext = cipher.encrypt(new TextEncoder().encode(plaintext))

  const mac = computeMAC(encKey, ciphertext, nonce, iv)

  const payload = new Uint8Array(1 + nonce.length + ciphertext.length + mac.length)
  payload[0] = NIP44_VERSION
  payload.set(nonce, 1)
  payload.set(ciphertext, 1 + nonce.length)
  payload.set(mac, 1 + nonce.length + ciphertext.length)

  return bytesToHex(payload)
}

export function decrypt(
  ciphertextHex: string,
  privateKey: Uint8Array,
  publicKey: Uint8Array,
): string {
  const payload = hexToBytes(ciphertextHex)
  if (payload.length < 1) throw new Error('payload too short')
  if (payload[0] !== NIP44_VERSION) throw new Error(`unsupported version: ${payload[0]}`)

  const nonce = payload.subarray(1, 1 + NONCE_SIZE)
  const ciphertext = payload.subarray(1 + NONCE_SIZE, payload.length - 16)
  const mac = payload.subarray(payload.length - 16)

  const sharedSecret = deriveSharedSecret(privateKey, publicKey)
  const keyMaterial = deriveKeyMaterial(sharedSecret)

  const encKey = keyMaterial.subarray(0, KEY_SIZE)
  const iv = keyMaterial.subarray(KEY_SIZE + NONCE_SIZE)

  const expectedMAC = computeMAC(encKey, ciphertext, nonce, iv)
  if (!constantTimeEqual(mac, expectedMAC)) {
    throw new Error('MAC mismatch: tampered ciphertext')
  }

  const cipher = chacha20poly1305(encKey, iv)
  const plaintext = cipher.decrypt(ciphertext)
  return new TextDecoder().decode(plaintext)
}

function computeMAC(
  key: Uint8Array,
  ciphertext: Uint8Array,
  nonce: Uint8Array,
  iv: Uint8Array,
): Uint8Array {
  const input = new Uint8Array(nonce.length + ciphertext.length + iv.length)
  input.set(nonce, 0)
  input.set(ciphertext, nonce.length)
  input.set(iv, nonce.length + ciphertext.length)
  return hmac(sha256, key, input).subarray(0, 16)
}

function constantTimeEqual(a: Uint8Array, b: Uint8Array): boolean {
  if (a.length !== b.length) return false
  let diff = 0
  for (let i = 0; i < a.length; i++) {
    diff |= a[i] ^ b[i]
  }
  return diff === 0
}
