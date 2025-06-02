package notephee

import (
	"log/slog"

	"github.com/epheer/notephee/config"
	"github.com/epheer/notephee/telegram"
)

func Init() {
	config.Get()
	logger := slog.Default()
	tg := telegram.NewTgClient(config.Cfg, logger)
	err := tg.CheckConnection()
	if err != nil {
		slog.Warn("Невозможно подключиться к Telegram. Проверьте валидность токена в NOTEPHEE_TELEGRAM_TOKEN.")
	}
	slog.Info("Notephee готов 🚀")
}
