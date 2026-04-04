package telegram_service

import (
	ts "telegramservice/internal/service/telegram_service"
	desc "telegramservice/pb/go"
)

// GRPCServer ...
type GRPCServer struct {
	desc.UnimplementedTelegramServiceServer
	service *ts.Service
}

// NewGRPCServer ...
func NewGRPCServer(servise *ts.Service) *GRPCServer {
	return &GRPCServer{
		service: servise,
	}
}
