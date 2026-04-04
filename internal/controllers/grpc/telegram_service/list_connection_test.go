package telegram_service

import (
	"context"
	service "telegramservice/internal/service/telegram_service"
	"telegramservice/internal/storage"
	desc "telegramservice/pb/go"
	"testing"

	"go.uber.org/mock/gomock"
)

func TestGRPCServer_ListConnections(t *testing.T) {
	testCases := map[string]struct {
		setupMock   func(mockStorage *service.MockStorageInterface)
		expectedLen int
	}{
		"success with connections": {
			setupMock: func(mockStorage *service.MockStorageInterface) {
				mockStorage.EXPECT().ListConnections().Return([]*storage.ConnectionInfo{
					{ID: "conn1", Status: "pending"},
					{ID: "conn2", Status: "authorized"},
				})
			},
			expectedLen: 2,
		},
		"empty list": {
			setupMock: func(mockStorage *service.MockStorageInterface) {
				mockStorage.EXPECT().ListConnections().Return([]*storage.ConnectionInfo{})
			},
			expectedLen: 0,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTelegramClient := service.NewMockTelegramClientInterface(ctrl)
			server, m := NewMocks(ctrl, *mockTelegramClient)

			if tt.setupMock != nil {
				tt.setupMock(m.storage)
			}

			resp, err := server.ListConnections(context.Background(), &desc.ListConnectionsRequest{})

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if len(resp.Connections) != tt.expectedLen {
				t.Errorf("expected %d connections, got %d", tt.expectedLen, len(resp.Connections))
			}
		})
	}
}
