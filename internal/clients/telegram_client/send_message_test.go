package telegram_client

import (
	"context"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestTelegramClient_SendMessage(t *testing.T) {
	testCases := map[string]struct {
		chatID           string
		text             string
		authorized       bool
		resolvePeerErr   error
		rawAPINil        bool
		expectedError    bool
		expectedErrorMsg string
	}{
		"error when client not authorized": {
			chatID:           "@username",
			text:             "Hello",
			authorized:       false,
			resolvePeerErr:   nil,
			rawAPINil:        false,
			expectedError:    true,
			expectedErrorMsg: "client not authorized",
		},
		"error when context cancelled": {
			chatID:           "@username",
			text:             "Hello",
			authorized:       true,
			resolvePeerErr:   nil,
			rawAPINil:        false,
			expectedError:    true,
			expectedErrorMsg: "context canceled",
		},
		"error when resolve peer fails": {
			chatID:           "@invalid",
			text:             "Hello",
			authorized:       true,
			resolvePeerErr:   errors.New("user not found"),
			rawAPINil:        false,
			expectedError:    true,
			expectedErrorMsg: "failed to resolve peerInfo",
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTelegram := NewMockTelegramClientInterface(ctrl)
			mockAPI := NewMockTGClientInterface(ctrl)

			ctx := context.Background()
			if name == "error when context cancelled" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			// Setup mocks для авторизованных кейсов
			if tt.authorized && name != "error when context cancelled" {
				// Разрешаем множественные вызовы API()
				mockTelegram.EXPECT().API().Return(mockAPI).Times(2)

				if tt.rawAPINil {
					mockAPI.EXPECT().Raw().Return(nil).Times(1)
				} else if tt.resolvePeerErr != nil {
					mockAPI.EXPECT().Raw().Return(&tg.Client{}).Times(1)
					// ContactsResolveUsername вызывается один раз
					mockAPI.EXPECT().ContactsResolveUsername(gomock.Any(), gomock.Any()).Return(nil, tt.resolvePeerErr).Times(1)
				}
			}

			client := &TelegramClient{
				client:     mockTelegram,
				logger:     zap.NewNop(),
				authorized: tt.authorized,
			}

			msgID, err := client.SendMessage(ctx, tt.chatID, tt.text)

			if tt.expectedError {
				require.Error(t, err)
				if tt.expectedErrorMsg != "" {
					require.Contains(t, err.Error(), tt.expectedErrorMsg)
				}
				require.Empty(t, msgID)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, msgID)
			}
		})
	}
}

func TestTelegramClient_resolvePeer(t *testing.T) {
	testCases := map[string]struct {
		chatID           string
		usernameResolve  bool
		resolvedUser     *tg.User
		resolvedChannel  *tg.Channel
		resolvedChat     *tg.Chat
		resolveError     error
		expectedPeer     tg.InputPeerClass
		expectedError    bool
		expectedErrorMsg string
	}{
		"success resolve by username to user": {
			chatID:          "@username",
			usernameResolve: true,
			resolvedUser: &tg.User{
				ID:         12345,
				AccessHash: 67890,
			},
			resolveError: nil,
			expectedPeer: &tg.InputPeerUser{
				UserID:     12345,
				AccessHash: 67890,
			},
			expectedError: false,
		},
		"success resolve by username to channel": {
			chatID:          "@channel",
			usernameResolve: true,
			resolvedChannel: &tg.Channel{
				ID:         11111,
				AccessHash: 22222,
			},
			resolveError: nil,
			expectedPeer: &tg.InputPeerChannel{
				ChannelID:  11111,
				AccessHash: 22222,
			},
			expectedError: false,
		},
		"success resolve by username to chat": {
			chatID:          "@chat",
			usernameResolve: true,
			resolvedChat: &tg.Chat{
				ID: 33333,
			},
			resolveError: nil,
			expectedPeer: &tg.InputPeerChat{
				ChatID: 33333,
			},
			expectedError: false,
		},
		"resolve by numeric ID as user": {
			chatID:          "12345",
			usernameResolve: false,
			resolveError:    errors.New("not found"),
			expectedPeer: &tg.InputPeerUser{
				UserID: 12345,
			},
			expectedError: false,
		},
		"resolve by numeric negative ID as channel": {
			chatID:          "-12345",
			usernameResolve: false,
			resolveError:    errors.New("not found"),
			expectedPeer: &tg.InputPeerChannel{
				ChannelID: -12345,
			},
			expectedError: false,
		},
		"resolve by username without @ prefix": {
			chatID:          "username",
			usernameResolve: true,
			resolvedUser: &tg.User{
				ID:         54321,
				AccessHash: 98765,
			},
			resolveError: nil,
			expectedPeer: &tg.InputPeerUser{
				UserID:     54321,
				AccessHash: 98765,
			},
			expectedError: false,
		},
		"resolve fails with invalid chatID": {
			chatID:           "invalid@@@",
			usernameResolve:  false,
			resolveError:     errors.New("not found"),
			expectedError:    true,
			expectedErrorMsg: "cannot resolve peer: invalid@@@",
		},
		"resolve with API error": {
			chatID:           "@unknown",
			usernameResolve:  true,
			resolveError:     errors.New("api error: user not found"),
			expectedError:    true,
			expectedErrorMsg: "cannot resolve peer: @unknown",
		},
		"empty chatID": {
			chatID:           "",
			usernameResolve:  false,
			resolveError:     errors.New("not found"),
			expectedError:    true,
			expectedErrorMsg: "cannot resolve peer: ",
		},
		"resolve with context cancelled": {
			chatID:          "@username",
			usernameResolve: true,
			expectedError:   true,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTelegram := NewMockTelegramClientInterface(ctrl)
			mockAPI := NewMockTGClientInterface(ctrl)

			ctx := context.Background()
			if name == "resolve with context cancelled" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			if tt.usernameResolve && name != "resolve with context cancelled" {
				mockTelegram.EXPECT().API().Return(mockAPI).Times(1)

				username := strings.TrimPrefix(tt.chatID, "@")
				mockAPI.EXPECT().ContactsResolveUsername(gomock.Any(), &tg.ContactsResolveUsernameRequest{
					Username: username,
				}).DoAndReturn(func(ctx context.Context, req *tg.ContactsResolveUsernameRequest) (*tg.ContactsResolvedPeer, error) {
					if tt.resolveError != nil {
						return nil, tt.resolveError
					}

					resolved := &tg.ContactsResolvedPeer{}
					if tt.resolvedUser != nil {
						resolved.Users = []tg.UserClass{tt.resolvedUser}
					}
					if tt.resolvedChannel != nil {
						resolved.Chats = []tg.ChatClass{tt.resolvedChannel}
					}
					if tt.resolvedChat != nil {
						resolved.Chats = []tg.ChatClass{tt.resolvedChat}
					}
					return resolved, nil
				}).Times(1)
			} else if name != "resolve with context cancelled" {
				mockTelegram.EXPECT().API().Return(mockAPI).Times(1)
				mockAPI.EXPECT().ContactsResolveUsername(gomock.Any(), gomock.Any()).Return(nil, tt.resolveError).Times(1)
			}

			client := &TelegramClient{
				client: mockTelegram,
				logger: zap.NewNop(),
			}

			peer, err := client.resolvePeer(ctx, tt.chatID)

			if tt.expectedError {
				require.Error(t, err)
				if tt.expectedErrorMsg != "" {
					require.Contains(t, err.Error(), tt.expectedErrorMsg)
				}
				require.Nil(t, peer)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedPeer, peer)
			}
		})
	}
}

func TestTelegramClient_IsAuthorized(t *testing.T) {
	testCases := map[string]struct {
		authorized bool
		expected   bool
	}{
		"authorized": {
			authorized: true,
			expected:   true,
		},
		"not authorized": {
			authorized: false,
			expected:   false,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			client := &TelegramClient{
				authorized: tt.authorized,
			}

			result := client.IsAuthorized()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func Test_generateMessageID(t *testing.T) {
	testCases := map[string]struct {
		wantLength int
	}{
		"generates valid message ID": {
			wantLength: 32, // 16 bytes = 32 hex chars
		},
		"generates unique IDs": {
			wantLength: 32,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			if name == "generates unique IDs" {
				ids := make(map[string]bool)
				for i := 0; i < 1000; i++ {
					id := generateMessageID()
					require.Len(t, id, tt.wantLength)
					require.False(t, ids[id], "duplicate ID generated: %s", id)
					ids[id] = true
				}
			} else {
				id := generateMessageID()
				require.Len(t, id, tt.wantLength)
				_, err := hex.DecodeString(id)
				require.NoError(t, err)
			}
		})
	}
}
