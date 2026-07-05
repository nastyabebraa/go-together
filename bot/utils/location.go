package bot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	tele "gopkg.in/telebot.v3"

	"ride-together-bot/conf"
	"ride-together-bot/db"
)

type Location struct {
	api        *tele.Bot
	db         *db.DB
	sticker    Sticker
	conf       *conf.Config
	httpClient *http.Client
}

type Coordinates struct {
	Lat float64 `json:"lat,string"`
	Lon float64 `json:"lon,string"`
}

func NewLocation(api *tele.Bot, database *db.DB, sticker Sticker, config *conf.Config) *Location {
	return &Location{
		api:        api,
		db:         database,
		sticker:    sticker,
		conf:       config,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (location *Location) Request(chatID int64) error {
	keyboard := &tele.ReplyMarkup{ResizeKeyboard: true}
	keyboard.Reply(keyboard.Row(tele.Btn{Text: "Отправить геолокацию", Location: true}))
	if _, err := location.api.Send(
		tele.ChatID(chatID),
		"Предоставьте свою геолокацию.",
		&keyboard,
	); err != nil {
		return fmt.Errorf("request geolocation: %w", err)
	}
	return location.sticker.SendSticker(chatID, location.conf.Stickers.Location)
}

func (location *Location) Handle(ctx context.Context, message *tele.Message) (string, error) {
	if message == nil || message.Chat == nil || message.Location == nil {
		return "", errors.New("location message is incomplete")
	}
	if err := location.db.UpsertLocation(
		ctx,
		message.Chat.ID,
		float64(message.Location.Lat),
		float64(message.Location.Lng),
	); err != nil {
		return "", err
	}

	addresses, err := location.db.GetAllDepartureAddresses(ctx)
	if err != nil {
		return "", err
	}

	eventIDs := make(map[int64]struct{})
	geocodedAddresses := 0
	var geocodingError error
	for _, address := range addresses {
		coordinates, err := location.GetCoordinates(ctx, address)
		if err != nil {
			geocodingError = err
			continue
		}
		geocodedAddresses++
		if !location.IsPointInCircle(
			float64(message.Location.Lat),
			float64(message.Location.Lng),
			coordinates.Lat,
			coordinates.Lon,
		) {
			continue
		}
		events, err := location.db.GetAllDataFromEvents(ctx, address)
		if err != nil {
			return "", err
		}
		for _, event := range events {
			eventIDs[event.IDEvent] = struct{}{}
		}
	}
	if len(addresses) > 0 && geocodedAddresses == 0 {
		return "", fmt.Errorf("geocode departure addresses: %w", geocodingError)
	}

	numericIDs := make([]int64, 0, len(eventIDs))
	for id := range eventIDs {
		numericIDs = append(numericIDs, id)
	}
	sort.Slice(numericIDs, func(i, j int) bool {
		return numericIDs[i] < numericIDs[j]
	})
	ids := make([]string, len(numericIDs))
	for i, id := range numericIDs {
		ids[i] = strconv.FormatInt(id, 10)
	}
	pageURL, err := addQuery(location.conf.URLs.Maps, map[string]string{
		"chat_id":   strconv.FormatInt(message.Chat.ID, 10),
		"id_events": strings.Join(ids, ","),
	})
	if err != nil {
		return "", fmt.Errorf("build map URL: %w", err)
	}
	return pageURL, nil
}

func (location *Location) IsPointInCircle(centerLat, centerLon, pointLat, pointLon float64) bool {
	return location.haversine(centerLat, centerLon, pointLat, pointLon) <= 1
}

func (location *Location) haversine(lat1, lon1, lat2, lon2 float64) float64 {
	lat1 = toRadians(lat1)
	lat2 = toRadians(lat2)
	deltaLat := lat2 - lat1
	deltaLon := toRadians(lon2 - lon1)
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Sin(deltaLon/2)*math.Sin(deltaLon/2)*math.Cos(lat1)*math.Cos(lat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return location.conf.EarthRadius * c
}

func (location *Location) GetCoordinates(ctx context.Context, address string) (*Coordinates, error) {
	endpoint, err := url.Parse(location.conf.OpenStreetMap)
	if err != nil {
		return nil, fmt.Errorf("parse geocoding URL: %w", err)
	}
	query := endpoint.Query()
	query.Set("q", address)
	query.Set("format", "json")
	query.Set("limit", "1")
	endpoint.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create geocoding request: %w", err)
	}
	request.Header.Set("User-Agent", "ride-together-bot/1.0")
	response, err := location.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request coordinates: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geocoding returned status %d", response.StatusCode)
	}

	var results []Coordinates
	if err := json.NewDecoder(response.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("decode coordinates: %w", err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("coordinates not found for %q", address)
	}
	return &results[0], nil
}

func toRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}
