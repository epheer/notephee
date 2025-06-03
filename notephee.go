package notephee

import (
	"github.com/epheer/notephee/config"
	"github.com/epheer/notephee/email"
	"github.com/epheer/notephee/telegram"
	"log/slog"
	"time"
)

func Init(logger *slog.Logger) {
	config.Get(logger)
	tg := telegram.NewTgClient(config.Cfg, logger)
	err := tg.CheckConnection()
	if err != nil {
		slog.Warn("Невозможно подключиться к Telegram. Проверьте валидность токена в NOTEPHEE_TELEGRAM_TOKEN.")
	}
	tg.NewBindingManager(10*time.Minute, logger)
	email.NewClient(config.Cfg, logger)
	slog.Info("Notephee готов 🚀")
}
