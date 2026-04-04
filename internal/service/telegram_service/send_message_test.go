package telegram_service

import (
	"context"
	"errors"
	"sync"
	"telegramservice/pkg/logger"
	"testing"

	"go.uber.org/mock/gomock"
)

func TestService_SendMessage(t *testing.T) {
	testCases := map[string]struct {
		connectionID  string
		chatID        string
		text          string
		setupMock     func(mockTelegram *MockTelegramClientInterface, mockStorage *MockStorageInterface)
		wantErr       bool
		expectedErr   string
		expectedMsgID string
	}{
		"success": {
			connectionID: "test_conn",
			chatID:       "@username",
			text:         "Hello, World!",
			setupMock: func(mockTelegram *MockTelegramClientInterface, mockStorage *MockStorageInterface) {
				mockTelegram.EXPECT().IsAuthorized().Return(true)
				mockTelegram.EXPECT().SendMessage(gomock.Any(), "@username", "Hello, World!").Return("msg_12345", nil)
			},
			wantErr:       false,
			expectedMsgID: "msg_12345",
		},
		"connection not found": {
			connectionID: "non_existent",
			chatID:       "@username",
			text:         "Hello",
			setupMock:    func(mockTelegram *MockTelegramClientInterface, mockStorage *MockStorageInterface) {},
			wantErr:      true,
			expectedErr:  "connection non_existent not found",
		},
		"connection not authorized": {
			connectionID: "unauth_conn",
			chatID:       "@username",
			text:         "Hello",
			setupMock: func(mockTelegram *MockTelegramClientInterface, mockStorage *MockStorageInterface) {
				mockTelegram.EXPECT().IsAuthorized().Return(false)
			},
			wantErr:     true,
			expectedErr: "connection unauth_conn not authorized",
		},
		"send message error": {
			connectionID: "error_conn",
			chatID:       "@username",
			text:         "Hello",
			setupMock: func(mockTelegram *MockTelegramClientInterface, mockStorage *MockStorageInterface) {
				mockTelegram.EXPECT().IsAuthorized().Return(true)
				mockTelegram.EXPECT().SendMessage(gomock.Any(), "@username", "Hello").Return("", errors.New("telegram send failed"))
			},
			wantErr:     true,
			expectedErr: "telegram send failed",
		},
		"context cancelled": {
			connectionID: "test_conn",
			chatID:       "@username",
			text:         "Hello",
			setupMock: func(mockTelegram *MockTelegramClientInterface, mockStorage *MockStorageInterface) {
			},
			wantErr:     true,
			expectedErr: "context canceled",
		},
		"empty connection id": {
			connectionID: "",
			chatID:       "@username",
			text:         "Hello",
			setupMock:    func(mockTelegram *MockTelegramClientInterface, mockStorage *MockStorageInterface) {},
			wantErr:      true,
			expectedErr:  "connection  not found",
		},
		"empty chat id": {
			connectionID: "test_conn",
			chatID:       "",
			text:         "Hello",
			setupMock: func(mockTelegram *MockTelegramClientInterface, mockStorage *MockStorageInterface) {
				mockTelegram.EXPECT().IsAuthorized().Return(true)
				mockTelegram.EXPECT().SendMessage(gomock.Any(), "", "Hello").Return("msg_123", nil)
			},
			wantErr:       false,
			expectedMsgID: "msg_123",
		},
		"empty text": {
			connectionID: "test_conn",
			chatID:       "@username",
			text:         "",
			setupMock: func(mockTelegram *MockTelegramClientInterface, mockStorage *MockStorageInterface) {
				mockTelegram.EXPECT().IsAuthorized().Return(true)
				mockTelegram.EXPECT().SendMessage(gomock.Any(), "@username", "").Return("msg_123", nil)
			},
			wantErr:       false,
			expectedMsgID: "msg_123",
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTelegram := NewMockTelegramClientInterface(ctrl)
			mockStorage := NewMockStorageInterface(ctrl)

			if tt.setupMock != nil {
				tt.setupMock(mockTelegram, mockStorage)
			}

			service := &Service{
				clients: &ExternalServicesClients{
					telegram: make(map[string]TelegramClientInterface),
				},
				storage: mockStorage,
				logger:  logger.NewLogger(),
				mu:      sync.RWMutex{},
			}

			if tt.connectionID != "" && tt.connectionID != "non_existent" {
				service.clients.telegram[tt.connectionID] = mockTelegram
			}

			ctx := context.Background()

			if name == "context cancelled" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			msgID, err := service.SendMessage(ctx, tt.connectionID, tt.chatID, tt.text)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if tt.expectedErr != "" && err.Error() != tt.expectedErr {
					t.Errorf("expected error %q, got %q", tt.expectedErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if msgID != tt.expectedMsgID {
				t.Errorf("expected msgID %q, got %q", tt.expectedMsgID, msgID)
			}
		})
	}
}
