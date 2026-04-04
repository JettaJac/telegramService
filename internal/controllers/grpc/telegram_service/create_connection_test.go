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

func TestGRPCServer_CreateConnection(t *testing.T) {
	testCases := map[string]struct {
		connectionID string
		setupMock    func(mockStorage *service.MockStorageInterface)
		wantErr      bool
		expectedCode codes.Code
	}{
		"success": {
			connectionID: "test_conn",
			setupMock: func(mockStorage *service.MockStorageInterface) {
				mockStorage.EXPECT().SaveConnection(gomock.Any()).Return().AnyTimes()
				mockStorage.EXPECT().UpdateConnectionStatus(gomock.Any(), gomock.Any()).Return().AnyTimes()
				mockStorage.EXPECT().GetConnection("test_conn").Return(nil, false).AnyTimes()
			},
			wantErr: false,
		},
		"empty connection id": {
			connectionID: "",
			setupMock:    func(mockStorage *service.MockStorageInterface) {},
			wantErr:      true,
			expectedCode: codes.InvalidArgument,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockTelegramClient := service.NewMockTelegramClientInterface(ctrl)

			server, m := NewMocks(ctrl, *mockTelegramClient)
			if tt.setupMock != nil {
				tt.setupMock(m.storage)
			}
			req := &desc.CreateConnectionRequest{ConnectionId: tt.connectionID}

			resp, err := server.CreateConnection(context.Background(), req)

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
				if resp.ConnectionId != tt.connectionID {
					t.Errorf("expected connection_id %q, got %q", tt.connectionID, resp.ConnectionId)
				}
			}
		})
	}
}

func TestGRPCServer_validateSendMessageRequest(t *testing.T) {
	server := &GRPCServer{}

	testCases := map[string]struct {
		req          *desc.SendMessageRequest
		wantErr      bool
		expectedCode codes.Code
	}{
		"valid request": {
			req: &desc.SendMessageRequest{
				ConnectionId: "test_conn",
				ChatId:       "@username",
				Text:         "Hello",
			},
			wantErr: false,
		},
		"empty connection_id": {
			req: &desc.SendMessageRequest{
				ConnectionId: "",
				ChatId:       "@username",
				Text:         "Hello",
			},
			wantErr:      true,
			expectedCode: codes.InvalidArgument,
		},
		"empty chat_id": {
			req: &desc.SendMessageRequest{
				ConnectionId: "test_conn",
				ChatId:       "",
				Text:         "Hello",
			},
			wantErr:      true,
			expectedCode: codes.InvalidArgument,
		},
		"empty text": {
			req: &desc.SendMessageRequest{
				ConnectionId: "test_conn",
				ChatId:       "@username",
				Text:         "",
			},
			wantErr:      true,
			expectedCode: codes.InvalidArgument,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			err := server.validateSendMessageRequest(tt.req)

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
			}
		})
	}
}
