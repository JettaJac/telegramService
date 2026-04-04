package telegram_service

import "telegramservice/internal/storage"

// ListConnections ...
func (s *Service) ListConnections() []*storage.ConnectionInfo {
	return s.storage.ListConnections()
}
