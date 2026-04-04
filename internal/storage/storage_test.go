package storage

import (
	"testing"
	"time"
)

func TestMemoryStorage_Connections(t *testing.T) {
	testCases := map[string]struct {
		connectionID string
		status       string
		wantErr      bool
	}{
		"success save connection": {
			connectionID: "test_conn_1",
			status:       "pending",
			wantErr:      false,
		},
		"success get connection": {
			connectionID: "test_conn_2",
			status:       "authorized",
			wantErr:      false,
		},
		"connection not found": {
			connectionID: "not_exists",
			status:       "",
			wantErr:      true,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			storage := NewMemoryStorage()

			if tt.connectionID != "not_exists" {
				conn := &ConnectionInfo{
					ID:        tt.connectionID,
					Status:    tt.status,
					CreatedAt: time.Now().Unix(),
				}
				storage.SaveConnection(conn)
			}

			conn, ok := storage.GetConnection(tt.connectionID)
			if tt.wantErr && ok {
				t.Errorf("expected error, but conn connection: %+v", conn)
			}
			if !tt.wantErr && !ok {
				t.Errorf("expected connection, but not found")
			}
			if !tt.wantErr && conn.Status != tt.status {
				t.Errorf("expected status %s, conn %s", tt.status, conn.Status)
			}
		})
	}
}

func TestMemoryStorage_Messages(t *testing.T) {
	testCases := map[string]struct {
		connectionID string
		messages     []*Message
		limit        int
		expectedLen  int
	}{
		"success add and get messages": {
			connectionID: "test_conn",
			messages: []*Message{
				{ID: "msg1", Text: "Hello", Timestamp: time.Now().Unix()},
				{ID: "msg2", Text: "World", Timestamp: time.Now().Unix()},
			},
			limit:       10,
			expectedLen: 2,
		},
		"get messages with limit": {
			connectionID: "test_conn",
			messages: []*Message{
				{ID: "msg1", Text: "1"},
				{ID: "msg2", Text: "2"},
				{ID: "msg3", Text: "3"},
			},
			limit:       2,
			expectedLen: 2,
		},
		"empty messages": {
			connectionID: "empty_conn",
			messages:     []*Message{},
			limit:        10,
			expectedLen:  0,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			storage := NewMemoryStorage()

			for _, msg := range tt.messages {
				msg.ConnectionID = tt.connectionID
				storage.AddMessage(msg)
			}

			msg := storage.GetMessages(tt.connectionID, tt.limit)
			if len(msg) != tt.expectedLen {
				t.Errorf("expected %d messages, msg %d", tt.expectedLen, len(msg))
			}
		})
	}
}

func TestMemoryStorage_ListConnections(t *testing.T) {
	testCases := map[string]struct {
		connections   []*ConnectionInfo
		expectedCount int
	}{
		"empty list": {
			connections:   []*ConnectionInfo{},
			expectedCount: 0,
		},
		"single connection": {
			connections: []*ConnectionInfo{
				{ID: "conn1", Status: "pending"},
			},
			expectedCount: 1,
		},
		"multiple connections": {
			connections: []*ConnectionInfo{
				{ID: "conn1", Status: "pending"},
				{ID: "conn2", Status: "authorized"},
				{ID: "conn3", Status: "error"},
			},
			expectedCount: 3,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			storage := NewMemoryStorage()

			for _, conn := range tt.connections {
				storage.SaveConnection(conn)
			}

			conns := storage.ListConnections()
			if len(conns) != tt.expectedCount {
				t.Errorf("expected %d connections, conns %d", tt.expectedCount, len(conns))
			}
		})
	}
}
