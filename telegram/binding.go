package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Binding представляет успешную привязку между внутренним userID и Telegram chatID.
type Binding struct {
	UserID string // Внутренний идентификатор пользователя
	ChatID int64  // Идентификатор чата в Telegram
}

// pendingBinding хранит временные данные до подтверждения пользователем /start.
type pendingBinding struct {
	UserID string    // Внутренний ID пользователя, инициировавшего инвайт
	Expiry time.Time // Время окончания действия инвайта
}

// BindingManager управляет созданием и проверкой Telegram-инвайтов.
type BindingManager struct {
	store  sync.Map      // Хранилище инвайтов по UUID
	ttl    time.Duration // Время жизни каждого инвайта
	logger *slog.Logger  // Логгер для отладки
	bot    string        // Имя Telegram-бота
}

// Update представляет одно обновление от Telegram API (например, входящее сообщение).
type Update struct {
	UpdateID int64 `json:"update_id"` // ID обновления
	Message  struct {
		Text string `json:"text"` // Текст сообщения
		Chat struct {
			ID int64 `json:"id"` // Chat ID, с которого пришло сообщение
		} `json:"chat"`
	} `json:"message"`
}

// UpdatesResponse — структура ответа Telegram API на метод getUpdates.
type UpdatesResponse struct {
	OK     bool     `json:"ok"`     // Статус ответа API
	Result []Update `json:"result"` // Список новых обновлений
}

// NewBindingManager создаёт новый BindingManager с заданным временем жизни инвайтов.
//
// Возвращает nil, если Telegram отключён.
func (c *TgClient) NewBindingManager(ttl time.Duration, logger *slog.Logger) *BindingManager {
	if !c.Enabled {
		logger.Warn("Попытка создать BindingManager, но Telegram отключён")
		return nil
	}

	return &BindingManager{
		ttl:    ttl,
		logger: logger,
		bot:    c.name,
	}
}

// CreateInvite создаёт инвайт-ссылку для Telegram, которая будет доступна в течение ttl.
// Возвращает ссылку вида: https://t.me/<bot>?start=<uuid>
//
// userID — идентификатор пользователя, которому создаётся инвайт.
func (bm *BindingManager) CreateInvite(userID string) string {
	inviteCode := uuid.New().String()
	bm.store.Store(inviteCode, pendingBinding{
		UserID: userID,
		Expiry: time.Now().Add(bm.ttl),
	})

	go func() {
		time.Sleep(bm.ttl)
		bm.store.Delete(inviteCode)
	}()

	return fmt.Sprintf("https://t.me/%s?start=%s", bm.bot, inviteCode)
}

// ResolveBinding проверяет, существует ли данный инвайт и создаёт привязку chatID к userID.
//
// uuid — код из ссылки Telegram (/start <uuid>).
// chatID — идентификатор Telegram-чата, инициировавшего запрос.
//
// Возвращает Binding, если UUID действителен, или ошибку — если нет.
func (bm *BindingManager) ResolveBinding(uuid string, chatID int64) (*Binding, error) {
	val, ok := bm.store.Load(uuid)
	if !ok {
		return nil, fmt.Errorf("инвайт просрочен или не найден")
	}
	bm.store.Delete(uuid)

	p := val.(pendingBinding)
	return &Binding{
		UserID: p.UserID,
		ChatID: chatID,
	}, nil
}

// StartPolling запускает постоянный опрос Telegram Bot API методом getUpdates.
// При получении команды /start с UUID пытается выполнить привязку и вызывает callback.
//
// ctx — контекст, по завершении которого polling будет остановлен.
// bm — менеджер инвайтов для проверки кодов /start.
// callback — вызывается при успешной привязке.
func (c *TgClient) StartPolling(ctx context.Context, bm *BindingManager, callback func(Binding)) {
	if !c.Enabled {
		c.logger.Warn("StartPolling не запущен: Telegram отключён")
		return
	}

	if bm == nil {
		c.logger.Warn("StartPolling не запущен: BindingManager == nil")
		return
	}

	var offset int64

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Polling stopped")
			return
		default:
		}

		url := fmt.Sprintf("%s/getUpdates?timeout=30&offset=%d", c.uri, offset)
		resp, err := c.http.Get(url)
		if err != nil {
			c.logger.Error("Ошибка при запросе getUpdates", "error", err)
			time.Sleep(2 * time.Second)
			continue
		}

		var updates UpdatesResponse
		if err := json.NewDecoder(resp.Body).Decode(&updates); err != nil {
			c.logger.Error("Ошибка декодирования ответа", "error", err)
			_ = resp.Body.Close()
			time.Sleep(2 * time.Second)
			continue
		}
		_ = resp.Body.Close()

		for _, upd := range updates.Result {
			offset = upd.UpdateID + 1

			text := upd.Message.Text
			chatID := upd.Message.Chat.ID

			if strings.HasPrefix(text, "/start ") {
				inviteCode := strings.TrimPrefix(text, "/start ")
				binding, err := bm.ResolveBinding(inviteCode, chatID)
				if err != nil {
					c.logger.Warn("uuid не найден", "uuid", inviteCode, "chatID", chatID)
					continue
				}
				callback(*binding)
			}
		}
	}
}
