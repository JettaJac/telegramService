package telegram_client

import (
	"context"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

// TGClientAdapter - адаптер, который реализует TGClientInterface
type TGClientAdapter struct {
	client *tg.Client
}

// NewTGClientAdapter создаем TGClientAdapter
func NewTGClientAdapter(client *tg.Client) *TGClientAdapter {
	return &TGClientAdapter{client: client}
}

// ContactsResolveUsername ...
func (a *TGClientAdapter) ContactsResolveUsername(ctx context.Context, request *tg.ContactsResolveUsernameRequest) (*tg.ContactsResolvedPeer, error) {
	return a.client.ContactsResolveUsername(ctx, request)
}

// TelegramClientAdapter - адаптер для основного клиента
type TelegramClientAdapter struct {
	client *telegram.Client
	api    TGClientInterface
}

// NewTelegramClientAdapter ...
func NewTelegramClientAdapter(client *telegram.Client) *TelegramClientAdapter {
	return &TelegramClientAdapter{
		client: client,
		api:    NewTGClientAdapter(client.API()),
	}
}

// API ...
func (a *TelegramClientAdapter) API() TGClientInterface {
	return a.api
}

// Run ...
func (a *TelegramClientAdapter) Run(ctx context.Context, f func(ctx context.Context) error) error {
	return a.client.Run(ctx, f)
}

// Auth ...
func (a *TelegramClientAdapter) Auth() AuthClientInterface {
	return a.client.Auth()
}

// QR ...
func (a *TelegramClientAdapter) QR() QRClientInterface {
	return a.client.QR()
}

// AuthClientAdapter - адаптер для auth.Client
type AuthClientAdapter struct {
	client *auth.Client
}

// Status ...
func (a *AuthClientAdapter) Status(ctx context.Context) (*auth.Status, error) {
	status, err := a.client.Status(ctx)
	if err != nil {
		return nil, err
	}
	return status, nil
}

// Raw возвращаем tg.Client
func (a *TGClientAdapter) Raw() *tg.Client {
	return a.client
}
