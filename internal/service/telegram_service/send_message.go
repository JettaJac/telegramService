package telegram_service

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// SendMessage ...
func (s *Service) SendMessage(ctx context.Context, connectionID, chatID, text string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	s.mu.RLock()
	client, exists := s.clients.telegram[connectionID]
	s.mu.RUnlock()

	if !exists {
		s.logger.Warn("connection not found for sending message",
			zap.String("connection_id", connectionID))
		return "", fmt.Errorf("connection %s not found", connectionID)
	}

	if !client.IsAuthorized() {
		s.logger.Warn("connection not authorized",
			zap.String("connection_id", connectionID))
		return "", fmt.Errorf("connection %s not authorized", connectionID)
	}

	s.logger.Info("sending message via connection",
		zap.String("connection_id", connectionID),
		zap.String("chat_id", chatID))

	return client.SendMessage(ctx, chatID, text)
}
