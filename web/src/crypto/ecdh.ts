import { secp256k1 } from '@noble/curves/secp256k1'

export function deriveSharedSecret(
  privateKey: Uint8Array,
  publicKey: Uint8Array,
): Uint8Array {
  const pk = publicKey.length === 32 ? toCompressedPubKey(publicKey) : publicKey
  return secp256k1.getSharedSecret(privateKey, pk)
}

function toCompressedPubKey(xOnly: Uint8Array): Uint8Array {
  const compressed = new Uint8Array(33)
  compressed[0] = 0x02
  compressed.set(xOnly, 1)
  return compressed
}
