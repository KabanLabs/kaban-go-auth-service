package hash

import (
	"testing"
)

func TestGenerateAndCompareHash(t *testing.T) {
	password := "super_secret_password"

	hashStr, err := GenerateFromPassword(password, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if hashStr == "" {
		t.Fatal("expected hash string to not be empty")
	}

	match, err := CompareHashAndPassword(hashStr, password)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !match {
		t.Error("expected passwords to match")
	}

	match, err = CompareHashAndPassword(hashStr, "wrong_password")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if match {
		t.Error("expected passwords to NOT match")
	}
}

func TestInvalidHashFormat(t *testing.T) {
	match, err := CompareHashAndPassword("invalid_hash_format", "password")
	if err != ErrInvalidHash {
		t.Errorf("expected ErrInvalidHash, got %v", err)
	}
	if match {
		t.Error("expected match to be false")
	}

	match, err = CompareHashAndPassword("$argon2i$v=19$m=65536,t=3,p=4$salt$hash", "password")
	if err != ErrInvalidHash {
		t.Errorf("expected ErrInvalidHash, got %v", err)
	}
	if match {
		t.Error("expected match to be false")
	}
}
