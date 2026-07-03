package crypto

import (
	"encoding/hex"
	"testing"
)

func TestNIP44Roundtrip(t *testing.T) {
	privA, _, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	_, pubB, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	plaintext := "Hello Bob! This is a secret message."
	ciphertext, err := Encrypt(plaintext, privA, pubB)
	if err != nil {
		t.Fatalf("encrypt error: %v", err)
	}
	if ciphertext == "" {
		t.Fatal("empty ciphertext")
	}

	decrypted, err := Decrypt(ciphertext, privA, pubB)
	if err != nil {
		t.Fatalf("decrypt error: %v", err)
	}
	if decrypted != plaintext {
		t.Fatalf("roundtrip mismatch: %s != %s", decrypted, plaintext)
	}
}

func TestNIP44RoundtripFromHexKeys(t *testing.T) {
	privA, err := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000001")
	if err != nil {
		t.Fatal(err)
	}
	pubA := "0279be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798"
	pubABytes, _ := hex.DecodeString(pubA)

	plaintext := "test message"
	ciphertext, err := Encrypt(plaintext, privA, pubABytes)
	if err != nil {
		t.Fatalf("encrypt error: %v", err)
	}

	decrypted, err := Decrypt(ciphertext, privA, pubABytes)
	if err != nil {
		t.Fatalf("decrypt error: %v", err)
	}
	if decrypted != plaintext {
		t.Fatalf("roundtrip mismatch: %s != %s", decrypted, plaintext)
	}
}

func TestNIP44WrongKeyFails(t *testing.T) {
	privA, _, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	privB, pubB, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	plaintext := "secret"
	ciphertext, err := Encrypt(plaintext, privA, pubB)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Decrypt(ciphertext, privB, pubB)
	if err == nil {
		t.Fatal("expected error decrypting with wrong key")
	}
}

func TestNIP44TamperedCiphertext(t *testing.T) {
	privA, _, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	_, pubB, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	ciphertext, err := Encrypt("test", privA, pubB)
	if err != nil {
		t.Fatal(err)
	}

	tampered := []byte(ciphertext)
	tampered[len(tampered)-1] ^= 0x01

	_, err = Decrypt(string(tampered), privA, pubB)
	if err == nil {
		t.Fatal("expected error for tampered ciphertext")
	}
}
