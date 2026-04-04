package telegram_service

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// DeleteConnection ...
func (s *Service) DeleteConnection(ctx context.Context, connectionID string, logout bool) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("deleting connection",
		zap.String("connection_id", connectionID),
		zap.Bool("logout", logout))

	client, exists := s.clients.telegram[connectionID]
	if !exists {
		s.logger.Warn("connection not found",
			zap.String("connection_id", connectionID))
		return fmt.Errorf("connection %s not found", connectionID)
	}

	// Отменяем контекст клиента
	if cancel, ok := s.cancelFuncs[connectionID]; ok {
		cancel()
		delete(s.cancelFuncs, connectionID)
	}
	if err := client.Stop(ctx, logout); err != nil {
		s.logger.Error("failed to stop client",
			zap.String("connection_id", connectionID),
			zap.Error(err))
		return fmt.Errorf("failed to stop client: %w", err)
	}

	delete(s.clients.telegram, connectionID)
	s.storage.DeleteConnection(connectionID)

	s.logger.Info("connection deleted",
		zap.String("connection_id", connectionID))

	return nil
}
