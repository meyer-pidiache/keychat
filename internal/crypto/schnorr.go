package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/schnorr"
)

func VerifyEventSignature(id, pubkeyHex, sigHex string) bool {
	pubBytes, err := hex.DecodeString(pubkeyHex)
	if err != nil || len(pubBytes) != 32 {
		return false
	}
	sigBytes, err := hex.DecodeString(sigHex)
	if err != nil || len(sigBytes) != 64 {
		return false
	}
	idBytes, err := hex.DecodeString(id)
	if err != nil || len(idBytes) != 32 {
		return false
	}

	sig, err := schnorr.ParseSignature(sigBytes)
	if err != nil {
		return false
	}

	compressed := make([]byte, 33)
	copy(compressed[1:], pubBytes[:32])
	compressed[0] = 0x02
	if pk, err := schnorr.ParsePubKey(compressed); err == nil && sig.Verify(idBytes, pk) {
		return true
	}
	compressed[0] = 0x03
	if pk, err := schnorr.ParsePubKey(compressed); err == nil && sig.Verify(idBytes, pk) {
		return true
	}

	return false
}

func ComputeEventID(serialized []byte) string {
	hash := sha256.Sum256(serialized)
	return hex.EncodeToString(hash[:])
}

func ValidatePubkeyHex(pubkeyHex string) error {
	if len(pubkeyHex) != 64 {
		return fmt.Errorf("pubkey must be 64 hex chars")
	}
	_, err := hex.DecodeString(pubkeyHex)
	return err
}

func ValidateSigHex(sigHex string) error {
	if len(sigHex) != 128 {
		return fmt.Errorf("sig must be 128 hex chars")
	}
	_, err := hex.DecodeString(sigHex)
	return err
}

func GenerateKeyPair() (privKeyHex, pubKeyHex string, err error) {
	priv, err := secp.GeneratePrivateKey()
	if err != nil {
		return "", "", err
	}
	pub := priv.PubKey()
	xOnly := pub.SerializeUncompressed()[1:33]
	return hex.EncodeToString(priv.Serialize()), hex.EncodeToString(xOnly), nil
}
