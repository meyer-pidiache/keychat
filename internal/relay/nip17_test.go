package relay

import (
	"encoding/hex"
	"testing"

	kcrypto "github.com/opc/keychat/internal/crypto"
)

func TestNIP17Roundtrip(t *testing.T) {
	senderPriv, _, err := kcrypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	recipientPriv, recipientPub, err := kcrypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	content := "Hello Bob! This is a secret message."
	giftWrap, err := CreateGiftWrap(content, senderPriv, recipientPub)
	if err != nil {
		t.Fatalf("create gift wrap: %v", err)
	}

	if giftWrap.Kind != 1059 {
		t.Fatalf("expected kind 1059, got %d", giftWrap.Kind)
	}

	recipientHex := hex.EncodeToString(recipientPub)
	hasPTag := false
	for _, tag := range giftWrap.Tags {
		if len(tag) > 1 && tag[0] == "p" && tag[1] == recipientHex {
			hasPTag = true
		}
	}
	if !hasPTag {
		t.Fatal("gift wrap missing p-tag for recipient")
	}

	rumor, err := UnwrapGiftWrap(giftWrap, recipientPriv)
	if err != nil {
		t.Fatalf("unwrap gift wrap: %v", err)
	}
	if rumor.Content != content {
		t.Fatalf("content mismatch: %s != %s", rumor.Content, content)
	}
}

func TestNIP17WrongKeyFails(t *testing.T) {
	senderPriv, _, err := kcrypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	recipientPriv, recipientPub, err := kcrypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	_, wrongPub, err := kcrypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	giftWrap, err := CreateGiftWrap("secret", senderPriv, recipientPub)
	if err != nil {
		t.Fatal(err)
	}

	_, err = UnwrapGiftWrap(giftWrap, wrongPub)
	if err == nil {
		t.Fatal("expected error decrypting with wrong key")
	}
	_ = recipientPriv
}

func TestNIP17GetRecipients(t *testing.T) {
	e := &Event{
		Kind: 1059,
		Tags: []Tag{{"p", "recipient1", ""}, {"p", "recipient2", ""}},
	}
	recips := GetGiftWrapRecipients(e)
	if len(recips) != 2 {
		t.Fatalf("expected 2 recipients, got %d", len(recips))
	}
	if recips[0] != "recipient1" {
		t.Fatalf("wrong recipient: %s", recips[0])
	}

	e2 := &Event{Kind: 1}
	if r := GetGiftWrapRecipients(e2); r != nil {
		t.Fatal("expected nil for non-gift-wrap")
	}
}
