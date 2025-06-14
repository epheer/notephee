package config

import (
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken   string
	TelegramBotName string

	EmailHost     string
	EmailPort     string
	EmailUser     string
	EmailPassword string
	EmailFromName string

	IsTelegramValid bool
	IsEmailValid    bool
}

var Cfg *Config

// LoadEnv загружает переменные из env-файла
func LoadEnv(path string) error {
	err := godotenv.Load(path)
	if err != nil {
		log.Printf("Не удалось загрузить .env файл по пути %s: %v", path, err)
	}
	return err
}

func getEnv(name string) string {
	return os.Getenv("NOTEPHEE_" + name)
}

// load загружает конфигурацию из переменных окружения
func load(logger *slog.Logger) {
	Cfg = &Config{
		TelegramToken:   getEnv("TELEGRAM_TOKEN"),
		TelegramBotName: getEnv("TELEGRAM_BOT_NAME"),
		EmailHost:       getEnv("SMTP_HOST"),
		EmailPort:       getEnv("SMTP_PORT"),
		EmailUser:       getEnv("SMTP_USER"),
		EmailPassword:   getEnv("SMTP_PASSWORD"),
		EmailFromName:   getEnv("SMTP_FROM_NAME"),
	}

	if !Cfg.IsTelegramEnabled() {
		logger.Info("Конфигурация Telegram-бота не заполнена или заполнена частично, функционал работы с этим сервисом ограничен")
	}
	if !Cfg.IsEmailEnabled() {
		logger.Info("Конфигурация для email не заполнена или заполнена частично, функционал отправки электронных писем ограничен")
	}
	if !Cfg.IsTelegramEnabled() && !Cfg.IsEmailEnabled() {
		logger.Error("Конфигурация Notephee не загружена, функционал недоступен")
	}
}

// Get возвращает текущий конфиг
func Get(logger *slog.Logger) *Config {
	if Cfg == nil {
		load(logger)
	}
	return Cfg
}

func (c *Config) IsEmailEnabled() bool {
	return c.EmailHost != "" && c.EmailPort != "" && c.EmailUser != "" && c.EmailPassword != ""
}

func (c *Config) IsTelegramEnabled() bool {
	return c.TelegramToken != "" && c.TelegramBotName != ""
}
