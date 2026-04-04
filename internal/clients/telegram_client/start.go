package telegram_client

import (
	"context"
	"fmt"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth/qrlogin"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

// Start запускаем соединение
func (c *TelegramClient) Start(ctx context.Context, qrCodeReady func(string)) error {
	c.logger.Info("TelegramClient.Start called",
		zap.String("client_id", c.id),
		zap.Bool("qrCodeReady_is_nil", qrCodeReady == nil))

	sessionStorage := &session.FileStorage{
		Path: fmt.Sprintf("%s/%s.session", c.sessionDir, c.id),
	}

	dispatcher := c.setupDispatcher()

	adapterClient := telegram.NewClient(c.apiID, c.apiHash, telegram.Options{
		SessionStorage: sessionStorage,
		Logger:         c.logger,
		UpdateHandler:  dispatcher,
	})

	c.client = NewTelegramClientAdapter(adapterClient)

	ctx, cancel := context.WithCancel(ctx)
	c.cancelFunc = cancel

	return c.client.Run(ctx, func(ctx context.Context) error {
		c.logger.Info("client.Run started", zap.String("client_id", c.id))

		if err := c.runAuthFlow(ctx, qrCodeReady); err != nil {
			return err
		}

		return c.listenMessages(ctx)
	})
}

// setupDispatcher настраивает обработчик сообщений
func (c *TelegramClient) setupDispatcher() tg.UpdateDispatcher {
	dispatcher := tg.NewUpdateDispatcher()
	dispatcher.OnNewMessage(c.handleNewMessage)
	return dispatcher
}

// runAuthFlow запускает процесс авторизации
func (c *TelegramClient) runAuthFlow(ctx context.Context, qrCodeReady func(string)) error {
	isAuth, err := c.isAlreadyAuthorized(ctx)
	if err != nil {
		return err
	}
	if isAuth {
		return nil
	}

	return c.performQRAuth(ctx, qrCodeReady)
}

// isAlreadyAuthorized проверяет и обрабатывает уже авторизованный клиент
func (c *TelegramClient) isAlreadyAuthorized(ctx context.Context) (bool, error) {
	authorized, err := c.handleAuthStatus(ctx)
	if err != nil {
		return false, err
	}

	if authorized {
		c.markAuthorized()
		c.logger.Info("client already authorized", zap.String("client_id", c.id))
		return true, nil
	}

	return false, nil
}

// performQRAuth выполняет QR авторизацию
func (c *TelegramClient) performQRAuth(ctx context.Context, qrCodeReady func(string)) error {
	c.logger.Info("starting QR authorization", zap.String("client_id", c.id))

	loggedIn := make(chan struct{})
	qr := c.client.QR()

	qrHandler := c.makeQRHandler(qrCodeReady)

	_, err := qr.Auth(ctx, loggedIn, qrHandler)
	if err != nil {
		return fmt.Errorf("QR authorization failed: %w", err)
	}

	c.logger.Info("waiting for QR scan...", zap.String("client_id", c.id))
	<-loggedIn

	c.markAuthorized()
	c.logger.Info("client authorized via QR", zap.String("client_id", c.id))

	return nil
}

// handleAuthStatus проверяет статус авторизации
func (c *TelegramClient) handleAuthStatus(ctx context.Context) (bool, error) {
	c.logger.Info("checking auth status", zap.String("client_id", c.id))

	status, err := c.client.Auth().Status(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get auth status: %w", err)
	}

	return status.Authorized, nil
}

// markAuthorized помечает клиент как авторизованный
func (c *TelegramClient) markAuthorized() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.authorized = true
	c.logger.Info("client marked as authorized", zap.String("client_id", c.id))
}

// makeQRHandler создает обработчик QR кода
func (c *TelegramClient) makeQRHandler(qrCodeReady func(string)) func(ctx context.Context, token qrlogin.Token) error {
	return func(ctx context.Context, token qrlogin.Token) error {
		qrCodeData := token.String()
		if qrCodeReady != nil {
			qrCodeReady(qrCodeData)
		}
		c.logger.Info("QR code generated",
			zap.String("client_id", c.id),
			zap.Int("code_length", len(qrCodeData)))
		return nil
	}
}

func (c *TelegramClient) handleNewMessage(ctx context.Context, entities tg.Entities, update *tg.UpdateNewMessage) error {
	msg, ok := update.Message.(*tg.Message)
	if !ok {
		return nil
	}

	if msg.PeerID == nil || msg.Message == "" {
		return nil
	}

	senderID := c.extractSenderID(msg)

	c.logger.Info("received message",
		zap.String("sender_id", senderID),
		zap.String("text", msg.Message),
		zap.Int("message_id", msg.ID))

	if c.messageHandler != nil {
		c.messageHandler(c.id, msg)
	}

	return nil
}

func (c *TelegramClient) extractSenderID(msg *tg.Message) string {
	switch id := msg.PeerID.(type) {
	case *tg.PeerUser:
		// Личное сообщение — возвращаем ID пользователя
		return fmt.Sprintf("user_%d", id.UserID)
	case *tg.PeerChat:
		// Сообщение из группы — возвращаем ID чата
		return fmt.Sprintf("chat_%d", id.ChatID)
	case *tg.PeerChannel:
		// Сообщение из канала/супергруппы — возвращаем ID канала
		return fmt.Sprintf("channel_%d", id.ChannelID)
	default:
		return "unknown"
	}
}

func (c *TelegramClient) listenMessages(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}
