package telegram_service

import (
	"context"
	"telegramservice/internal/storage"
	log "telegramservice/pkg/logger"

	"go.uber.org/zap"
)

type Mocks struct {
	storage      *storage.MemoryStorage
	mockTelegram map[string]MockTelegramClientInterface
	mockLogger   *zap.Logger
}

func NewMocks(
	mockTelegram map[string]MockTelegramClientInterface,
) (*Service, *Mocks) {
	logger := log.NewLogger()
	strg := storage.NewMemoryStorage()

	s := &Service{
		clients:     NewExternalServicesClients(),
		storage:     strg,
		logger:      logger,
		cancelFuncs: map[string]context.CancelFunc{},
	}

	for id, mockClient := range mockTelegram {
		s.clients.telegram[id] = &mockClient
	}

	m := &Mocks{
		storage:      strg,
		mockTelegram: mockTelegram,
	}
	return s, m
}
