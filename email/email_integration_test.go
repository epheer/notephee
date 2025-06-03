package email_test

import (
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/epheer/notephee/config"
	"github.com/epheer/notephee/email"
)

func TestEmailIntegration(t *testing.T) {
	err := config.LoadEnv("../.env")
	if err != nil {
		t.Fatalf("Не найден .env")
	}
	config.Get(slog.Default())

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := config.Cfg
	client := email.NewClient(cfg, logger)
	if !client.Enabled {
		t.Skip("Email отправка отключена в конфиге, пропускаем тест")
	}

	testRecipient := os.Getenv("EMAIL_TEST_RECIPIENT")
	if testRecipient == "" {
		t.Fatal("Не задан test email получателя (TestEmailRecipient) в конфиге")
	}

	subject := "Тестовое письмо от notephee"
	body := fmt.Sprintf("Письмо отправлено в %v\nВремя: %v", testRecipient, time.Now().Format(time.RFC1123))

	fmt.Println("Отправка письма...")
	err = client.SendText(email.MessageOptions{
		To:      testRecipient,
		Subject: subject,
		Body:    body,
	})
	fmt.Println("SendText завершился")

	if err != nil {
		t.Fatalf("Ошибка отправки email: %v", err)
	}

	fmt.Println("Email успешно отправлен на:", testRecipient)
}
