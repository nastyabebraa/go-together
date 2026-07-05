package bot

import (
	"context"
	"fmt"

	tele "gopkg.in/telebot.v3"

	utils "ride-together-bot/bot/utils"
	"ride-together-bot/conf"
	"ride-together-bot/db"
	"ride-together-bot/entity"
)

type API struct {
	api      *tele.Bot
	db       *db.DB
	config   *conf.Config
	sticker  utils.Sticker
	contact  *utils.Contact
	location *utils.Location
	event    *utils.Event
}

func NewBot(config *conf.Config, api *tele.Bot, database *db.DB) *API {
	sticker := utils.NewSticker(api)
	return &API{
		api:      api,
		db:       database,
		config:   config,
		sticker:  sticker,
		contact:  utils.NewContact(api, database, sticker, config.Stickers.Cat),
		location: utils.NewLocation(api, database, sticker, config),
		event:    utils.NewEvent(config, api, database, sticker),
	}
}

func (bot *API) Start(ctx context.Context) {
	bot.api.Handle("/start", bot.handleStart(ctx))
	bot.api.Handle("/auth", bot.handleAuth(ctx))
	bot.api.Handle("/new_ride", bot.handleNewRide(ctx))
	bot.api.Handle("/find", bot.handleFind(ctx))
	bot.api.Handle("/trips_management", bot.handleTripsManagement(ctx))
	bot.api.Handle("/history", bot.handleHistory(ctx))
	bot.api.Handle(tele.OnContact, bot.handleContact(ctx))
	bot.api.Handle(tele.OnLocation, bot.handleLocation(ctx))
	bot.api.Handle("Отмена", bot.handleCancel)

	go func() {
		<-ctx.Done()
		bot.api.Stop()
	}()
	bot.api.Start()
}

func (bot *API) handleStart(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		exists, err := bot.db.IsExists(ctx, c.Chat().ID)
		if err != nil {
			return err
		}
		message := fmt.Sprintf(
			"Привет, %s. Я бот для поиска попутчиков в любой системе каршеринга.\nПриятной экономии!",
			c.Sender().FirstName,
		)
		if !exists {
			message += "\n\n" + entity.NeedAuth
		}
		if _, err := bot.api.Send(c.Sender(), message); err != nil {
			return fmt.Errorf("send start message: %w", err)
		}
		return bot.sticker.SendSticker(c.Chat().ID, bot.config.Stickers.Start)
	}
}

func (bot *API) handleAuth(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		exists, err := bot.db.IsExists(ctx, c.Chat().ID)
		if err != nil {
			return err
		}
		if exists {
			if err := c.Send("Пользователь уже зарегистрирован."); err != nil {
				return fmt.Errorf("send registration status: %w", err)
			}
			return bot.sticker.SendSticker(c.Chat().ID, bot.config.Stickers.Shrek)
		}
		return bot.contact.Request(c.Chat().ID)
	}
}

func (bot *API) handleContact(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		exists, err := bot.db.IsExists(ctx, c.Chat().ID)
		if err != nil {
			return err
		}
		if exists {
			return c.Send("Пользователь уже зарегистрирован.", &tele.ReplyMarkup{RemoveKeyboard: true})
		}
		return bot.contact.Handle(ctx, c.Message())
	}
}

func (bot *API) handleNewRide(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		authorized, err := bot.ensureAuthorized(ctx, c)
		if err != nil || !authorized {
			return err
		}
		return bot.event.Create(c.Chat().ID)
	}
}

func (bot *API) handleFind(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		authorized, err := bot.ensureAuthorized(ctx, c)
		if err != nil || !authorized {
			return err
		}
		return bot.location.Request(c.Chat().ID)
	}
}

func (bot *API) handleLocation(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		authorized, err := bot.ensureAuthorized(ctx, c)
		if err != nil || !authorized {
			return err
		}
		pageURL, err := bot.location.Handle(ctx, c.Message())
		if err != nil {
			return err
		}
		keyboard := &tele.ReplyMarkup{ResizeKeyboard: true}
		keyboard.Reply(keyboard.Row(tele.Btn{Text: "Поездки", WebApp: &tele.WebApp{URL: pageURL}}))
		if err := c.Send("Список поездок в радиусе 1 км", keyboard); err != nil {
			return fmt.Errorf("send nearby trips: %w", err)
		}
		return nil
	}
}

func (bot *API) handleTripsManagement(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		authorized, err := bot.ensureAuthorized(ctx, c)
		if err != nil || !authorized {
			return err
		}
		return bot.event.TripsManagement(ctx, c.Message())
	}
}

func (bot *API) handleHistory(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		authorized, err := bot.ensureAuthorized(ctx, c)
		if err != nil || !authorized {
			return err
		}
		pageURL, err := bot.event.History(ctx, c.Chat().ID)
		if err != nil {
			return err
		}
		keyboard := &tele.ReplyMarkup{ResizeKeyboard: true}
		keyboard.Reply(keyboard.Row(tele.Btn{Text: "История", WebApp: &tele.WebApp{URL: pageURL}}))
		if err := c.Send("Откройте историю поездок.", keyboard); err != nil {
			return fmt.Errorf("send history button: %w", err)
		}
		return bot.sticker.SendSticker(c.Chat().ID, bot.config.Stickers.Cat)
	}
}

func (bot *API) handleCancel(c tele.Context) error {
	if err := c.Send("Регистрация отменена.", &tele.ReplyMarkup{RemoveKeyboard: true}); err != nil {
		return fmt.Errorf("cancel registration: %w", err)
	}
	return nil
}

func (bot *API) ensureAuthorized(ctx context.Context, c tele.Context) (bool, error) {
	exists, err := bot.db.IsExists(ctx, c.Chat().ID)
	if err != nil {
		return false, err
	}
	if exists {
		return true, nil
	}
	if _, err := bot.api.Send(c.Sender(), entity.NeedAuth); err != nil {
		return false, fmt.Errorf("send authorization requirement: %w", err)
	}
	return false, nil
}
