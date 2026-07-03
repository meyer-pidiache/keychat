import { secp256k1 } from '@noble/curves/secp256k1'

export function deriveSharedSecret(
  privateKey: Uint8Array,
  publicKey: Uint8Array,
): Uint8Array {
  return secp256k1.getSharedSecret(privateKey, publicKey)
}
