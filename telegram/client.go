package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/epheer/notephee/config"
)

// MessageOptions содержит параметры для отправки одного текстового сообщения через Telegram Bot API.
type MessageOptions struct {
	ChatID int64  `json:"chat_id"` // Идентификатор чата Telegram
	Text   string `json:"text"`    // Текст сообщения
}

// SendingOptions используется для массовой отправки сообщений по нескольким chatID.
type SendingOptions struct {
	ChatIDs []int64 `json:"chat_ids"` // Список идентификаторов чатов
	Text    string  `json:"text"`     // Текст сообщения
}

// TgResponse представляет ответ Telegram Bot API на любой метод.
type TgResponse struct {
	OK          bool            `json:"ok"`                    // Успешность запроса
	Description string          `json:"description,omitempty"` // Описание ошибки (если есть)
	Result      json.RawMessage `json:"result,omitempty"`      // Сырые данные результата (если есть)
	ErrorCode   int             `json:"error_code,omitempty"`  // Код ошибки (если есть)
	Parameters  struct {
		RetryAfter int `json:"retry_after,omitempty"` // Рекомендованная задержка перед повтором (rate limit)
	} `json:"parameters,omitempty"`
}

// TgClient инкапсулирует клиента Telegram Bot API.
type TgClient struct {
	token  string       // Токен Telegram бота
	name   string       // Имя Telegram бота
	uri    string       // Базовый URL API
	http   *http.Client // HTTP-клиент
	logger *slog.Logger // Логгер для отладки
}

// SendResult представляет результат отправки одного сообщения.
type SendResult struct {
	ChatID   int64       // Идентификатор получателя
	Response *TgResponse // Ответ Telegram API
	Error    error       // Ошибка, если произошла
}

// Константы Telegram API методов
const (
	GetMe       = "/getMe"
	SendMessage = "/sendMessage"
)

// NewTgClient создаёт и возвращает нового клиента Telegram.
//
// cfg — конфигурация приложения с токеном и именем бота.
// logger — логгер для ведения журнала.
func NewTgClient(cfg *config.Config, logger *slog.Logger) *TgClient {
	uri := fmt.Sprintf("https://api.telegram.org/bot%s", cfg.TelegramToken)
	return &TgClient{
		token:  cfg.TelegramToken,
		name:   cfg.TelegramBotName,
		uri:    uri,
		http:   &http.Client{Timeout: 10 * time.Second},
		logger: logger,
	}
}

// tg возвращает полный URL для метода Telegram API.
func (c *TgClient) tg(method string) string {
	return fmt.Sprintf("%s%s", c.uri, method)
}

// parseResponse декодирует HTTP-ответ от Telegram API в TgResponse.
//
// Возвращает ошибку, если API ответил неуспешно или формат JSON некорректен.
func (c *TgClient) parseResponse(resp *http.Response) (*TgResponse, error) {
	if resp == nil {
		return nil, fmt.Errorf("ответ пустой")
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("невозможно прочесть body: %w", err)
	}

	var tgResp TgResponse
	err = json.Unmarshal(body, &tgResp)
	if err != nil {
		return nil, fmt.Errorf("некорректный формат JSON: %w (body: %s)", err, string(body))
	}

	if !tgResp.OK {
		errMsg := tgResp.Description
		if tgResp.ErrorCode != 0 {
			errMsg = fmt.Sprintf("Код ошибки %d: %s", tgResp.ErrorCode, errMsg)
		}
		if tgResp.Parameters.RetryAfter > 0 {
			errMsg = fmt.Sprintf("%s (повторите через %d секунд)", errMsg, tgResp.Parameters.RetryAfter)
		}
		return &tgResp, fmt.Errorf("ошибка Telegram Bot Api: %s", errMsg)
	}

	return &tgResp, nil
}

// postReq отправляет POST-запрос с JSON-данными к Telegram API.
//
// Возвращает результат и ошибку (если есть).
func (c *TgClient) postReq(data json.RawMessage, method string) (*TgResponse, error) {
	res, err := c.http.Post(c.tg(method), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	return c.parseResponse(res)
}

// CheckConnection проверяет доступность Telegram API через метод getMe.
//
// Возвращает ошибку, если соединение не удалось или API вернул ошибку.
func (c *TgClient) CheckConnection() error {
	res, err := c.http.Get(c.tg(GetMe))
	if err != nil {
		return err
	}
	_, err = c.parseResponse(res)
	return err
}

// SendText отправляет одно текстовое сообщение.
//
// Возвращает TgResponse и ошибку (если произошла).
func (c *TgClient) SendText(options MessageOptions) (TgResponse, error) {
	data, err := json.Marshal(options)
	if err != nil {
		return TgResponse{}, err
	}
	res, err := c.postReq(data, SendMessage)
	if err != nil {
		return TgResponse{}, err
	}
	return *res, nil
}

// SendMessaging отправляет одно и то же сообщение множеству получателей с соблюдением rate limit.
//
// Возвращает срез результатов по каждому получателю.
func (c *TgClient) SendMessaging(options SendingOptions) []SendResult {
	limiter := rate.NewLimiter(rate.Every(time.Second/30), 1)

	var (
		results = make([]SendResult, 0, len(options.ChatIDs))
		mu      sync.Mutex
		wg      sync.WaitGroup
	)

	for _, chatID := range options.ChatIDs {
		wg.Add(1)

		go func(chatID int64) {
			defer wg.Done()

			// Ожидание разрешения лимитера
			if err := limiter.Wait(context.Background()); err != nil {
				c.logger.Error("лимитер не пропустил", "chat_id", chatID, "error", err)
				mu.Lock()
				results = append(results, SendResult{ChatID: chatID, Error: err})
				mu.Unlock()
				return
			}

			msg := MessageOptions{ChatID: chatID, Text: options.Text}
			resp, err := c.SendText(msg)

			mu.Lock()
			results = append(results, SendResult{ChatID: chatID, Response: &resp, Error: err})
			mu.Unlock()
		}(chatID)
	}

	wg.Wait()
	return results
}
