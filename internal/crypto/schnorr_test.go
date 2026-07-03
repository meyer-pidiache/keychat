package crypto

import (
	"encoding/hex"
	"testing"

	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/schnorr"
)

func TestVerifyValidSignature(t *testing.T) {
	priv, err := secp.GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	pub := priv.PubKey()

	msg := []byte("hello world")
	hash := make([]byte, 32)
	copy(hash, msg)
	sig, err := schnorr.Sign(priv, hash)
	if err != nil {
		t.Fatal(err)
	}

	idHex := hex.EncodeToString(hash)
	xOnly := pub.SerializeUncompressed()[1:33]
	pubHex := hex.EncodeToString(xOnly)
	sigHex := hex.EncodeToString(sig.Serialize())

	if !VerifyEventSignature(idHex, pubHex, sigHex) {
		t.Fatal("expected valid sig to verify")
	}
}

func TestVerifyInvalidSignature(t *testing.T) {
	priv, err := secp.GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	pub := priv.PubKey()

	msg := []byte("hello world")
	hash := make([]byte, 32)
	copy(hash, msg)
	sig, err := schnorr.Sign(priv, hash)
	if err != nil {
		t.Fatal(err)
	}

	pubHex := hex.EncodeToString(pub.SerializeUncompressed()[1:33])
	sigHex := hex.EncodeToString(sig.Serialize())

	tamperedID := "0000000000000000000000000000000000000000000000000000000000000000"
	if VerifyEventSignature(tamperedID, pubHex, sigHex) {
		t.Fatal("expected invalid sig to fail verification")
	}
}

func TestVerifyMalformedPubkey(t *testing.T) {
	if VerifyEventSignature("abcd", "short", "also_short") {
		t.Fatal("expected malformed pubkey to fail")
	}
}

func TestValidatePubkeyHex(t *testing.T) {
	if err := ValidatePubkeyHex("abc"); err == nil {
		t.Fatal("expected error for short pubkey")
	}
	valid := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	if err := ValidatePubkeyHex(valid); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateKeyPair(t *testing.T) {
	priv, pub, err := GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	if len(priv) != 64 {
		t.Fatal("private key should be 64 hex chars")
	}
	if len(pub) != 64 {
		t.Fatal("public key should be 64 hex chars (x-only)")
	}
}
