package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
)

const (
	nip44Version    = 0x02
	nonceSize       = 32
	keyMaterialSize = 76
	keySize         = 32
	ivSize          = 12
)

func sharedSecretECDH(privKey, pubKey []byte) ([]byte, error) {
	priv := secp.PrivKeyFromBytes(privKey)
	pub, err := secp.ParsePubKey(pubKey)
	if err != nil {
		return nil, fmt.Errorf("parse pubkey: %w", err)
	}
	return secp.GenerateSharedSecret(priv, pub), nil
}

func deriveKeyMaterial(sharedSec []byte) []byte {
	salt := []byte("nip44-v2")
	info := []byte("nip44-key-material")
	hkdf := hkdf.New(sha256.New, sharedSec, salt, info)
	keyMaterial := make([]byte, keyMaterialSize)
	io.ReadFull(hkdf, keyMaterial)
	return keyMaterial
}

func Encrypt(plaintext string, privateKey, publicKey []byte) (string, error) {
	sharedSec, err := sharedSecretECDH(privateKey, publicKey)
	if err != nil {
		return "", fmt.Errorf("ecdh: %w", err)
	}

	keyMaterial := deriveKeyMaterial(sharedSec)
	encKey := keyMaterial[:keySize]
	nonce := keyMaterial[keySize : keySize+nonceSize]
	iv := keyMaterial[keySize+nonceSize:]

	aead, err := chacha20poly1305.New(encKey)
	if err != nil {
		return "", fmt.Errorf("new cipher: %w", err)
	}

	ciphertext := aead.Seal(nil, iv, []byte(plaintext), nil)

	mac := computeMAC(encKey, ciphertext, nonce, iv)

	payload := make([]byte, 1+nonceSize+len(ciphertext)+len(mac))
	payload[0] = nip44Version
	copy(payload[1:], nonce)
	copy(payload[1+nonceSize:], ciphertext)
	copy(payload[1+nonceSize+len(ciphertext):], mac)

	return base64.StdEncoding.EncodeToString(payload), nil
}

func Decrypt(ciphertextB64 string, privateKey, publicKey []byte) (string, error) {
	payload, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return "", fmt.Errorf("base64 decode: %w", err)
	}
	if len(payload) < 1 {
		return "", fmt.Errorf("payload too short")
	}
	if payload[0] != nip44Version {
		return "", fmt.Errorf("unsupported version: %d", payload[0])
	}

	minLen := 1 + nonceSize + 1 + 16
	if len(payload) < minLen {
		return "", fmt.Errorf("payload too short for v2")
	}

	nonce := payload[1 : 1+nonceSize]
	ciphertext := payload[1+nonceSize : len(payload)-16]
	mac := payload[len(payload)-16:]

	sharedSec, err := sharedSecretECDH(privateKey, publicKey)
	if err != nil {
		return "", fmt.Errorf("ecdh: %w", err)
	}
	keyMaterial := deriveKeyMaterial(sharedSec)
	encKey := keyMaterial[:keySize]
	iv := keyMaterial[keySize+nonceSize:]

	expectedMAC := computeMAC(encKey, ciphertext, nonce, iv)
	if !hmac.Equal(mac, expectedMAC) {
		return "", fmt.Errorf("MAC mismatch: tampered ciphertext")
	}

	aead, err := chacha20poly1305.New(encKey)
	if err != nil {
		return "", fmt.Errorf("new cipher: %w", err)
	}

	plaintext, err := aead.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}

	return string(plaintext), nil
}

func computeMAC(key, ciphertext, nonce, iv []byte) []byte {
	macInput := make([]byte, 0, len(nonce)+len(ciphertext)+len(iv))
	macInput = append(macInput, nonce...)
	macInput = append(macInput, ciphertext...)
	macInput = append(macInput, iv...)
	mac := hmac.New(sha256.New, key)
	mac.Write(macInput)
	return mac.Sum(nil)[:16]
}

func GenerateKey() (priv, pub []byte, err error) {
	privateKey, err := secp.GeneratePrivateKey()
	if err != nil {
		return nil, nil, err
	}
	return privateKey.Serialize(), privateKey.PubKey().SerializeCompressed(), nil
}

func RandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return b, err
}
