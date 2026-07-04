import { secp256k1 } from '@noble/curves/secp256k1'
import { bytesToHex, hexToBytes } from '@noble/hashes/utils'

export interface KeyPair {
  privateKey: Uint8Array
  publicKey: Uint8Array
}

export function generateKeyPair(): KeyPair {
  const privateKey = secp256k1.utils.randomPrivateKey()
  const publicKey = secp256k1.getPublicKey(privateKey, false).subarray(1, 33)
  return { privateKey, publicKey }
}

export function getPublicKey(privateKey: Uint8Array): Uint8Array {
  const point = secp256k1.ProjectivePoint.fromPrivateKey(privateKey)
  return new Uint8Array(point.toRawBytes(false).subarray(1, 33))
}

export function privateKeyToHex(key: Uint8Array): string {
  return bytesToHex(key)
}

export function publicKeyToHex(key: Uint8Array): string {
  return bytesToHex(key)
}

export function hexToPrivateKey(hex: string): Uint8Array {
  return hexToBytes(hex)
}

export function hexToPublicKey(hex: string): Uint8Array {
  const xOnly = hexToBytes(hex)
  return toCompressedPubKey(xOnly)
}

export function toCompressedPubKey(xOnly: Uint8Array): Uint8Array {
  const compressed = new Uint8Array(33)
  compressed[0] = 0x02
  compressed.set(xOnly, 1)
  return compressed
}

export function validatePubkey(pubkey: string): boolean {
  return /^[0-9a-f]{64}$/i.test(pubkey)
}
