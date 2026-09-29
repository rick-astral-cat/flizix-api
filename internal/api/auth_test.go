package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

func newTestAuth() *AuthHandler {
	return &AuthHandler{
		JWTSecret:        testJWTSecret,
		TelegramBotToken: "test-bot-token",
	}
}

func TestGenerateAndValidateToken(t *testing.T) {
	auth := newTestAuth()

	token, err := auth.GenerateToken(42)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken error: %v", err)
	}

	if claims.UserID != 42 {
		t.Errorf("expected userID 42, got %d", claims.UserID)
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	auth := newTestAuth()
	token, _ := auth.GenerateToken(1)

	other := &AuthHandler{JWTSecret: "wrong-secret"}
	_, err := other.ValidateToken(token)
	if err == nil {
		t.Error("expected error with wrong secret, got nil")
	}
}

func TestValidateToken_MalformedToken(t *testing.T) {
	auth := newTestAuth()
	_, err := auth.ValidateToken("not.a.valid.token")
	if err == nil {
		t.Error("expected error for malformed token, got nil")
	}
}

func TestVerifyTelegramHash_Valid(t *testing.T) {
	botToken := "test-bot-token"
	auth := &AuthHandler{TelegramBotToken: botToken}

	req := TelegramAuthRequest{
		ID:        123456,
		FirstName: "Rick",
		AuthDate:  time.Now().Unix(),
	}

	// Build the correct hash matching the implementation's sort order
	data := fmt.Sprintf("auth_date=%d\nfirst_name=%s\nid=%d", req.AuthDate, req.FirstName, req.ID)
	sha := sha256.New()
	sha.Write([]byte(botToken))
	secretKey := sha.Sum(nil)
	hm := hmac.New(sha256.New, secretKey)
	hm.Write([]byte(data))
	req.Hash = hex.EncodeToString(hm.Sum(nil))

	if err := auth.VerifyTelegramHash(req); err != nil {
		t.Errorf("expected valid hash, got error: %v", err)
	}
}

func TestVerifyTelegramHash_Invalid(t *testing.T) {
	auth := &AuthHandler{TelegramBotToken: "test-bot-token"}
	req := TelegramAuthRequest{
		ID:        123456,
		FirstName: "Rick",
		AuthDate:  time.Now().Unix(),
		Hash:      "invalidhash",
	}

	if err := auth.VerifyTelegramHash(req); err == nil {
		t.Error("expected error for invalid hash, got nil")
	}
}
