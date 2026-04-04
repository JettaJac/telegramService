package telegram_service

import "context"

//go:generate ../../../bin/mockgen -typed -source telegram_interface.go -destination telegram_interface_mock.go -package=telegram_service
type TelegramClientInterface interface {
	Start(ctx context.Context, qrCodeReady func(string)) error
	Stop(ctx context.Context, logout bool) error
	IsAuthorized() bool
	SendMessage(ctx context.Context, chatID, text string) (string, error)
}
