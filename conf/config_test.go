package conf

import (
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_API_KEY", "token")
	t.Setenv("MYSQL_DSN", "user:password@tcp(localhost:3306)/rides")
	t.Setenv("CREATE_EVENT_URL", "https://example.com/events/new")
	t.Setenv("DRIVER_EVENTS_URL", "https://example.com/events/manage")
	t.Setenv("EVENTS_HISTORY_URL", "https://example.com/events/history")
	t.Setenv("MAPS_URL", "https://example.com/maps")
	t.Setenv("EARTH_RADIUS_KM", "6372.8")

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if config.EarthRadius != 6372.8 {
		t.Fatalf("EarthRadius = %v, want 6372.8", config.EarthRadius)
	}
	if config.OpenStreetMap != defaultNominatimURL {
		t.Fatalf("OpenStreetMap = %q, want %q", config.OpenStreetMap, defaultNominatimURL)
	}
}

func TestLoadReportsMissingValues(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_API_KEY", "")
	t.Setenv("MYSQL_DSN", "")
	t.Setenv("CREATE_EVENT_URL", "")
	t.Setenv("DRIVER_EVENTS_URL", "")
	t.Setenv("EVENTS_HISTORY_URL", "")
	t.Setenv("MAPS_URL", "")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil")
	}
	for _, name := range []string{
		"TELEGRAM_BOT_API_KEY",
		"MYSQL_DSN",
		"CREATE_EVENT_URL",
		"DRIVER_EVENTS_URL",
		"EVENTS_HISTORY_URL",
		"MAPS_URL",
	} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("Load() error %q does not contain %q", err, name)
		}
	}
}

func TestLoadRejectsInvalidRadius(t *testing.T) {
	t.Setenv("EARTH_RADIUS_KM", "0")
	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "EARTH_RADIUS_KM") {
		t.Fatalf("Load() error = %v", err)
	}
}
