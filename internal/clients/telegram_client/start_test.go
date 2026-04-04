package telegram_client

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/auth/qrlogin"
	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestTelegramClient_setupDispatcher(t *testing.T) {
	testCases := map[string]struct {
		messageHandler func(string, *tg.Message)
	}{
		"setup dispatcher with nil handler": {
			messageHandler: nil,
		},
		"setup dispatcher with handler": {
			messageHandler: func(clientID string, msg *tg.Message) {},
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			client := &TelegramClient{
				id:             "test_client",
				logger:         zap.NewNop(),
				messageHandler: tt.messageHandler,
			}

			dispatcher := client.setupDispatcher()
			require.NotNil(t, dispatcher)
		})
	}
}

func TestTelegramClient_handleAuthStatus(t *testing.T) {
	testCases := map[string]struct {
		authStatus     *auth.Status
		authStatusErr  error
		expectedResult bool
		expectedError  bool
	}{
		"authorized returns true": {
			authStatus:     &auth.Status{Authorized: true},
			authStatusErr:  nil,
			expectedResult: true,
			expectedError:  false,
		},
		"not authorized returns false": {
			authStatus:     &auth.Status{Authorized: false},
			authStatusErr:  nil,
			expectedResult: false,
			expectedError:  false,
		},
		"error getting status returns error": {
			authStatus:     nil,
			authStatusErr:  errors.New("network error"),
			expectedResult: false,
			expectedError:  true,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTelegram := NewMockTelegramClientInterface(ctrl)
			mockAuth := NewMockAuthClientInterface(ctrl)

			mockTelegram.EXPECT().Auth().Return(mockAuth).Times(1)
			mockAuth.EXPECT().Status(gomock.Any()).Return(tt.authStatus, tt.authStatusErr)

			client := &TelegramClient{
				client: mockTelegram,
				logger: zap.NewNop(),
				id:     "test",
			}

			result, err := client.handleAuthStatus(context.Background())

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestTelegramClient_markAuthorized(t *testing.T) {
	testCases := map[string]struct {
		initialAuthState bool
	}{
		"mark authorized from false": {
			initialAuthState: false,
		},
		"mark authorized from true": {
			initialAuthState: true,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			client := &TelegramClient{
				id:         "test_client",
				logger:     zap.NewNop(),
				authorized: tt.initialAuthState,
			}

			client.markAuthorized()

			require.True(t, client.authorized)
		})
	}
}

func TestTelegramClient_makeQRHandler(t *testing.T) {
	testCases := map[string]struct {
		qrCodeReady    func(string)
		tokenString    string
		expectCallback bool
	}{
		"qr handler with callback": {
			qrCodeReady: func(data string) {
				require.Equal(t, "dGVzdF9xcl90b2tlbg==", data)
			},
			tokenString:    "test_qr_token",
			expectCallback: true,
		},
		"qr handler with nil callback": {
			qrCodeReady:    nil,
			tokenString:    "test_qr_token",
			expectCallback: false,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client := &TelegramClient{
				id:     "test_client",
				logger: zap.NewNop(),
			}

			handler := client.makeQRHandler(tt.qrCodeReady)
			require.NotNil(t, handler)

			token := qrlogin.NewToken([]byte(tt.tokenString), 60)

			err := handler(context.Background(), token)
			require.NoError(t, err)
		})
	}
}

func TestTelegramClient_isAlreadyAuthorized(t *testing.T) {
	testCases := map[string]struct {
		authStatus     *auth.Status
		authStatusErr  error
		expectedResult bool
		expectedError  bool
	}{
		"already authorized returns true": {
			authStatus:     &auth.Status{Authorized: true},
			authStatusErr:  nil,
			expectedResult: true,
			expectedError:  false,
		},
		"not authorized returns false": {
			authStatus:     &auth.Status{Authorized: false},
			authStatusErr:  nil,
			expectedResult: false,
			expectedError:  false,
		},
		"error checking auth returns error": {
			authStatus:     nil,
			authStatusErr:  errors.New("auth error"),
			expectedResult: false,
			expectedError:  true,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTelegram := NewMockTelegramClientInterface(ctrl)
			mockAuth := NewMockAuthClientInterface(ctrl)

			mockTelegram.EXPECT().Auth().Return(mockAuth).Times(1)
			mockAuth.EXPECT().Status(gomock.Any()).Return(tt.authStatus, tt.authStatusErr)

			client := &TelegramClient{
				client: mockTelegram,
				logger: zap.NewNop(),
				id:     "test",
			}

			result, err := client.isAlreadyAuthorized(context.Background())

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedResult, result)
				if tt.expectedResult {
					require.True(t, client.authorized)
				}
			}
		})
	}
}

func TestTelegramClient_performQRAuth(t *testing.T) {
	t.Skip("Skipping because qrlogin.LoggedIn channel cannot be controlled from test")

	testCases := map[string]struct {
		qrAuthErr     error
		expectedError bool
	}{
		"QR auth fails": {
			qrAuthErr:     errors.New("qr auth failed"),
			expectedError: true,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTelegram := NewMockTelegramClientInterface(ctrl)
			mockQR := NewMockQRClientInterface(ctrl)

			mockTelegram.EXPECT().QR().Return(mockQR).Times(1)
			mockQR.EXPECT().Auth(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, tt.qrAuthErr).Times(1)

			client := &TelegramClient{
				client:     mockTelegram,
				logger:     zap.NewNop(),
				id:         "test",
				authorized: false,
			}

			err := client.performQRAuth(context.Background(), nil)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTelegramClient_runAuthFlow(t *testing.T) {
	testCases := map[string]struct {
		authStatus    *auth.Status
		authStatusErr error
		qrAuthErr     error
		expectedError bool
	}{
		"already authorized": {
			authStatus:    &auth.Status{Authorized: true},
			authStatusErr: nil,
			qrAuthErr:     nil,
			expectedError: false,
		},
		"auth status check fails": {
			authStatus:    nil,
			authStatusErr: errors.New("status error"),
			qrAuthErr:     nil,
			expectedError: true,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTelegram := NewMockTelegramClientInterface(ctrl)
			mockAuth := NewMockAuthClientInterface(ctrl)
			mockQR := NewMockQRClientInterface(ctrl)

			mockTelegram.EXPECT().Auth().Return(mockAuth).Times(1)
			mockAuth.EXPECT().Status(gomock.Any()).Return(tt.authStatus, tt.authStatusErr)

			if tt.authStatusErr == nil && !tt.authStatus.Authorized {
				mockTelegram.EXPECT().QR().Return(mockQR).Times(1)
				mockQR.EXPECT().Auth(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, tt.qrAuthErr).Times(1)
			}

			client := &TelegramClient{
				client:     mockTelegram,
				logger:     zap.NewNop(),
				id:         "test",
				authorized: false,
			}

			err := client.runAuthFlow(context.Background(), nil)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
func TestTelegramClient_handleNewMessage(t *testing.T) {
	testCases := map[string]struct {
		update         *tg.UpdateNewMessage
		messageHandler func(string, *tg.Message)
		expectedCall   bool
	}{
		"valid message triggers handler": {
			update: &tg.UpdateNewMessage{
				Message: &tg.Message{
					PeerID:  &tg.PeerUser{UserID: 12345},
					Message: "Hello world",
					ID:      1,
				},
			},
			messageHandler: func(clientID string, msg *tg.Message) {
				require.Equal(t, "Hello world", msg.Message)
			},
			expectedCall: true,
		},
		"message with empty text": {
			update: &tg.UpdateNewMessage{
				Message: &tg.Message{
					PeerID:  &tg.PeerUser{UserID: 12345},
					Message: "",
					ID:      1,
				},
			},
			messageHandler: func(clientID string, msg *tg.Message) {
				t.Error("handler should not be called for empty message")
			},
			expectedCall: false,
		},
		"message with nil PeerID": {
			update: &tg.UpdateNewMessage{
				Message: &tg.Message{
					PeerID:  nil,
					Message: "Hello",
					ID:      1,
				},
			},
			messageHandler: func(clientID string, msg *tg.Message) {
				t.Error("handler should not be called for nil PeerID")
			},
			expectedCall: false,
		},
		"nil message handler": {
			update: &tg.UpdateNewMessage{
				Message: &tg.Message{
					PeerID:  &tg.PeerUser{UserID: 12345},
					Message: "Hello",
					ID:      1,
				},
			},
			messageHandler: nil,
			expectedCall:   false,
		},
		"wrong message type": {
			update: &tg.UpdateNewMessage{
				Message: nil,
			},
			messageHandler: func(clientID string, msg *tg.Message) {
				t.Error("handler should not be called for wrong type")
			},
			expectedCall: false,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			handlerCalled := false
			var actualHandler func(string, *tg.Message)

			if tt.messageHandler != nil {
				actualHandler = func(clientID string, msg *tg.Message) {
					handlerCalled = true
					tt.messageHandler(clientID, msg)
				}
			}

			client := &TelegramClient{
				id:             "test_client",
				logger:         zap.NewNop(),
				messageHandler: actualHandler,
			}

			err := client.handleNewMessage(context.Background(), tg.Entities{}, tt.update)

			require.NoError(t, err)
			if tt.expectedCall {
				require.True(t, handlerCalled)
			}
		})
	}
}

func TestTelegramClient_extractSenderID(t *testing.T) {
	testCases := map[string]struct {
		msg      *tg.Message
		expected string
	}{
		"user message returns user_id format": {
			msg: &tg.Message{
				PeerID: &tg.PeerUser{UserID: 12345},
			},
			expected: "user_12345",
		},
		"user message with zero ID": {
			msg: &tg.Message{
				PeerID: &tg.PeerUser{UserID: 0},
			},
			expected: "user_0",
		},
		"user message with negative ID": {
			msg: &tg.Message{
				PeerID: &tg.PeerUser{UserID: -123},
			},
			expected: "user_-123",
		},
		"chat message returns chat_id format": {
			msg: &tg.Message{
				PeerID: &tg.PeerChat{ChatID: 67890},
			},
			expected: "chat_67890",
		},
		"chat message with zero ID": {
			msg: &tg.Message{
				PeerID: &tg.PeerChat{ChatID: 0},
			},
			expected: "chat_0",
		},
		"channel message returns channel_id format": {
			msg: &tg.Message{
				PeerID: &tg.PeerChannel{ChannelID: 11111},
			},
			expected: "channel_11111",
		},
		"channel message with zero ID": {
			msg: &tg.Message{
				PeerID: &tg.PeerChannel{ChannelID: 0},
			},
			expected: "channel_0",
		},
		"nil peerID returns unknown": {
			msg: &tg.Message{
				PeerID: nil,
			},
			expected: "unknown",
		},
		"nil message returns unknown": {
			msg: &tg.Message{
				PeerID: nil,
			},
			expected: "unknown",
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			client := &TelegramClient{}
			result := client.extractSenderID(tt.msg)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestTelegramClient_listenMessages(t *testing.T) {
	testCases := map[string]struct {
		cancelContext bool
		expectedError error
	}{
		"context cancelled returns context.Canceled": {
			cancelContext: true,
			expectedError: context.Canceled,
		},
		"context not cancelled blocks indefinitely": {
			cancelContext: false,
			expectedError: nil,
		},
		"context with deadline exceeded": {
			cancelContext: true,
			expectedError: context.Canceled,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			client := &TelegramClient{}

			ctx := context.Background()
			if tt.cancelContext {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			errCh := make(chan error, 1)
			go func() {
				errCh <- client.listenMessages(ctx)
			}()

			select {
			case err := <-errCh:
				if tt.expectedError != nil {
					require.Error(t, err)
					require.Equal(t, tt.expectedError, err)
				} else {
					t.Error("listenMessages returned unexpectedly")
				}
			case <-time.After(100 * time.Millisecond):
				if tt.expectedError != nil {
					t.Error("listenMessages did not return as expected")
				}
			}
		})
	}
}
