package relay

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	kcrypto "github.com/opc/keychat/internal/crypto"
)

type Rumor struct {
	Content   string `json:"content"`
	Kind      int    `json:"kind"`
	Tags      []Tag  `json:"tags"`
	CreatedAt int64  `json:"created_at"`
	Pubkey    string `json:"pubkey"`
}

type Seal struct {
	Content   string `json:"content"`
	Kind      int    `json:"kind"`
	Tags      []Tag  `json:"tags"`
	CreatedAt int64  `json:"created_at"`
	Pubkey    string `json:"pubkey"`
}

func CreateGiftWrap(rumorContent string, senderPrivKey, recipientPubKey []byte) (*Event, error) {
	senderPubKey := kcrypto.GetPubkey(senderPrivKey)

	rumor := Rumor{
		Content:   rumorContent,
		Kind:      14,
		Tags:      []Tag{{"p", hex.EncodeToString(recipientPubKey)}},
		CreatedAt: 1700000000,
		Pubkey:    hex.EncodeToString(senderPubKey),
	}
	rumorJSON, err := json.Marshal(rumor)
	if err != nil {
		return nil, fmt.Errorf("marshal rumor: %w", err)
	}

	encryptedRumor, err := kcrypto.Encrypt(string(rumorJSON), senderPrivKey, recipientPubKey)
	if err != nil {
		return nil, fmt.Errorf("encrypt rumor: %w", err)
	}

	seal := Seal{
		Content:   encryptedRumor,
		Kind:      13,
		Tags:      []Tag{{"p", hex.EncodeToString(senderPubKey)}},
		CreatedAt: 1700000001,
		Pubkey:    hex.EncodeToString(senderPubKey),
	}
	sealJSON, err := json.Marshal(seal)
	if err != nil {
		return nil, fmt.Errorf("marshal seal: %w", err)
	}

	ephemeralPriv, _, err := kcrypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("generate ephemeral key: %w", err)
	}

	encryptedSeal, err := kcrypto.Encrypt(string(sealJSON), ephemeralPriv, recipientPubKey)
	if err != nil {
		return nil, fmt.Errorf("encrypt seal: %w", err)
	}

	ephemeralPub := kcrypto.GetPubkey(ephemeralPriv)
	giftWrap := &Event{
		Pubkey:    hex.EncodeToString(ephemeralPub),
		CreatedAt: 1700000002,
		Kind:      1059,
		Tags:      []Tag{{"p", hex.EncodeToString(recipientPubKey)}},
		Content:   encryptedSeal,
	}
	giftWrap.ID = giftWrap.ComputeID()

	return giftWrap, nil
}

func UnwrapGiftWrap(giftWrap *Event, recipientPrivKey []byte) (*Rumor, error) {
	if giftWrap.Kind != 1059 {
		return nil, fmt.Errorf("not a gift wrap: kind %d", giftWrap.Kind)
	}

	ephemeralPub, err := hex.DecodeString(giftWrap.Pubkey)
	if err != nil {
		return nil, fmt.Errorf("decode ephemeral pubkey: %w", err)
	}

	sealJSON, err := kcrypto.Decrypt(giftWrap.Content, recipientPrivKey, ephemeralPub)
	if err != nil {
		return nil, fmt.Errorf("decrypt gift wrap: %w", err)
	}

	var seal Seal
	if err := json.Unmarshal([]byte(sealJSON), &seal); err != nil {
		return nil, fmt.Errorf("parse seal: %w", err)
	}

	senderPubKey, err := hex.DecodeString(seal.Pubkey)
	if err != nil {
		return nil, fmt.Errorf("decode sender pubkey: %w", err)
	}

	rumorJSON, err := kcrypto.Decrypt(seal.Content, recipientPrivKey, senderPubKey)
	if err != nil {
		return nil, fmt.Errorf("decrypt seal: %w", err)
	}

	var rumor Rumor
	if err := json.Unmarshal([]byte(rumorJSON), &rumor); err != nil {
		return nil, fmt.Errorf("parse rumor: %w", err)
	}

	return &rumor, nil
}

func GetGiftWrapRecipients(event *Event) []string {
	if event.Kind != 1059 {
		return nil
	}
	var recipients []string
	for _, tag := range event.Tags {
		if len(tag) > 1 && tag[0] == "p" {
			recipients = append(recipients, tag[1])
		}
	}
	return recipients
}
