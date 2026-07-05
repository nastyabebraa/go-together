package conf

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	defaultEarthRadius  = 6371.0
	defaultNominatimURL = "https://nominatim.openstreetmap.org/search"
)

type URLs struct {
	CreateEventPage string
	DriverEvents    string
	EventsHistory   string
	Maps            string
}

type Stickers struct {
	Start       string
	Shrek       string
	Cat         string
	CreateEvent string
	Location    string
}

type Config struct {
	TelegramBotAPIKey string
	DSN               string
	OpenStreetMap     string
	EarthRadius       float64
	URLs              URLs
	Stickers          Stickers
}

func Load() (*Config, error) {
	earthRadius, err := readPositiveFloat("EARTH_RADIUS_KM", defaultEarthRadius)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		TelegramBotAPIKey: strings.TrimSpace(os.Getenv("TELEGRAM_BOT_API_KEY")),
		DSN:               strings.TrimSpace(os.Getenv("MYSQL_DSN")),
		OpenStreetMap:     valueOrDefault("NOMINATIM_URL", defaultNominatimURL),
		EarthRadius:       earthRadius,
		URLs: URLs{
			CreateEventPage: strings.TrimSpace(os.Getenv("CREATE_EVENT_URL")),
			DriverEvents:    strings.TrimSpace(os.Getenv("DRIVER_EVENTS_URL")),
			EventsHistory:   strings.TrimSpace(os.Getenv("EVENTS_HISTORY_URL")),
			Maps:            strings.TrimSpace(os.Getenv("MAPS_URL")),
		},
		Stickers: Stickers{
			Start:       strings.TrimSpace(os.Getenv("STICKER_START")),
			Shrek:       strings.TrimSpace(os.Getenv("STICKER_SHREK")),
			Cat:         strings.TrimSpace(os.Getenv("STICKER_CAT")),
			CreateEvent: strings.TrimSpace(os.Getenv("STICKER_CREATE_EVENT")),
			Location:    strings.TrimSpace(os.Getenv("STICKER_LOCATION")),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	var validationErrors []error
	required := []struct {
		name  string
		value string
	}{
		{name: "TELEGRAM_BOT_API_KEY", value: c.TelegramBotAPIKey},
		{name: "MYSQL_DSN", value: c.DSN},
		{name: "CREATE_EVENT_URL", value: c.URLs.CreateEventPage},
		{name: "DRIVER_EVENTS_URL", value: c.URLs.DriverEvents},
		{name: "EVENTS_HISTORY_URL", value: c.URLs.EventsHistory},
		{name: "MAPS_URL", value: c.URLs.Maps},
	}
	for _, item := range required {
		if strings.TrimSpace(item.value) == "" {
			validationErrors = append(validationErrors, fmt.Errorf("%s is required", item.name))
		}
	}

	endpoints := []struct {
		name  string
		value string
	}{
		{name: "CREATE_EVENT_URL", value: c.URLs.CreateEventPage},
		{name: "DRIVER_EVENTS_URL", value: c.URLs.DriverEvents},
		{name: "EVENTS_HISTORY_URL", value: c.URLs.EventsHistory},
		{name: "MAPS_URL", value: c.URLs.Maps},
		{name: "NOMINATIM_URL", value: c.OpenStreetMap},
	}
	for _, item := range endpoints {
		if item.value == "" {
			continue
		}
		parsed, err := url.ParseRequestURI(item.value)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			validationErrors = append(validationErrors, fmt.Errorf("%s must be an absolute URL", item.name))
		}
	}

	if c.EarthRadius <= 0 {
		validationErrors = append(validationErrors, errors.New("EARTH_RADIUS_KM must be positive"))
	}
	return errors.Join(validationErrors...)
}

func valueOrDefault(name, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

func readPositiveFloat(name string, fallback float64) (float64, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive number", name)
	}
	return parsed, nil
}
