package telegram_service

import (
	"context"

	desc "telegramservice/pb/go"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DeleteConnection ...
func (s *GRPCServer) DeleteConnection(ctx context.Context, req *desc.DeleteConnectionRequest) (*desc.DeleteConnectionResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if req.ConnectionId == "" {
		return nil, status.Error(codes.InvalidArgument, "connection_id is required")
	}

	err := s.service.DeleteConnection(ctx, req.ConnectionId, req.Logout)
	if err != nil {
		return &desc.DeleteConnectionResponse{
			ConnectionId: req.ConnectionId,
			Success:      false,
			Message:      err.Error(),
		}, nil
	}

	return &desc.DeleteConnectionResponse{
		ConnectionId: req.ConnectionId,
		Success:      true,
		Message:      "connection deleted successfully",
	}, nil
}
