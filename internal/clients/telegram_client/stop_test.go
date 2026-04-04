package telegram_client

import (
	"context"
	"testing"

	"go.uber.org/zap"
)

func TestTelegramClient_Stop(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	testCases := map[string]struct {
		logout      bool
		hasCancel   bool
		expectedErr bool
	}{
		"stop without logout": {
			logout:      false,
			hasCancel:   true,
			expectedErr: false,
		},
		"stop with logout": {
			logout:      true,
			hasCancel:   true,
			expectedErr: false,
		},
		"stop without cancel func": {
			logout:      false,
			hasCancel:   false,
			expectedErr: false,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			var cancelFunc context.CancelFunc
			if tt.hasCancel {
				_, cancelFunc = context.WithCancel(context.Background())
			}

			client := &TelegramClient{
				id:         "test_client",
				logger:     logger,
				authorized: true,
				cancelFunc: cancelFunc,
				sessionDir: "./test_sessions",
			}

			ctx := context.Background()
			err := client.Stop(ctx, tt.logout)

			if tt.expectedErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if client.IsAuthorized() {
				t.Errorf("client should not be authorized after stop")
			}
		})
	}
}
