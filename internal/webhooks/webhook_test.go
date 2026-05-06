package webhooks

import (
	"context"
	"database/sql"
	"encoding/base64"
	"net/http"
	"strconv"
	"testing"
	"time"

	standardwebhooks "github.com/standard-webhooks/standard-webhooks/libraries/go"
	_ "modernc.org/sqlite"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestStoreRegisterAndList(t *testing.T) {
	db := testDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	wh, err := store.Register(ctx, "https://example.com/hook", "pages/**")
	if err != nil {
		t.Fatal(err)
	}
	if wh.ID == "" || wh.Secret == "" {
		t.Fatal("expected non-empty ID and Secret")
	}

	hooks, err := store.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(hooks) != 1 {
		t.Fatalf("expected 1 webhook, got %d", len(hooks))
	}
	if hooks[0].URL != "https://example.com/hook" {
		t.Fatalf("unexpected URL: %s", hooks[0].URL)
	}
}

func TestStoreDelete(t *testing.T) {
	db := testDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	wh, _ := store.Register(ctx, "https://example.com/hook", "pages/**")
	if err := store.Delete(ctx, wh.ID); err != nil {
		t.Fatal(err)
	}

	hooks, _ := store.List(ctx)
	if len(hooks) != 0 {
		t.Fatalf("expected 0 webhooks after delete, got %d", len(hooks))
	}
}

func TestSignPayloadVerifiesWithStandardWebhooks(t *testing.T) {
	rawKey := []byte("test-secret-key-for-webhooks1234")
	b64Secret := base64.StdEncoding.EncodeToString(rawKey)

	msgID := "msg_abc123"
	now := time.Now().UTC()
	payload := []byte(`{"type":"write","path":"pages/test.md"}`)

	sig, err := signPayload(b64Secret, msgID, now, payload)
	if err != nil {
		t.Fatalf("signPayload: %v", err)
	}

	verifier, err := standardwebhooks.NewWebhookRaw(rawKey)
	if err != nil {
		t.Fatalf("NewWebhookRaw: %v", err)
	}
	headers := http.Header{}
	headers.Set("webhook-id", msgID)
	headers.Set("webhook-timestamp", strconv.FormatInt(now.Unix(), 10))
	headers.Set("webhook-signature", sig)

	if err := verifier.Verify(payload, headers); err != nil {
		t.Fatalf("standard-webhooks verification failed: %v", err)
	}
}

func TestGlobMatch(t *testing.T) {
	tests := []struct {
		pattern, path string
		want          bool
	}{
		{"pages/**", "pages/auth.md", true},
		{"pages/**", "episodes/run-1.md", false},
		{"*.md", "test.md", true},
		{"*.md", "test.txt", false},
	}
	for _, tt := range tests {
		got := matchGlob(tt.pattern, tt.path)
		if got != tt.want {
			t.Errorf("matchGlob(%q, %q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
		}
	}
}
