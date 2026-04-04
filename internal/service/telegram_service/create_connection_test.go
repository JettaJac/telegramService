package telegram_service

import (
	"context"
	"telegramservice/internal/storage"
	"telegramservice/pkg/logger"
	"testing"

	"github.com/gotd/td/tg"
	"go.uber.org/mock/gomock"
)

func TestService_CreateConnection(t *testing.T) {
	log := logger.NewLogger()
	strg := storage.NewMemoryStorage()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testCases := map[string]struct {
		connectionID string
		setupMock    func(mock *MockTelegramClientInterface)
		wantErr      bool
		errMsg       string
	}{
		"success create connection": {
			connectionID: "test_conn",
			setupMock: func(mock *MockTelegramClientInterface) {
				mock.EXPECT().IsAuthorized().Return(false).AnyTimes()
			},
			wantErr: false,
		},
		"duplicate connection": {
			connectionID: "test_conn",
			setupMock: func(mock *MockTelegramClientInterface) {
				mock.EXPECT().IsAuthorized().Return(true).AnyTimes()
			},
			wantErr: true,
			errMsg:  "already exists",
		},
		"empty connection id": {
			connectionID: "",
			setupMock:    func(mock *MockTelegramClientInterface) {},
			wantErr:      true,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			mockClient := NewMockTelegramClientInterface(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockClient)
			}

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

			if tt.connectionID == "test_conn" && tt.wantErr {
				s.clients.telegram["test_conn"] = mockClient
			}

			ctx := context.Background()
			_, err := s.CreateConnection(ctx, tt.connectionID)

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

func TestService_validateConnectionID(t *testing.T) {
	log := logger.NewLogger()
	strg := storage.NewMemoryStorage()

	testCases := map[string]struct {
		connectionID string
		wantErr      bool
	}{
		"valid connection id": {
			connectionID: "test_conn",
			wantErr:      false,
		},
		"empty connection id": {
			connectionID: "",
			wantErr:      true,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			s := &Service{
				storage: strg,
				logger:  log,
			}
			err := s.validateConnectionID(tt.connectionID)
			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestService_checkConnectionExists(t *testing.T) {
	log := logger.NewLogger()
	strg := storage.NewMemoryStorage()

	testCases := map[string]struct {
		connectionID string
		setup        func(s *Service)
		wantExists   bool
		wantErr      bool
	}{
		//"connection exists and authorized": {
		//	connectionID: "existing_conn",
		//	setup: func(s *Service) {
		//		mockClient := &MockTelegramClientInterface{
		//			IsAuthorizedFn: func() bool { return true },
		//			IsAuthorized: func() bool {}
		//			IsAuthorizedFunc: func() bool { return true },
		//		}
		//		s.clients.telegram["existing_conn"] = mockClient
		//	},
		//	wantExists: true,
		//	wantErr:    true,
		//},
		"connection does not exist": {
			connectionID: "non_existent",
			setup:        func(s *Service) {},
			wantExists:   false,
			wantErr:      false,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			s := &Service{
				clients: NewExternalServicesClients(),
				storage: strg,
				logger:  log,
			}
			if tt.setup != nil {
				tt.setup(s)
			}

			exists, err := s.checkConnectionExists(tt.connectionID)

			if exists != tt.wantExists {
				t.Errorf("expected exists=%v, got %v", tt.wantExists, exists)
			}
			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestService_createQRReadyCallback(t *testing.T) {
	log := logger.NewLogger()
	strg := storage.NewMemoryStorage()

	testCases := map[string]struct {
		connectionID string
		qrData       string
	}{
		"valid qr data": {
			connectionID: "test_conn",
			qrData:       "tg://login?token=123",
		},
		"empty qr data": {
			connectionID: "test_conn",
			qrData:       "",
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			s := &Service{
				storage: strg,
				logger:  log,
			}
			var qrData string
			callback := s.createQRReadyCallback(tt.connectionID, &qrData)

			callback(tt.qrData)

			if tt.qrData != "" && qrData != tt.qrData {
				t.Errorf("expected qrData=%q, got %q", tt.qrData, qrData)
			}
		})
	}
}

func TestService_createMessageHandler(t *testing.T) {
	log := logger.NewLogger()
	strg := storage.NewMemoryStorage()

	s := &Service{
		storage: strg,
		logger:  log,
	}

	handler := s.createMessageHandler("test_conn")
	if handler == nil {
		t.Errorf("expected non-nil handler")
	}
}

func TestService_createTelegramClient(t *testing.T) {
	log := logger.NewLogger()
	strg := storage.NewMemoryStorage()

	testCases := map[string]struct {
		connectionID string
		apiID        int
		apiHash      string
		sessionDir   string
	}{
		"valid client": {
			connectionID: "test_conn",
			apiID:        123456,
			apiHash:      "test_hash",
			sessionDir:   "./sessions",
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			s := &Service{
				storage:    strg,
				logger:     log,
				apiID:      tt.apiID,
				apiHash:    tt.apiHash,
				sessionDir: tt.sessionDir,
			}

			tgClient := s.createTelegramClient(tt.connectionID)
			if tgClient == nil {
				t.Errorf("expected non-nil client")
			}
		})
	}
}

func TestService_extractSenderID(t *testing.T) {
	log := logger.NewLogger()
	strg := storage.NewMemoryStorage()

	testCases := map[string]struct {
		msg      *tg.Message
		expected string
	}{
		"user peer": {
			msg:      &tg.Message{PeerID: &tg.PeerUser{UserID: 123456}},
			expected: "123456",
		},
		"chat peer": {
			msg:      &tg.Message{PeerID: &tg.PeerChat{ChatID: 789012}},
			expected: "chat_789012",
		},
		"channel peer": {
			msg:      &tg.Message{PeerID: &tg.PeerChannel{ChannelID: 345678}},
			expected: "channel_345678",
		},
		"nil peer": {
			msg:      &tg.Message{},
			expected: "unknown",
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			s := &Service{
				storage: strg,
				logger:  log,
			}
			result := s.extractSenderID(tt.msg)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestService_extractChatID(t *testing.T) {
	log := logger.NewLogger()
	strg := storage.NewMemoryStorage()

	testCases := map[string]struct {
		msg      *tg.Message
		expected string
	}{
		"user peer": {
			msg:      &tg.Message{PeerID: &tg.PeerUser{UserID: 123456}},
			expected: "123456",
		},
		"chat peer": {
			msg:      &tg.Message{PeerID: &tg.PeerChat{ChatID: 789012}},
			expected: "789012",
		},
		"channel peer": {
			msg:      &tg.Message{PeerID: &tg.PeerChannel{ChannelID: 345678}},
			expected: "345678",
		},
		"nil peer": {
			msg:      &tg.Message{},
			expected: "unknown",
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			s := &Service{
				storage: strg,
				logger:  log,
			}
			result := s.extractChatID(tt.msg)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestService_saveConnectionInfo(t *testing.T) {
	log := logger.NewLogger()
	strg := storage.NewMemoryStorage()

	testCases := map[string]struct {
		connectionID string
	}{
		"save connection": {
			connectionID: "test_conn",
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			s := &Service{
				storage: strg,
				logger:  log,
			}
			s.saveConnectionInfo(tt.connectionID)

			conn, exists := strg.GetConnection(tt.connectionID)
			if !exists {
				t.Errorf("connection not saved")
			}
			if conn.Status != "pending" {
				t.Errorf("expected status 'pending', got %q", conn.Status)
			}
		})
	}
}

func TestService_generateID(t *testing.T) {
	id1 := generateID()
	id2 := generateID()

	if id1 == "" {
		t.Errorf("expected non-empty ID")
	}
	if id1 == id2 {
		t.Errorf("expected different IDs, got same: %s", id1)
	}
}

//func TestService_GetMessages(t *testing.T) {
//	log := logger.NewLogger()
// strg := storage.NewMemoryStorage()
//
//	strg.AddMessage(&storage.Message{ID: "msg1", ConnectionID: "test_conn", Text: "Hello"})
//	strg.AddMessage(&storage.Message{ID: "msg2", ConnectionID: "test_conn", Text: "World"})
//
//	testCases := map[string]struct {
//		connectionID string
//		limit        int
//		expectedLen  int
//	}{
//		"get all messages": {
//			connectionID: "test_conn",
//			limit:        0,
//			expectedLen:  2,
//		},
//		"get limited messages": {
//			connectionID: "test_conn",
//			limit:        1,
//			expectedLen:  1,
//		},
//		"empty connection": {
//			connectionID: "empty_conn",
//			limit:        10,
//			expectedLen:  0,
//		},
//	}
//
//	for name, tt := range testCases {
//		t.Run(name, func(t *testing.T) {
//			s := &Service{
//				storage: strg,
//				logger:  log,
//			}
//			messages := s.GetMessages(tt.connectionID, tt.limit)
//
//			if len(messages) != tt.expectedLen {
//				t.Errorf("expected %d messages, got %d", tt.expectedLen, len(messages))
//			}
//		})
//	}
//}

//func TestService_ListConnections(t *testing.T) {
//	log := logger.NewLogger()
// strg := storage.NewMemoryStorage()
//
//	storage.SaveConnection(&storage.ConnectionInfo{ID: "conn1", Status: "pending"})
//	storage.SaveConnection(&storage.ConnectionInfo{ID: "conn2", Status: "authorized"})
//
//	s := &Service{
//		storage: strg,
//		logger:  log,
//	}
//
//	connections := s.ListConnections()
//
//	if len(connections) != 2 {
//		t.Errorf("expected 2 connections, got %d", len(connections))
//	}
//}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > 0 && len(substr) > 0 && (s[0:len(substr)] == substr ||
			(len(s) > len(substr) && contains(s[1:], substr)))))
}
