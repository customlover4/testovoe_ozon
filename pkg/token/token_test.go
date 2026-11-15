package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAndParse(t *testing.T) {
	secret := []byte("test-secret")

	raw, err := GenerateToken("42", secret, time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	id, err := Parse(raw, secret)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if id != "42" {
		t.Fatalf("user id = %q", id)
	}
}

func TestParseWrongSecret(t *testing.T) {
	raw, err := GenerateToken("7", []byte("one"), time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	if _, err := Parse(raw, []byte("two")); err == nil {
		t.Fatal("expected error for a different secret")
	}
}

func TestParseExpired(t *testing.T) {
	raw, err := GenerateToken("7", []byte("secret"), -time.Minute)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	if _, err := Parse(raw, []byte("secret")); err == nil {
		t.Fatal("expected error for an expired token")
	}
}

func TestParseMalformed(t *testing.T) {
	tests := []string{
		"",
		"not-a-token",
		"a.b.c",
	}
	for _, raw := range tests {
		t.Run(raw, func(t *testing.T) {
			if _, err := Parse(raw, []byte("secret")); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestParseUnexpectedSigningMethod(t *testing.T) {
	claims := jwt.RegisteredClaims{Subject: "1"}
	raw, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).
		SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	if _, err := Parse(raw, []byte("secret")); err == nil {
		t.Fatal("expected error for a non-HMAC token")
	}
}
