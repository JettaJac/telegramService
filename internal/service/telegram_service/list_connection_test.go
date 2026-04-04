package telegram_service

import (
	"telegramservice/internal/storage"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestSessionManager_ListConnections(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	strg := storage.NewMemoryStorage()

	// Добавляем тестовые соединения
	strg.SaveConnection(&storage.ConnectionInfo{
		ID:        "conn1",
		Status:    "pending",
		CreatedAt: time.Now().Unix(),
	})
	strg.SaveConnection(&storage.ConnectionInfo{
		ID:        "conn2",
		Status:    "authorized",
		CreatedAt: time.Now().Unix(),
	})

	testCases := map[string]struct {
		expectedCount int
	}{
		"list connections": {
			expectedCount: 2,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			sm := &Service{
				storage: strg,
				logger:  logger,
			}

			connections := sm.ListConnections()
			if len(connections) != tt.expectedCount {
				t.Errorf("expected %d connections, got %d", tt.expectedCount, len(connections))
			}
		})
	}
}
