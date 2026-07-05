package bot

import (
	"net/url"
	"testing"
)

func TestAddQuery(t *testing.T) {
	result, err := addQuery("https://example.com/map?lang=ru", map[string]string{
		"chat_id":   "123",
		"id_events": "1,2,3",
	})
	if err != nil {
		t.Fatalf("addQuery() error = %v", err)
	}
	parsed, err := url.Parse(result)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	if parsed.Query().Get("lang") != "ru" {
		t.Errorf("lang = %q, want ru", parsed.Query().Get("lang"))
	}
	if parsed.Query().Get("chat_id") != "123" {
		t.Errorf("chat_id = %q, want 123", parsed.Query().Get("chat_id"))
	}
	if parsed.Query().Get("id_events") != "1,2,3" {
		t.Errorf("id_events = %q, want 1,2,3", parsed.Query().Get("id_events"))
	}
}
