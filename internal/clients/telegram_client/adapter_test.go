package telegram_client

import (
	"testing"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestTGClientAdapter тестирует адаптер для tg.Client
func TestTGClientAdapter(t *testing.T) {
	t.Run("ContactsResolveUsername calls underlying client", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		t.Skip("requires real tg.Client or integration test")
	})

	t.Run("Raw returns underlying client", func(t *testing.T) {
		// Создаем реальный tg.Client (nil для теста, так как он нам не нужен)
		var realClient *tg.Client = nil
		adapter := NewTGClientAdapter(realClient)

		result := adapter.Raw()
		assert.Equal(t, realClient, result)
	})
}

// TestQRClientAdapter тестирует адаптер для qrlogin.QR
func TestQRClientAdapter(t *testing.T) {
	t.Run("Auth calls underlying QR Auth", func(t *testing.T) {
		t.Skip("requires mock of qrlogin.QR")
	})
}

// TestTelegramClientAdapter тестирует основной адаптер
func TestTelegramClientAdapter(t *testing.T) {
	t.Run("API returns TGClientInterface", func(t *testing.T) {
		t.Skip("requires real telegram.Client")
	})

	t.Run("Run calls underlying client Run", func(t *testing.T) {
		t.Skip("requires real telegram.Client")
	})

	t.Run("Auth returns AuthClientInterface", func(t *testing.T) {
		t.Skip("requires real telegram.Client")
	})

	t.Run("QR returns QRClientInterface", func(t *testing.T) {
		t.Skip("requires real telegram.Client")
	})
}
