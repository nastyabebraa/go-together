package bot

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"ride-together-bot/conf"
)

func TestIsPointInCircle(t *testing.T) {
	location := &Location{conf: &conf.Config{EarthRadius: 6371}}
	if !location.IsPointInCircle(48.8566, 2.3522, 48.8606, 2.3522) {
		t.Error("point at about 445 meters must be inside the circle")
	}
	if location.IsPointInCircle(48.8566, 2.3522, 48.8766, 2.3522) {
		t.Error("point at about 2.2 kilometers must be outside the circle")
	}
}

func TestGetCoordinates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("q") != "Москва, Кремль" {
			t.Errorf("q = %q", r.URL.Query().Get("q"))
		}
		if r.Header.Get("User-Agent") == "" {
			t.Error("User-Agent is empty")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"lat":"55.7520233","lon":"37.6174994"}]`))
	}))
	defer server.Close()

	location := &Location{
		conf:       &conf.Config{OpenStreetMap: server.URL},
		httpClient: server.Client(),
	}
	coordinates, err := location.GetCoordinates(context.Background(), "Москва, Кремль")
	if err != nil {
		t.Fatalf("GetCoordinates() error = %v", err)
	}
	if coordinates.Lat != 55.7520233 || coordinates.Lon != 37.6174994 {
		t.Fatalf("coordinates = %+v", coordinates)
	}
}

func TestGetCoordinatesHandlesEmptyResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	location := &Location{
		conf:       &conf.Config{OpenStreetMap: server.URL},
		httpClient: server.Client(),
	}
	if _, err := location.GetCoordinates(context.Background(), "unknown"); err == nil {
		t.Fatal("GetCoordinates() error = nil")
	}
}
