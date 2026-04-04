package telegram_service

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	desc "telegramservice/pb/go"
)

// CreateConnection ...
func (s *GRPCServer) CreateConnection(ctx context.Context, req *desc.CreateConnectionRequest) (*desc.CreateConnectionResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if req.ConnectionId == "" {
		return nil, status.Error(codes.InvalidArgument, "connection_id is required")
	}

	qrData, err := s.service.CreateConnection(ctx, req.ConnectionId)
	if err != nil {
		return nil, err
	}

	return &desc.CreateConnectionResponse{
		ConnectionId: req.ConnectionId,
		QrCodeData:   qrData,
		Status:       "pending",
	}, nil
}
