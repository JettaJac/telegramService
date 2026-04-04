package storage

import (
	"sync"
)

// Message сообщение из телеграм
type Message struct {
	ID           string
	ConnectionID string
	SenderID     string
	Text         string
	Timestamp    int64
	ChatID       string
}

// ConnectionInfo информация о соединение
type ConnectionInfo struct {
	ID        string
	Status    string
	CreatedAt int64
}

// MemoryStorage хранилище
type MemoryStorage struct {
	mu          sync.RWMutex
	connections map[string]*ConnectionInfo
	messages    map[string][]*Message
}

// NewMemoryStorage создаем хранилище
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		connections: make(map[string]*ConnectionInfo),
		messages:    make(map[string][]*Message),
	}
}

// SaveConnection сохраняем соединение
func (s *MemoryStorage) SaveConnection(conn *ConnectionInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connections[conn.ID] = conn
}

// GetConnection получаем информацию о соединение
func (s *MemoryStorage) GetConnection(id string) (*ConnectionInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	conn, ok := s.connections[id]
	return conn, ok
}

// DeleteConnection удаляем соединение
func (s *MemoryStorage) DeleteConnection(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.connections, id)
	delete(s.messages, id)
}

// ListConnections всв соединения
func (s *MemoryStorage) ListConnections() []*ConnectionInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*ConnectionInfo, 0, len(s.connections))
	for _, conn := range s.connections {
		result = append(result, conn)
	}
	return result
}

// UpdateConnectionStatus обновчление статуса соединения
func (s *MemoryStorage) UpdateConnectionStatus(id, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if conn, ok := s.connections[id]; ok {
		conn.Status = status
	}
}

// AddMessage добавляем сообщение к соединению
func (s *MemoryStorage) AddMessage(msg *Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages[msg.ConnectionID] = append(s.messages[msg.ConnectionID], msg)
}

// GetMessages получаем сообщения по соединению
func (s *MemoryStorage) GetMessages(connectionID string, limit int) []*Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	msgs, ok := s.messages[connectionID]
	if !ok {
		return []*Message{}
	}

	if limit > 0 && len(msgs) > limit {
		return msgs[len(msgs)-limit:]
	}
	return msgs
}
