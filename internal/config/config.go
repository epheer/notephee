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
}

var cfg *Config

func getEnv(name string) string {
	return os.Getenv("NOTEPHEE_" + name)
}

// load загружает конфигурацию из переменных окружения
func load() {
	_ = godotenv.Load()

	cfg = &Config{
		TelegramToken: getEnv("TELEGRAM_TOKEN"),
		EmailHost:     getEnv("EMAIL_HOST"),
		EmailPort:     getEnv("EMAIL_PORT"),
		EmailUser:     getEnv("EMAIL_USER"),
		EmailPassword: getEnv("EMAIL_PASSWORD"),
	}

	if !cfg.IsTelegramEnabled() {
		slog.Info("Конфигурация Telegram-бота не заполнена, функционал работы с этим сервисом ограничен")
	}
	if !cfg.IsEmailEnabled() {
		slog.Info("Конфигурация для email не заполнена или заполнена частично, функционал отправки электронных писем ограничен")
	}
	if !cfg.IsTelegramEnabled() && !cfg.IsEmailEnabled() {
		slog.Error("Конфигурация Notephee не загружена, функционал недоступен")
	}
}

// Get возвращает текущий конфиг
func Get() *Config {
	if cfg == nil {
		load()
	}
	return cfg
}

func (c *Config) IsEmailEnabled() bool {
	return c.EmailHost != "" && c.EmailPort != "" && c.EmailUser != "" && c.EmailPassword != ""
}

func (c *Config) IsTelegramEnabled() bool {
	return c.TelegramToken != ""
}
