package telegram

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	
	"github.com/epheer/notephee/config"
)

// MessageOptions описывают параметры для отправки сообщения в чат
type MessageOptions struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

// SendingOptions описывают параметры для рассылки
type SendingOptions struct {
	ChatIDs []int64 `json:"chat_ids"`
	Text    string  `json:"text"`
}

// TgResponse описывает ответ Telegram
type TgResponse struct {
	OK          bool            `json:"ok"`
	Description string          `json:"description,omitempty"`
	Result      json.RawMessage `json:"result,omitempty"`
	ErrorCode   int             `json:"error_code,omitempty"`
	Parameters  struct {
		RetryAfter int `json:"retry_after,omitempty"`
	} `json:"parameters,omitempty"`
}

const (
	GetMe       = "/getMe"
	SendMessage = "/sendMessage"
)

var (
	telegramToken = config.Cfg.TelegramToken
	telegramURI   = fmt.Sprintf("https://api.telegram.org/bot%s", telegramToken)
)

func tg(method string) string {
	return fmt.Sprintf("%s%s", telegramURI, method)
}

func parseResponse(resp *http.Response) (*TgResponse, error) {
	if resp == nil {
		return nil, fmt.Errorf("ответ пустой")
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("Ошибка закрытия body в response: %v", err)
		}
	}()

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

// CheckConnection проверяет валидность Telegram
func CheckConnection() error {
	res, err := http.Get(tg(GetMe))
	if err != nil {
		return err
	}
	_, err = parseResponse(res)
	if err != nil {
		return err
	}
	config.Cfg.IsTelegramValid = true
	return nil
}
