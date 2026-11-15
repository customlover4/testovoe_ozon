package cfg

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMustLoad(t *testing.T) {
	unsetConfigEnv(t)

	path := writeConfig(t, `
db_conn: "postgres://localhost/app"
handler:
  debug: true
web:
  addr: "0.0.0.0"
  port: "8081"
  read_header_timeout: 3s
  read_timeout: 10s
  write_timeout: 15s
logger:
  level: "info"
auth:
  secret: "file-secret"
`)

	got := MustLoad(path)

	if got.ConnString != "postgres://localhost/app" {
		t.Fatalf("conn = %q", got.ConnString)
	}
	if !got.Debug {
		t.Fatal("expected debug")
	}
	if got.Addr != "0.0.0.0" || got.Port != "8081" {
		t.Fatalf("addr = %s:%s", got.Addr, got.Port)
	}
	if got.ReadHeaderTimeout != 3*time.Second || got.ReadTimeout != 10*time.Second || got.WriteTimeout != 15*time.Second {
		t.Fatalf("timeouts = %s %s %s", got.ReadHeaderTimeout, got.ReadTimeout, got.WriteTimeout)
	}
	if got.Level != "info" {
		t.Fatalf("level = %q", got.Level)
	}
	if got.Secret != "file-secret" {
		t.Fatalf("secret = %q", got.Secret)
	}
}

func TestMustLoadDefaults(t *testing.T) {
	unsetConfigEnv(t)

	path := writeConfig(t, `
auth:
  secret: "only-secret"
`)

	got := MustLoad(path)

	if got.Addr != "127.0.0.1" {
		t.Fatalf("addr = %q", got.Addr)
	}
	if got.Port != "8080" {
		t.Fatalf("port = %q", got.Port)
	}
	if got.ReadTimeout != 10*time.Second || got.ReadHeaderTimeout != 3*time.Second || got.WriteTimeout != 10*time.Second {
		t.Fatalf("timeouts = %s %s %s", got.ReadTimeout, got.ReadHeaderTimeout, got.WriteTimeout)
	}
	if got.Debug {
		t.Fatal("expected debug to stay false")
	}
	if got.Secret != "only-secret" {
		t.Fatalf("secret = %q", got.Secret)
	}
}

func TestMustLoadEnvOverridesFile(t *testing.T) {
	unsetConfigEnv(t)
	t.Setenv("PORT", "9090")
	t.Setenv("SECRET", "env-secret")
	t.Setenv("ADDR", "10.0.0.1")

	path := writeConfig(t, `
web:
  addr: "127.0.0.1"
  port: "8081"
auth:
  secret: "file-secret"
`)

	got := MustLoad(path)

	if got.Port != "9090" || got.Addr != "10.0.0.1" || got.Secret != "env-secret" {
		t.Fatalf("addr=%s port=%s secret=%s", got.Addr, got.Port, got.Secret)
	}
}

func TestMustLoadPanics(t *testing.T) {
	unsetConfigEnv(t)

	t.Run("missing file", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		MustLoad(filepath.Join(t.TempDir(), "missing.yml"))
	})

	t.Run("missing secret", func(t *testing.T) {
		path := writeConfig(t, "logger:\n  level: debug\n")
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		MustLoad(path)
	})
}

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

func unsetConfigEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"ADDR", "PORT", "READ_TIMEOUT", "READ_HEADER_TIMEOUT", "WRITE_TIMEOUT", "SECRET",
	} {
		t.Setenv(key, "placeholder")
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
	}
}
