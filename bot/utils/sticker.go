package bot

import (
	"fmt"
	"strings"

	tele "gopkg.in/telebot.v3"
)

type Sticker struct {
	api *tele.Bot
}

func NewSticker(api *tele.Bot) Sticker {
	return Sticker{api: api}
}

func (sticker Sticker) SendSticker(chatID int64, stickerID string) error {
	if strings.TrimSpace(stickerID) == "" {
		return nil
	}
	file := &tele.Sticker{File: tele.File{FileID: stickerID}}
	if _, err := sticker.api.Send(tele.ChatID(chatID), file); err != nil {
		return fmt.Errorf("send sticker: %w", err)
	}
	return nil
}
