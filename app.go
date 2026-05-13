package main

import (
	"bot/api"
	"bot/bot"
	"bot/repo"
	"log"
)

type App struct {
	cfgPath string
	cfg     *Config
	repo    *repo.Repo
	api     *api.API
	bot     *bot.BotManager
}

func NewApp(path string) *App {
	return &App{
		cfgPath: path,
	}
}

func (app *App) init() {
	var err error
	app.cfg, err = LoadConfig(app.cfgPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	app.repo = repo.NewRepo(&app.cfg.Database)
	app.api = api.NewAPI(&app.cfg.HttpServe, app.repo)
	app.bot = bot.NewBotManager(app.repo, "") // TODO: webhookURL
}

func (app *App) Start() {
	app.init()
	go app.api.Run()
	go app.bot.Start()
}

func (app *App) Stop() {
	app.bot.Stop()
	app.api.Shutdown()
}
