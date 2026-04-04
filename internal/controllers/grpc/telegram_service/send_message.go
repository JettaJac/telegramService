package telegram_service

import (
	"context"
	desc "telegramservice/pb/go"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SendMessage ...
func (s *GRPCServer) SendMessage(ctx context.Context, req *desc.SendMessageRequest) (*desc.SendMessageResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if err := s.validateSendMessageRequest(req); err != nil {
		return nil, err
	}

	// Передаем контекст в сервис
	msgID, err := s.service.SendMessage(ctx, req.ConnectionId, req.ChatId, req.Text)
	if err != nil {
		return &desc.SendMessageResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &desc.SendMessageResponse{
		MessageId: msgID,
		Success:   true,
	}, nil
}

// validateSendMessageRequest валидирует запрос на отправку сообщения
func (s *GRPCServer) validateSendMessageRequest(req *desc.SendMessageRequest) error {
	if req.ConnectionId == "" {
		return status.Error(codes.InvalidArgument, "connection_id is required")
	}
	if req.ChatId == "" {
		return status.Error(codes.InvalidArgument, "chat_id is required")
	}
	if req.Text == "" {
		return status.Error(codes.InvalidArgument, "text is required")
	}
	return nil
}
