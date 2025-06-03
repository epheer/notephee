package telegram_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/epheer/notephee/config"
	"github.com/epheer/notephee/telegram"
)

func TestTelegramIntegration(t *testing.T) {
	err := config.LoadEnv("../.env")
	if err != nil {
		t.Fatalf("Не найден .env")
	}
	config.Get(slog.Default())

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := config.Cfg
	client := telegram.NewTgClient(cfg, logger)
	bm := client.NewBindingManager(10*time.Minute, logger)

	userID := "notephee_test"
	inviteLink := bm.CreateInvite(userID)

	fmt.Println("\n=== INVITE ===")
	fmt.Printf("Переходи по ссылке в Telegram: %s\n", inviteLink)

	parts := strings.Split(inviteLink, "=")
	if len(parts) != 2 {
		t.Fatalf("неверный формат invite ссылки")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	chatIDCh := make(chan int64, 1)

	callback := func(binding telegram.Binding) {
		fmt.Println("\n=== CALLBACK ВЫЗВАН ===")
		fmt.Printf("UserID: %s\n", binding.UserID)
		fmt.Printf("ChatID: %d\n", binding.ChatID)
		chatIDCh <- binding.ChatID
	}

	go client.StartPolling(ctx, bm, callback)

	fmt.Println("Запущен polling. Перейди по ссылке в Telegram и нажми /start.")

	var chatID int64
	select {
	case chatID = <-chatIDCh:
		fmt.Println("Тестирование получения chatID пройдено успешно")
	case <-ctx.Done():
		t.Fatal("Таймаут ожидания chatID из Telegram")
	}

	if err := client.CheckConnection(); err != nil {
		t.Fatalf("Ошибка CheckConnection: %v", err)
	}

	resp, err := client.SendText(telegram.MessageOptions{
		ChatID: chatID,
		Text:   fmt.Sprintf("✅ Интеграционный тест прошёл успешно \nChat ID: %v", chatID),
	})
	if err != nil {
		t.Fatalf("Ошибка при отправке сообщения: %v", err)
	}
	if !resp.OK {
		t.Fatalf("Telegram API вернул ошибку: %s", resp.Description)
	}

	fmt.Println("Тест успешно завершён.")
}
