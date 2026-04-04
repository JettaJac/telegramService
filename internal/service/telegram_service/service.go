package telegram_service

import (
	"context"
	"sync"
	"telegramservice/pkg/logger"
	"time"

	"go.uber.org/zap"
)

// Service ...
type Service struct {
	clients     *ExternalServicesClients
	cancelFuncs map[string]context.CancelFunc
	storage     StorageInterface
	logger      *zap.Logger
	mu          sync.RWMutex
	apiID       int
	apiHash     string
	sessionDir  string
	qrTimeout   time.Duration
}

// ExternalServicesClients сторонние сервисы
type ExternalServicesClients struct {
	telegram map[string]TelegramClientInterface
}

// NewExternalServicesClients создаем новый экземпляр
func NewExternalServicesClients() *ExternalServicesClients {
	return &ExternalServicesClients{
		telegram: make(map[string]TelegramClientInterface),
	}
}

// NewService ...
func NewService(apiID int, apiHash, sessionDir string, storage StorageInterface, qrTimeout time.Duration) *Service {
	return &Service{
		clients:     NewExternalServicesClients(),
		cancelFuncs: make(map[string]context.CancelFunc),
		storage:     storage,
		logger:      logger.NewLogger(),
		apiID:       apiID,
		apiHash:     apiHash,
		sessionDir:  sessionDir,
		qrTimeout:   qrTimeout,
	}
}
