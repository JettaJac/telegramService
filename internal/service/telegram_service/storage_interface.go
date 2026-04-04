package telegram_service

import (
	"telegramservice/internal/storage"
)

//go:generate ../../../bin/mockgen -typed -source storage_interface.go -destination storage_interface_mock.go -package=telegram_service

type StorageInterface interface {
	SaveConnection(conn *storage.ConnectionInfo)
	GetConnection(id string) (*storage.ConnectionInfo, bool)
	DeleteConnection(id string)
	ListConnections() []*storage.ConnectionInfo
	UpdateConnectionStatus(id, status string)
	AddMessage(msg *storage.Message)
	GetMessages(connectionID string, limit int) []*storage.Message
}
