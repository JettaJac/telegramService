package telegram_service

import "telegramservice/internal/storage"

// GetMessages ...
func (s *Service) GetMessages(connectionID string, limit int) []*storage.Message {
	return s.storage.GetMessages(connectionID, limit)
}
