import { encrypt, decrypt } from './nip44'
import { getPublicKey } from './keys'
import { bytesToHex } from '@noble/hashes/utils'
import { secp256k1 } from '@noble/curves/secp256k1'

export interface GiftWrap {
  kind: 1059
  pubkey: string
  content: string
  tags: string[][]
}

export function createGiftWrap(
  rumorContent: string,
  senderPrivKey: Uint8Array,
  recipientPubKey: Uint8Array,
): GiftWrap {
  const senderPubKey = getPublicKey(senderPrivKey)

  const rumor = JSON.stringify({
    content: rumorContent,
    kind: 14,
    tags: [['p', bytesToHex(recipientPubKey)]],
    pubkey: bytesToHex(senderPubKey),
  })

  const encryptedRumor = encrypt(rumor, senderPrivKey, recipientPubKey)

  const seal = JSON.stringify({
    content: encryptedRumor,
    kind: 13,
    tags: [['p', bytesToHex(senderPubKey)]],
    pubkey: bytesToHex(senderPubKey),
  })

  const ephemeralPriv = secp256k1.utils.randomPrivateKey()
  const ephemeralPub = getPublicKey(ephemeralPriv)

  const encryptedSeal = encrypt(seal, ephemeralPriv, recipientPubKey)

  return {
    kind: 1059,
    pubkey: bytesToHex(ephemeralPub),
    content: encryptedSeal,
    tags: [['p', bytesToHex(recipientPubKey)]],
  }
}

export function unwrapGiftWrap(
  giftWrap: GiftWrap,
  recipientPrivKey: Uint8Array,
): string {
  if (giftWrap.kind !== 1059) {
    throw new Error(`not a gift wrap: kind ${giftWrap.kind}`)
  }

  const ephemeralPub = hexToBytes(giftWrap.pubkey)
  const sealJSON = decrypt(giftWrap.content, recipientPrivKey, ephemeralPub)
  const seal = JSON.parse(sealJSON)

  const senderPubKey = hexToBytes(seal.pubkey)
  const rumorJSON = decrypt(seal.content, recipientPrivKey, senderPubKey)
  const rumor = JSON.parse(rumorJSON)

  return rumor.content
}

function hexToBytes(hex: string): Uint8Array {
  const bytes = new Uint8Array(hex.length / 2)
  for (let i = 0; i < bytes.length; i++) {
    bytes[i] = parseInt(hex.substr(i * 2, 2), 16)
  }
  return bytes
}
