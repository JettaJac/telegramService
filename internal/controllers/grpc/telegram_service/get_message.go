package telegram_service

import (
	desc "telegramservice/pb/go"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetMessages ...
func (s *GRPCServer) GetMessages(req *desc.GetMessagesRequest, stream desc.TelegramService_GetMessagesServer) error {
	ctx := stream.Context()
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if req.ConnectionId == "" {
		return status.Error(codes.InvalidArgument, "connection_id is required")
	}

	messages := s.service.GetMessages(req.ConnectionId, int(req.Limit))

	for _, msg := range messages {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		pbMsg := &desc.Message{
			Id:           msg.ID,
			ConnectionId: msg.ConnectionID,
			SenderId:     msg.SenderID,
			Text:         msg.Text,
			Timestamp:    msg.Timestamp,
			ChatId:       msg.ChatID,
		}

		if err := stream.Send(pbMsg); err != nil {
			return err
		}
	}

	return nil
}
