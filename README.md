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

2. Скопируйте `.env.dist` в свой `.env` и задайте значения переменных среды
```dotenv
# Настройка Telegram для Notephee
NOTEPHEE_TELEGRAM_TOKEN=
NOTEPHEE_TELEGRAM_BOT_NAME=

# Настройка Email для Notephee
NOTEPHEE_SMTP_HOST=
NOTEPHEE_SMTP_PORT=
NOTEPHEE_SMTP_USER=
NOTEPHEE_SMTP_PASSWORD=
NOTEPHEE_SMTP_FROM_NAME=
```

3. Инициализируйте Notephee

```go
package main

import (
	"log/slog"
	
	"github.com/epheer/notephee"
)

func main() {
	notephee.Init(slog.Default())
}
```

## Зависимости

- [github.com/google/uuid](https://pkg.go.dev/github.com/google/uuid) – v1.6.0
- [github.com/joho/godotenv](https://pkg.go.dev/github.com/joho/godotenv) – v1.5.1
- [golang.org/x/time](https://pkg.go.dev/golang.org/x/time) – v0.11.0

## Тестирование

Модуль покрыт интеграционными тестами. Для тестирования вам необходимо:

1. Скопировать `.env.dist` в `.env` и задать валидные значения, а так же добавить следующие записи
```dotenv
# Переменные для тестов
EMAIL_TEST_RECIPIENT= 
```
2. Вызвать тест пакета командой
```bash
go test -v ./PACKAGE_DIRECTORY
```