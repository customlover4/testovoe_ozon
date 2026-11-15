package password

import (
	"strings"
	"testing"
)

func TestGenerateAndCompare(t *testing.T) {
	first, err := Generate("correct horse")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	second, err := Generate("correct horse")
	if err != nil {
		t.Fatalf("Generate second: %v", err)
	}
	if first == second {
		t.Fatal("two hashes of the same password are identical")
	}
	if !strings.HasPrefix(first, "$argon2id$") {
		t.Fatalf("hash prefix = %q", first)
	}

	ok, err := Compare("correct horse", first)
	if err != nil {
		t.Fatalf("Compare: %v", err)
	}
	if !ok {
		t.Fatal("expected the original password to match")
	}

	ok, err = Compare("wrong", first)
	if err != nil {
		t.Fatalf("Compare wrong: %v", err)
	}
	if ok {
		t.Fatal("expected a different password to miss")
	}
}

func TestCompareInvalid(t *testing.T) {
	tests := []struct {
		name    string
		encoded string
	}{
		{name: "empty", encoded: ""},
		{name: "wrong algorithm", encoded: "$bcrypt$v=1$m=1,t=1,p=1$aa$bb"},
		{name: "bad params", encoded: "$argon2id$v=19$not-params$YQ$YQ"},
		{name: "bad salt", encoded: "$argon2id$v=19$m=8,t=1,p=1$%%%$YQ"},
		{name: "bad hash", encoded: "$argon2id$v=19$m=8,t=1,p=1$YQ$%%%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := Compare("password", tt.encoded)
			if err == nil {
				t.Fatal("expected error")
			}
			if ok {
				t.Fatal("expected match to be false")
			}
		})
	}
}
