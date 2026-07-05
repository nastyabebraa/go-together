package bot

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v3"

	"ride-together-bot/conf"
	"ride-together-bot/db"
)

type Event struct {
	conf    *conf.Config
	api     *tele.Bot
	db      *db.DB
	sticker Sticker
}

func NewEvent(config *conf.Config, api *tele.Bot, database *db.DB, sticker Sticker) *Event {
	return &Event{conf: config, api: api, db: database, sticker: sticker}
}

func (event *Event) Create(chatID int64) error {
	pageURL, err := addQuery(event.conf.URLs.CreateEventPage, map[string]string{
		"chatID": strconv.FormatInt(chatID, 10),
	})
	if err != nil {
		return fmt.Errorf("build create event URL: %w", err)
	}
	keyboard := &tele.ReplyMarkup{
		ReplyKeyboard: [][]tele.ReplyButton{{{
			Text:   "Создать поездку",
			WebApp: &tele.WebApp{URL: pageURL},
		}}},
		ResizeKeyboard: true,
	}
	if _, err := event.api.Send(tele.ChatID(chatID), "Откройте форму создания поездки.", keyboard); err != nil {
		return fmt.Errorf("send create event button: %w", err)
	}
	return event.sticker.SendSticker(chatID, event.conf.Stickers.CreateEvent)
}

func (event *Event) TripsManagement(ctx context.Context, message *tele.Message) error {
	if message == nil || message.Chat == nil {
		return errors.New("trip management message is incomplete")
	}
	chatID := message.Chat.ID
	isDriver, err := event.db.IsDriver(ctx, chatID)
	if err != nil {
		return err
	}
	pageURL, err := addQuery(event.conf.URLs.DriverEvents, map[string]string{
		"chatID":   strconv.FormatInt(chatID, 10),
		"isDriver": strconv.FormatBool(isDriver),
	})
	if err != nil {
		return fmt.Errorf("build trips management URL: %w", err)
	}
	keyboard := &tele.ReplyMarkup{ResizeKeyboard: true}
	keyboard.Reply(keyboard.Row(tele.Btn{Text: "Менеджер поездок", WebApp: &tele.WebApp{URL: pageURL}}))
	if _, err := event.api.Send(
		tele.ChatID(chatID),
		"Используйте кнопку ниже для управления поездками.",
		&tele.SendOptions{ReplyMarkup: keyboard},
	); err != nil {
		return fmt.Errorf("send trips management button: %w", err)
	}
	return event.sticker.SendSticker(chatID, event.conf.Stickers.Cat)
}

func (event *Event) History(ctx context.Context, chatID int64) (string, error) {
	ids, err := event.db.GetEventIDs(ctx, chatID)
	if err != nil {
		return "", err
	}
	stringIDs := make([]string, len(ids))
	for i, id := range ids {
		stringIDs[i] = strconv.FormatInt(id, 10)
	}
	pageURL, err := addQuery(event.conf.URLs.EventsHistory, map[string]string{
		"chat_id":   strconv.FormatInt(chatID, 10),
		"id_events": strings.Join(stringIDs, ","),
	})
	if err != nil {
		return "", fmt.Errorf("build history URL: %w", err)
	}
	return pageURL, nil
}

func addQuery(rawURL string, values map[string]string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	query := parsed.Query()
	for key, value := range values {
		query.Set(key, value)
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}
