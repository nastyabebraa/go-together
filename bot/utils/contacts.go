package bot

import (
	"context"
	"errors"
	"fmt"

	tele "gopkg.in/telebot.v3"

	"ride-together-bot/db"
	"ride-together-bot/entity"
)

type Contact struct {
	api       *tele.Bot
	db        *db.DB
	sticker   Sticker
	stickerID string
}

func NewContact(api *tele.Bot, database *db.DB, sticker Sticker, stickerID string) *Contact {
	return &Contact{api: api, db: database, sticker: sticker, stickerID: stickerID}
}

func (contact *Contact) Handle(ctx context.Context, message *tele.Message) error {
	if message == nil || message.Sender == nil || message.Chat == nil || message.Contact == nil {
		return errors.New("contact message is incomplete")
	}
	if message.Contact.UserID != message.Sender.ID {
		return contact.send(message.Sender, "Отправьте собственный контакт, используя кнопку ниже.")
	}
	name := message.Sender.FirstName
	if name == "" {
		name = message.Sender.Username
	}
	if name == "" {
		name = "Пользователь"
	}
	user := entity.User{
		Name:   name,
		Login:  message.Sender.Username,
		Phone:  message.Contact.PhoneNumber,
		ChatID: message.Chat.ID,
	}
	if err := contact.db.RegisterUser(ctx, user); err != nil {
		return fmt.Errorf("register contact: %w", err)
	}
	if _, err := contact.api.Send(message.Sender, "Спасибо!", &tele.ReplyMarkup{RemoveKeyboard: true}); err != nil {
		return fmt.Errorf("send registration confirmation: %w", err)
	}
	if err := contact.sticker.SendSticker(message.Chat.ID, contact.stickerID); err != nil {
		return err
	}
	return nil
}

func (contact *Contact) Request(chatID int64) error {
	keyboard := &tele.ReplyMarkup{
		ReplyKeyboard: [][]tele.ReplyButton{{
			{Text: "Предоставить контакт", Contact: true},
			{Text: "Отмена"},
		}},
		ResizeKeyboard: true,
	}
	if _, err := contact.api.Send(
		tele.ChatID(chatID),
		"Предоставьте номер телефона для регистрации в системе.",
		keyboard,
	); err != nil {
		return fmt.Errorf("request contact: %w", err)
	}
	return nil
}

func (contact *Contact) send(recipient tele.Recipient, text string) error {
	if _, err := contact.api.Send(recipient, text); err != nil {
		return fmt.Errorf("send contact message: %w", err)
	}
	return nil
}
