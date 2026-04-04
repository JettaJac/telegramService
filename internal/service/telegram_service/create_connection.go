package telegram_service

import (
	"context"
	"fmt"
	client "telegramservice/internal/clients/telegram_client"
	"telegramservice/internal/storage"
	"telegramservice/pkg/browser"
	"time"

	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

// CreateConnection ...
func (s *Service) CreateConnection(ctx context.Context, connectionID string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.validateConnectionID(connectionID); err != nil {
		return "", err
	}

	if connectionID == "" {
		s.logger.Error("connectionID empty")
		return "", fmt.Errorf("connectionID empty")
	}

	if _, err := s.checkConnectionExists(connectionID); err != nil {
		return "", err
	}

	if _, exists := s.clients.telegram[connectionID]; exists && s.clients.telegram[connectionID].IsAuthorized() {
		s.logger.Warn("connection already exists",
			zap.String("connection_id", connectionID))
		return "", fmt.Errorf("connection %s already exists", connectionID)
	}

	s.logger.Info("creating new connection",
		zap.String("connection_id", connectionID))

	var qrData string

	qrReady := s.createQRReadyCallback(connectionID, &qrData)

	tgClient := s.createTelegramClient(connectionID)

	s.startTelegramClient(tgClient, connectionID, qrReady)

	s.clients.telegram[connectionID] = tgClient

	s.saveConnectionInfo(connectionID)

	return qrData, nil
}

// validateConnectionID проверяет корректность connectionID
func (s *Service) validateConnectionID(connectionID string) error {
	if connectionID == "" {
		s.logger.Error("connectionID empty")
		return fmt.Errorf("connectionID empty")
	}
	return nil
}

// checkConnectionExists проверяет, существует ли уже соединение
func (s *Service) checkConnectionExists(connectionID string) (bool, error) {
	if c, exists := s.clients.telegram[connectionID]; exists && c.IsAuthorized() {
		s.logger.Warn("connection already exists",
			zap.String("connection_id", connectionID))
		return true, fmt.Errorf("connection %s already exists", connectionID)
	}
	return false, nil
}

// createQRReadyCallback создает callback для обработки QR кода
func (s *Service) createQRReadyCallback(connectionID string, qrData *string) func(string) {
	return func(data string) {
		*qrData = data
		s.storage.UpdateConnectionStatus(connectionID, "pending")
		s.logger.Info("QR code generated",
			zap.String("connection_id", connectionID),
			zap.Int("code_length", len(data)))

		s.openBrowserWithQR(connectionID, data)
	}
}

// openBrowserWithQR открывает браузер с QR кодом
func (s *Service) openBrowserWithQR(connectionID, qrData string) {
	if qrData == "" {
		return
	}

	qrServer := browser.NewQRServer(s.logger)
	url, err := qrServer.Start(qrData)
	if err != nil {
		s.logger.Error("failed to start QR server",
			zap.String("connection_id", connectionID),
			zap.Error(err))
		return
	}

	s.logger.Info("opening browser with QR code",
		zap.String("connection_id", connectionID),
		zap.String("url", url))

	if err := browser.OpenBrowser(url); err != nil {
		s.logger.Error("failed to open browser",
			zap.String("connection_id", connectionID),
			zap.Error(err))
	}
}

// createMessageHandler создает обработчик входящих сообщений
func (s *Service) createMessageHandler(connectionID string) func(string, *tg.Message) {
	return func(connID string, msg *tg.Message) {
		s.handleIncomingMessage(connID, msg)
	}
}

// createTelegramClient создает новый Telegram клиент
func (s *Service) createTelegramClient(connectionID string) *client.TelegramClient {
	clientConfig := client.Config{
		APIID:      s.apiID,
		APIHash:    s.apiHash,
		SessionDir: s.sessionDir,
	}

	handler := s.createMessageHandler(connectionID)

	return client.NewTelegramClient(connectionID, clientConfig, handler, s.logger)
}

func (s *Service) handleIncomingMessage(connectionID string, msg *tg.Message) {
	// Извлекаем информацию из сообщения
	senderID := s.extractSenderID(msg)
	chatID := s.extractChatID(msg)

	storageMsg := &storage.Message{
		ID:           generateID(),
		ConnectionID: connectionID,
		SenderID:     senderID,
		Text:         msg.Message,
		Timestamp:    time.Now().Unix(),
		ChatID:       chatID,
	}

	s.storage.AddMessage(storageMsg)

	s.logger.Info("received incoming message",
		zap.String("connection_id", connectionID),
		zap.String("message_id", storageMsg.ID),
		zap.String("sender_id", senderID),
		zap.String("chat_id", chatID),
		zap.String("text", msg.Message))
}

func (s *Service) extractSenderID(msg *tg.Message) string {
	switch peer := msg.PeerID.(type) {
	case *tg.PeerUser:
		return fmt.Sprintf("%d", peer.UserID)
	case *tg.PeerChat:
		return fmt.Sprintf("chat_%d", peer.ChatID)
	case *tg.PeerChannel:
		return fmt.Sprintf("channel_%d", peer.ChannelID)
	}
	return "unknown"
}

// startTelegramClient запускает Telegram клиент в фоне
func (s *Service) startTelegramClient(
	tgClient *client.TelegramClient,
	connectionID string,
	qrReady func(string),
) {
	go func() {
		clientCtx, cancelClient := context.WithCancel(context.Background())
		s.cancelFuncs[connectionID] = cancelClient

		if err := tgClient.Start(clientCtx, qrReady); err != nil {
			s.logger.Error("failed to start client",
				zap.String("connection_id", connectionID),
				zap.Error(err))
			s.storage.UpdateConnectionStatus(connectionID, "error")
		} else {
			s.storage.UpdateConnectionStatus(connectionID, "authorized")
		}
	}()
}

// saveConnectionInfo сохраняет информацию о соединении
func (s *Service) saveConnectionInfo(connectionID string) {
	connInfo := &storage.ConnectionInfo{
		ID:        connectionID,
		Status:    "pending",
		CreatedAt: time.Now().Unix(),
	}
	s.storage.SaveConnection(connInfo)
}

func (s *Service) extractChatID(msg *tg.Message) string {
	switch peer := msg.PeerID.(type) {
	case *tg.PeerUser:
		return fmt.Sprintf("%d", peer.UserID)
	case *tg.PeerChat:
		return fmt.Sprintf("%d", peer.ChatID)
	case *tg.PeerChannel:
		return fmt.Sprintf("%d", peer.ChannelID)
	}
	return "unknown"
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
