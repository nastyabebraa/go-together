package main

import (
	"context"
	"embed"
	"log"
	"os/signal"
	"ride-together-bot/bot"
	"ride-together-bot/conf"
	"ride-together-bot/db"
	"syscall"

	"github.com/pressly/goose/v3"
	tele "gopkg.in/telebot.v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := conf.Load()
	if err != nil {
		log.Fatal(err)
	}

	database, err := db.NewDatabase(ctx, cfg.DSN)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	goose.SetBaseFS(embedMigrations)

	err = goose.SetDialect("mysql")
	if err != nil {
		log.Fatal(err)
	}

	if err := goose.Up(database.Conn, "migrations"); err != nil {
		log.Fatal("goose up: ", err)
	}

	pref := tele.Settings{
		Token: cfg.TelegramBotAPIKey,
		Poller: &tele.LongPoller{
			LastUpdateID: 0,
		},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal("create telegram bot: ", err)
	}

	botInstance := bot.NewBot(cfg, b, database)
	botInstance.Start(ctx)
}
