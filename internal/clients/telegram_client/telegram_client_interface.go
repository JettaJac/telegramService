package telegram_client

import (
	"context"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/auth/qrlogin"
	"github.com/gotd/td/tg"
)

//go:generate ../../../bin/mockgen -typed -source telegram_client_interface.go -destination telegram_client_interface_mock.go -package=telegram_client

type TelegramClientInterface interface {
	API() TGClientInterface
	Run(ctx context.Context, f func(ctx context.Context) error) (err error)
	Auth() AuthClientInterface
	QR() QRClientInterface
}

type TGClientInterface interface {
	ContactsResolveUsername(ctx context.Context, request *tg.ContactsResolveUsernameRequest) (*tg.ContactsResolvedPeer, error)
	Raw() *tg.Client
}

type AuthClientInterface interface {
	Status(ctx context.Context) (*auth.Status, error)
}

type QRClientInterface interface {
	Auth(ctx context.Context, loggedIn qrlogin.LoggedIn, show func(ctx context.Context, token qrlogin.Token) error, exceptIDs ...int64) (*tg.AuthAuthorization, error)
}
