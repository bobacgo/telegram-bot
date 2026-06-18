package bot

import (
	"bot/api"
	"bot/bot"
	"bot/bus"
	"bot/repo"
	"log"
)

type App struct {
	cfgPath string
	cfg     *Config
	bus     *bus.Bus
	repo    *repo.Repo
	api     *api.API
	bot     *bot.Manager
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

	app.bus = bus.NewBus()
	app.repo = repo.NewRepo(&app.cfg.Database)
	app.api = api.NewAPI(&app.cfg.HttpServe, app.bus, app.repo)
	app.bot = bot.NewManager("", app.bus, app.repo) // TODO: webhookURL
}

func (app *App) Start() {
	app.init()
	go app.api.Run()
	go app.bot.Start()
}

func (app *App) Stop() {
	app.bus.Stop()
	app.bot.Stop()
	app.api.Shutdown()
}
