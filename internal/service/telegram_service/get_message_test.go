package telegram_service

import (
	"context"
	"errors"
	"telegramservice/pkg/logger"
	"testing"
	"time"

	"telegramservice/internal/storage"

	"go.uber.org/mock/gomock"
)

func TestService_GetMessages(t *testing.T) {
	log := logger.NewLogger()
	strg := storage.NewMemoryStorage()

	testConnectionID := "test_conn_messages"

	testMessages := []*storage.Message{
		{
			ID:           "msg_001",
			ConnectionID: testConnectionID,
			SenderID:     "user_001",
			Text:         "First message",
			Timestamp:    time.Now().Unix(),
			ChatID:       "@user1",
		},
		{
			ID:           "msg_002",
			ConnectionID: testConnectionID,
			SenderID:     "user_002",
			Text:         "Second message",
			Timestamp:    time.Now().Unix(),
			ChatID:       "@user2",
		},
		{
			ID:           "msg_003",
			ConnectionID: testConnectionID,
			SenderID:     "user_001",
			Text:         "Third message",
			Timestamp:    time.Now().Unix(),
			ChatID:       "@user1",
		},
	}

	for _, msg := range testMessages {
		strg.AddMessage(msg)
	}

	testCases := map[string]struct {
		connectionID string
		limit        int
		expectedLen  int
		wantErr      bool
	}{
		"success get all messages": {
			connectionID: testConnectionID,
			limit:        0,
			expectedLen:  3,
			wantErr:      false,
		},
		"success get limited messages": {
			connectionID: testConnectionID,
			limit:        2,
			expectedLen:  2,
			wantErr:      false,
		},
		"connection with no messages": {
			connectionID: "empty_conn",
			limit:        10,
			expectedLen:  0,
			wantErr:      false,
		},
		"zero limit returns all": {
			connectionID: testConnectionID,
			limit:        0,
			expectedLen:  3,
			wantErr:      false,
		},
		"negative limit returns all": {
			connectionID: testConnectionID,
			limit:        -1,
			expectedLen:  3,
			wantErr:      false,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			s := &Service{
				clients:     NewExternalServicesClients(),
				cancelFuncs: make(map[string]context.CancelFunc),
				storage:     strg,
				logger:      log,
				apiID:       123456,
				apiHash:     "test_hash",
				sessionDir:  "./test_sessions",
				qrTimeout:   60 * time.Second,
			}

			messages := s.GetMessages(tt.connectionID, tt.limit)

			if len(messages) != tt.expectedLen {
				t.Errorf("expected %d messages, got %d", tt.expectedLen, len(messages))
			}

			if tt.expectedLen > 1 && tt.limit != 1 {
				for i := 1; i < len(messages); i++ {
					if messages[i].Timestamp < messages[i-1].Timestamp {
						t.Errorf("messages not in order: msg %d timestamp %d < msg %d timestamp %d",
							i, messages[i].Timestamp, i-1, messages[i-1].Timestamp)
					}
				}
			}
		})
	}
}

func TestService_SendMessageWithMock(t *testing.T) {
	log := logger.NewLogger()
	strg := storage.NewMemoryStorage()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testCases := map[string]struct {
		connectionID string
		chatID       string
		text         string
		setupMock    func(mock *MockTelegramClientInterface)
		wantErr      bool
		errMsg       string
	}{
		"success send message": {
			connectionID: "test_conn",
			chatID:       "@username",
			text:         "Hello!",
			setupMock: func(mock *MockTelegramClientInterface) {
				mock.EXPECT().IsAuthorized().Return(true)
				mock.EXPECT().SendMessage(gomock.Any(), "@username", "Hello!").Return("msg_123", nil)
			},
			wantErr: false,
		},
		"connection not found": {
			connectionID: "non_existent",
			chatID:       "@username",
			text:         "Hello",
			setupMock:    nil,
			wantErr:      true,
			errMsg:       "not found",
		},
		"connection not authorized": {
			connectionID: "unauth_conn",
			chatID:       "@username",
			text:         "Hello",
			setupMock: func(mock *MockTelegramClientInterface) {
				mock.EXPECT().IsAuthorized().Return(false)
			},
			wantErr: true,
			errMsg:  "not authorized",
		},
		"send message error": {
			connectionID: "error_conn",
			chatID:       "@username",
			text:         "Hello",
			setupMock: func(mock *MockTelegramClientInterface) {
				mock.EXPECT().IsAuthorized().Return(true)
				mock.EXPECT().SendMessage(gomock.Any(), "@username", "Hello").Return("", errors.New("send failed"))
			},
			wantErr: true,
			errMsg:  "send failed",
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			s := &Service{
				clients:     NewExternalServicesClients(),
				cancelFuncs: make(map[string]context.CancelFunc),
				storage:     strg,
				logger:      log,
				apiID:       123456,
				apiHash:     "test_hash",
				sessionDir:  "./test_sessions",
				qrTimeout:   60,
			}

			if tt.setupMock != nil {
				mockClient := NewMockTelegramClientInterface(ctrl)
				tt.setupMock(mockClient)
				s.clients.telegram[tt.connectionID] = mockClient
			}

			ctx := context.Background()
			_, err := s.SendMessage(ctx, tt.connectionID, tt.chatID, tt.text)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if tt.errMsg != "" && err != nil && !contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error to contain %q, got %q", tt.errMsg, err.Error())
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
