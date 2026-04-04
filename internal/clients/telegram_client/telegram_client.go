package telegram_client

import (
	"context"
	"sync"

	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

// TelegramClient телеграм клиент
type TelegramClient struct {
	id             string
	apiID          int
	apiHash        string
	client         TelegramClientInterface
	sessionDir     string
	messageHandler MessageHandler
	logger         *zap.Logger
	mu             sync.RWMutex
	authorized     bool
	cancelFunc     context.CancelFunc
	//msgSender      *message.Sender
	//peerResolver   *peer.Resolver
}

// MessageHandler ...
type MessageHandler func(connectionID string, msg *tg.Message)

// Config ...
type Config struct {
	APIID      int
	APIHash    string
	SessionDir string
}

// NewTelegramClient созаем телеграмм клиента
func NewTelegramClient(id string, config Config, handler MessageHandler, logger *zap.Logger) *TelegramClient {
	return &TelegramClient{
		id:             id,
		apiID:          config.APIID,
		apiHash:        config.APIHash,
		sessionDir:     config.SessionDir,
		messageHandler: handler,
		logger:         logger.With(zap.String("client_id", id)),
	}
}
