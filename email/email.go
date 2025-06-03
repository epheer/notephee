package email

import (
	"context"
	"fmt"
	"log/slog"
	"mime"
	"net/smtp"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/epheer/notephee/config"
)

// MessageOptions содержит параметры для отправки одного письма.
type MessageOptions struct {
	To      string // Email получателя
	Subject string // Тема письма
	Body    string // Содержимое письма (в формате text/plain)
}

// SendingOptions содержит данные для массовой рассылки.
type SendingOptions struct {
	Recipients []string // Список email-адресов
	Subject    string   // Общая тема письма
	Body       string   // Общий текст письма
}

// EmailResponse содержит результат одной отправки.
type EmailResponse struct {
	To    string // Адрес получателя
	Error error  // Ошибка отправки (если была)
}

// Client инкапсулирует SMTP-клиент.
type Client struct {
	auth     smtp.Auth    // SMTP авторизация
	url      string       // Полный адрес SMTP-сервера (host:port)
	from     string       // От кого отправлять письма
	fromName string       // Отображаемое имя
	logger   *slog.Logger // Логгер
	Enabled  bool         // Разрешена ли отправка
}

// NewClient создаёт и возвращает Email клиента.
func NewClient(cfg *config.Config, logger *slog.Logger) *Client {
	url := fmt.Sprintf("%s:%s", cfg.EmailHost, cfg.EmailPort)

	return &Client{
		auth:     smtp.PlainAuth("", cfg.EmailUser, cfg.EmailPassword, cfg.EmailHost),
		url:      url,
		from:     cfg.EmailUser,
		fromName: cfg.EmailFromName,
		logger:   logger,
		Enabled:  cfg.IsEmailEnabled(),
	}
}

func encodeSubject(subject string) string {
	return mime.BEncoding.Encode("utf-8", subject)
}

// formatMessage формирует SMTP-сообщение из входных данных.
func (c *Client) formatMessage(to, subject, body string) []byte {
	encodedName := mime.BEncoding.Encode("utf-8", c.fromName)
	fromHeader := fmt.Sprintf("%s <%s>", encodedName, c.from)

	subjectHeader := mime.BEncoding.Encode("utf-8", subject)

	return []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s",
		fromHeader, to, subjectHeader, body,
	))
}

// SendText отправляет одно текстовое сообщение на email.
func (c *Client) SendText(options MessageOptions) error {
	if !c.Enabled {
		return fmt.Errorf("email-отправка отключена: конфигурация недоступна")
	}

	msg := c.formatMessage(options.To, options.Subject, options.Body)
	err := smtp.SendMail(c.url, c.auth, c.from, []string{options.To}, msg)
	if err != nil {
		return fmt.Errorf("ошибка отправки на %s: %w", options.To, err)
	}
	return nil
}

// SendMessaging отправляет письмо нескольким получателям с rate limit.
func (c *Client) SendMessaging(options SendingOptions) []EmailResponse {
	if !c.Enabled {
		c.logger.Warn("отправка email отключена: возвращаем заглушку")
		results := make([]EmailResponse, 0, len(options.Recipients))
		for _, to := range options.Recipients {
			results = append(results, EmailResponse{
				To:    to,
				Error: fmt.Errorf("email-отправка отключена"),
			})
		}
		return results
	}

	limiter := rate.NewLimiter(rate.Every(2*time.Second), 1)

	var (
		results = make([]EmailResponse, 0, len(options.Recipients))
		mu      sync.Mutex
		wg      sync.WaitGroup
	)

	for _, to := range options.Recipients {
		wg.Add(1)

		go func(to string) {
			defer wg.Done()

			if err := limiter.Wait(context.Background()); err != nil {
				c.logger.Error("лимитер не пропустил", "to", to, "error", err)
				mu.Lock()
				results = append(results, EmailResponse{To: to, Error: err})
				mu.Unlock()
				return
			}

			msg := MessageOptions{
				To:      to,
				Subject: options.Subject,
				Body:    options.Body,
			}

			err := c.SendText(msg)

			if err != nil {
				c.logger.Error("не удалось отправить email", "to", to, "error", err)
			}

			mu.Lock()
			results = append(results, EmailResponse{To: to, Error: err})
			mu.Unlock()
		}(to)
	}

	wg.Wait()
	return results
}
