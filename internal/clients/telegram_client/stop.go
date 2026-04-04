package telegram_client

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/zap"
)

// Stop разрываем соединение
func (c *TelegramClient) Stop(ctx context.Context, logout bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger.Info("stopping client",
		zap.Bool("logout", logout),
		zap.String("client_id", c.id))

	if c.cancelFunc != nil {
		c.cancelFunc()
	}

	if logout {
		sessionPath := fmt.Sprintf("%s/%s.session", c.sessionDir, c.id)
		c.logger.Info("removing session file",
			zap.String("path", sessionPath),
			zap.String("client_id", c.id))

		if err := os.Remove(sessionPath); err != nil && !os.IsNotExist(err) {
			c.logger.Error("failed to remove session file",
				zap.String("client_id", c.id),
				zap.String("path", sessionPath),
				zap.Error(err))
		}
	}
	// Синхронизируем логгер
	if c.logger != nil {
		_ = c.logger.Sync()
	}

	c.authorized = false
	c.logger.Info("client stopped",
		zap.String("client_id", c.id))

	return nil
}
