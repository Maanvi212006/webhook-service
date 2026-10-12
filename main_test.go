package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func sign(body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestValidSignature(t *testing.T) {
	body := []byte(`{"id":"evt_1","action":"pushed","user":"rahul"}`)
	if !isValid(body, sign(body)) {
		t.Error("expected a correct signature to be accepted")
	}
}

func TestWrongSignature(t *testing.T) {
	body := []byte(`{"id":"evt_1","action":"pushed","user":"rahul"}`)
	if isValid(body, "sha256=wrong") {
		t.Error("expected a wrong signature to be rejected")
	}
}

func TestTamperedBody(t *testing.T) {
	original := []byte(`{"id":"evt_1","action":"pushed","user":"rahul"}`)
	tampered := []byte(`{"id":"evt_1","action":"pushed","user":"rohan"}`)
	if isValid(tampered, sign(original)) {
		t.Error("expected a changed message to be rejected")
	}
}
