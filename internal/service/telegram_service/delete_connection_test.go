package telegram_service

import (
	"context"
	"errors"
	"fmt"
	"telegramservice/internal/storage"
	log "telegramservice/pkg/logger"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestService_DeleteConnection(t *testing.T) {
	testCases := map[string]struct {
		connectionID  string
		logout        bool
		setupMocks    func(*gomock.Controller) (*Service, *Mocks)
		expectedError string
		checkResult   func(*testing.T, *Service, string)
	}{
		"success delete with logout": {
			connectionID: "test_conn_to_delete",
			logout:       true,
			setupMocks: func(ctrl *gomock.Controller) (*Service, *Mocks) {
				mockTelegram := make(map[string]MockTelegramClientInterface)

				s, m := NewMocks(mockTelegram)

				mockClient := NewMockTelegramClientInterface(ctrl)
				mockClient.EXPECT().
					Stop(gomock.Any(), true).
					Return(nil)

				s.clients.telegram["test_conn_to_delete"] = mockClient
				m.mockTelegram["test_conn_to_delete"] = *mockClient

				// Добавляем cancel функцию
				cancelFunc := func() {}
				s.cancelFuncs["test_conn_to_delete"] = cancelFunc

				return s, m
			},
			expectedError: "",
			checkResult: func(t *testing.T, s *Service, connID string) {
				// Проверяем, что клиент удален
				_, exists := s.clients.telegram[connID]
				assert.False(t, exists, "client should be removed from map")

				// Проверяем, что cancelFunc удалена
				_, exists = s.cancelFuncs[connID]
				assert.False(t, exists, "cancelFunc should be removed")
			},
		},

		"success delete without logout": {
			connectionID: "test_conn_to_delete",
			logout:       false,
			setupMocks: func(ctrl *gomock.Controller) (*Service, *Mocks) {
				mockTelegram := make(map[string]MockTelegramClientInterface)

				s, m := NewMocks(mockTelegram)

				mockClient := NewMockTelegramClientInterface(ctrl)
				mockClient.EXPECT().
					Stop(gomock.Any(), false).
					Return(nil)

				s.clients.telegram["test_conn_to_delete"] = mockClient
				m.mockTelegram["test_conn_to_delete"] = *mockClient

				cancelFunc := func() {}
				s.cancelFuncs["test_conn_to_delete"] = cancelFunc

				return s, m
			},
			expectedError: "",
			checkResult: func(t *testing.T, s *Service, connID string) {
				_, exists := s.clients.telegram[connID]
				assert.False(t, exists)
			},
		},

		"connection not found": {
			connectionID: "non_existent",
			logout:       false,
			setupMocks: func(ctrl *gomock.Controller) (*Service, *Mocks) {
				mockTelegram := make(map[string]MockTelegramClientInterface)

				s, m := NewMocks(mockTelegram)

				return s, m
			},
			expectedError: "connection non_existent not found",
			checkResult:   nil,
		},

		"stop client error": {
			connectionID: "test_conn_to_delete",
			logout:       false,
			setupMocks: func(ctrl *gomock.Controller) (*Service, *Mocks) {
				mockTelegram := make(map[string]MockTelegramClientInterface)

				s, m := NewMocks(mockTelegram)

				mockClient := NewMockTelegramClientInterface(ctrl)
				mockClient.EXPECT().
					Stop(gomock.Any(), false).
					Return(fmt.Errorf("stop failed"))

				s.clients.telegram["test_conn_to_delete"] = mockClient
				m.mockTelegram["test_conn_to_delete"] = *mockClient

				return s, m
			},
			expectedError: "failed to stop client: stop failed",
			checkResult: func(t *testing.T, s *Service, connID string) {
				// При ошибке клиент НЕ должен удаляться
				_, exists := s.clients.telegram[connID]
				assert.True(t, exists, "client should still exist after error")

				_, exists = s.cancelFuncs[connID]
				assert.False(t, exists, "cancelFunc should be removed even on error")
			},
		},

		"with cancel context": {
			connectionID: "test_conn_to_delete",
			logout:       true,
			setupMocks: func(ctrl *gomock.Controller) (*Service, *Mocks) {
				//mockStorage := NewMockStorageInterface(ctrl)
				mockTelegram := make(map[string]MockTelegramClientInterface)
				//mockStorage.EXPECT().
				//	DeleteConnection("test_conn_to_delete").
				//	Return()

				s, m := NewMocks(mockTelegram)

				mockClient := NewMockTelegramClientInterface(ctrl)
				mockClient.EXPECT().
					Stop(gomock.Any(), true).
					Return(nil)

				s.clients.telegram["test_conn_to_delete"] = mockClient
				m.mockTelegram["test_conn_to_delete"] = *mockClient

				// Возвращаем cancelFunc для проверки
				return s, m
			},
			expectedError: "",
			checkResult: func(t *testing.T, s *Service, connID string) {
				_, exists := s.clients.telegram[connID]
				assert.False(t, exists)
			},
		},

		"empty connection id": {
			connectionID: "",
			logout:       false,
			setupMocks: func(ctrl *gomock.Controller) (*Service, *Mocks) {
				mockTelegram := make(map[string]MockTelegramClientInterface)
				s, m := NewMocks(mockTelegram)

				return s, m
			},
			expectedError: "connection  not found",
			checkResult:   nil,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s, _ := tt.setupMocks(ctrl)

			ctx := context.Background()
			err := s.DeleteConnection(ctx, tt.connectionID, tt.logout)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			// Дополнительные проверки
			if tt.checkResult != nil {
				tt.checkResult(t, s, tt.connectionID)
			}
		})
	}
}

func TestService_DeleteConnectionWithMock(t *testing.T) {
	logger := log.NewLogger()
	strg := storage.NewMemoryStorage()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testCases := map[string]struct {
		connectionID string
		logout       bool
		setupMock    func(mock *MockTelegramClientInterface)
		wantErr      bool
		errMsg       string
	}{
		"success delete": {
			connectionID: "test_conn",
			logout:       true,
			setupMock: func(mock *MockTelegramClientInterface) {
				mock.EXPECT().Stop(gomock.Any(), true).Return(nil)
			},
			wantErr: false,
		},
		"connection not found": {
			connectionID: "non_existent",
			logout:       false,
			setupMock:    nil,
			wantErr:      true,
			errMsg:       "not found",
		},
		"stop error": {
			connectionID: "error_conn",
			logout:       false,
			setupMock: func(mock *MockTelegramClientInterface) {
				mock.EXPECT().Stop(gomock.Any(), false).Return(errors.New("stop failed"))
			},
			wantErr: true,
			errMsg:  "stop failed",
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			s := &Service{
				clients:     NewExternalServicesClients(),
				cancelFuncs: make(map[string]context.CancelFunc),
				storage:     strg,
				logger:      logger,
				apiID:       123456,
				apiHash:     "test_hash",
				sessionDir:  "./test_sessions",
				qrTimeout:   60,
			}

			if tt.setupMock != nil {
				mockClient := NewMockTelegramClientInterface(ctrl)
				tt.setupMock(mockClient)
				s.clients.telegram[tt.connectionID] = mockClient
				s.cancelFuncs[tt.connectionID] = func() {}
			}

			ctx := context.Background()
			err := s.DeleteConnection(ctx, tt.connectionID, tt.logout)

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
