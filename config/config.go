package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken string

	EmailHost     string
	EmailPort     string
	EmailUser     string
	EmailPassword string

	IsTelegramValid bool
	IsEmailValid    bool
}

var Cfg *Config

func getEnv(name string) string {
	return os.Getenv("NOTEPHEE_" + name)
}

// load загружает конфигурацию из переменных окружения
func load() {
	_ = godotenv.Load()

	Cfg = &Config{
		TelegramToken: getEnv("TELEGRAM_TOKEN"),
		EmailHost:     getEnv("EMAIL_HOST"),
		EmailPort:     getEnv("EMAIL_PORT"),
		EmailUser:     getEnv("EMAIL_USER"),
		EmailPassword: getEnv("EMAIL_PASSWORD"),
	}

	if !Cfg.IsTelegramEnabled() {
		slog.Info("Конфигурация Telegram-бота не заполнена, функционал работы с этим сервисом ограничен")
	}
	if !Cfg.IsEmailEnabled() {
		slog.Info("Конфигурация для email не заполнена или заполнена частично, функционал отправки электронных писем ограничен")
	}
	if !Cfg.IsTelegramEnabled() && !Cfg.IsEmailEnabled() {
		slog.Error("Конфигурация Notephee не загружена, функционал недоступен")
	}
}

// Get возвращает текущий конфиг
func Get() *Config {
	if Cfg == nil {
		load()
	}
	return Cfg
}

func (c *Config) IsEmailEnabled() bool {
	return c.EmailHost != "" && c.EmailPort != "" && c.EmailUser != "" && c.EmailPassword != ""
}

func (c *Config) IsTelegramEnabled() bool {
	return c.TelegramToken != ""
}
