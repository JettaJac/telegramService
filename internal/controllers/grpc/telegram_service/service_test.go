package telegram_service

import (
	service "telegramservice/internal/service/telegram_service"
	"telegramservice/pkg/logger"
	"time"

	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type Mocks struct {
	storage      *service.MockStorageInterface
	mockTelegram service.MockTelegramClientInterface
	logger       *zap.Logger
}

func NewMocks(
	ctrl *gomock.Controller,
	mockTelegram service.MockTelegramClientInterface,
) (*GRPCServer, *Mocks) {
	log := logger.NewLogger()
	strg := service.NewMockStorageInterface(ctrl)
	m := &Mocks{
		storage:      strg,
		mockTelegram: mockTelegram,
		logger:       log,
	}

	s := service.NewService(1, "", "../../../sessions", m.storage, 1*time.Second)

	app := NewGRPCServer(s)
	return app, m
}
