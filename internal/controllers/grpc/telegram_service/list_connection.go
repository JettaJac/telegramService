package telegram_service

import (
	"context"
	desc "telegramservice/pb/go"
)

// ListConnections ...
func (s *GRPCServer) ListConnections(ctx context.Context, req *desc.ListConnectionsRequest) (*desc.ListConnectionsResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	connections := s.service.ListConnections()

	pbConnections := make([]*desc.ConnectionInfo, len(connections))
	for i, conn := range connections {
		pbConnections[i] = &desc.ConnectionInfo{
			ConnectionId: conn.ID,
			Status:       conn.Status,
			CreatedAt:    conn.CreatedAt,
		}
	}

	return &desc.ListConnectionsResponse{
		Connections: pbConnections,
	}, nil
}
