package telegram_service

import (
	"context"
	service "telegramservice/internal/service/telegram_service"
	desc "telegramservice/pb/go"
	"testing"

	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGRPCServer_DeleteConnection(t *testing.T) {
	testCases := map[string]struct {
		req          *desc.DeleteConnectionRequest
		setupMock    func(mockStorage *service.MockStorageInterface)
		wantErr      bool
		expectedCode codes.Code
	}{
		"success": {
			req: &desc.DeleteConnectionRequest{
				ConnectionId: "test_conn",
				Logout:       true,
			},
			setupMock: func(mockStorage *service.MockStorageInterface) {
			},
			wantErr: false,
		},
		"empty connection id": {
			req: &desc.DeleteConnectionRequest{
				ConnectionId: "",
				Logout:       false,
			},
			setupMock:    func(mockStorage *service.MockStorageInterface) {},
			wantErr:      true,
			expectedCode: codes.InvalidArgument,
		},
		"connection not found": {
			req: &desc.DeleteConnectionRequest{
				ConnectionId: "non_existent",
				Logout:       false,
			},
			setupMock: func(mockStorage *service.MockStorageInterface) {},
			wantErr:   false,
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

			resp, err := server.DeleteConnection(context.Background(), tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if st, ok := status.FromError(err); ok {
					if st.Code() != tt.expectedCode {
						t.Errorf("expected code %v, got %v", tt.expectedCode, st.Code())
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if resp == nil {
					t.Errorf("expected response, got nil")
				}
			}
		})
	}
}
