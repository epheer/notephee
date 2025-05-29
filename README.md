# Notephee

**Notephee** - модуль Go для гибких уведомлений пользователя по электронной почте или Telegram. С помощью Notephee можно организовать:

- массовые или целевые рассылки;
- сервисные уведомления;
- отправку кодов аутентификации

и многое другое.

## Требования к окружению

- Go 1.24 и выше

## Установка модуля

1. Выполните команду
```bash
go get github.com/epheer/notephee
```

2. Скопируйте `.env.dist` в свой `.env` и задайте эти значения переменных среды
```dotenv
# Настройка Telegram для Notephee
NOTEPHEE_TELEGRAM_TOKEN=

# Настройка Email для Notephee
NOTEPHEE_SMTP_HOST=
NOTEPHEE_SMTP_PORT=
NOTEPHEE_SMTP_USER=
NOTEPHEE_SMTP_PASSWORD= 
```