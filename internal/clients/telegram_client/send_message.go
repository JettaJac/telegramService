package telegram_client

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

// SendMessage отправляем сообщения в телеграмм
func (c *TelegramClient) SendMessage(ctx context.Context, chatID, text string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	c.mu.RLock()
	if !c.authorized {
		c.mu.RUnlock()
		return "", fmt.Errorf("client not authorized")
	}
	c.mu.RUnlock()

	c.logger.Info("sending message",
		zap.String("chat_id", chatID),
		zap.Int("text_length", len(text)))

	rawAPI := c.client.API().Raw()
	if rawAPI == nil {
		return "", fmt.Errorf("failed to get raw API client")
	}
	sender := message.NewSender(rawAPI)

	peerInfo, err := c.resolvePeer(ctx, chatID)
	if err != nil {
		return "", fmt.Errorf("failed to resolve peerInfo: %w", err)
	}

	_, err = sender.To(peerInfo).Text(ctx, text)
	if err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	msgID := generateMessageID()
	c.logger.Info("message sent",
		zap.String("message_id", msgID),
		zap.String("chat_id", chatID))

	return msgID, nil
}

func (c *TelegramClient) resolvePeer(ctx context.Context, chatID string) (tg.InputPeerClass, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	username := strings.TrimPrefix(chatID, "@")

	resolved, err := c.client.API().ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
		Username: username,
	})
	if err == nil {
		// Пользователи
		if len(resolved.Users) > 0 {
			if user, ok := resolved.Users[0].(*tg.User); ok {
				return &tg.InputPeerUser{
					UserID:     user.ID,
					AccessHash: user.AccessHash,
				}, nil
			}
		}

		if len(resolved.Chats) > 0 {
			switch chat := resolved.Chats[0].(type) {
			case *tg.Channel:
				return &tg.InputPeerChannel{
					ChannelID:  chat.ID,
					AccessHash: chat.AccessHash,
				}, nil
			case *tg.Chat:
				return &tg.InputPeerChat{
					ChatID: chat.ID,
				}, nil
			}
		}
	}

	if id, err := strconv.ParseInt(chatID, 10, 64); err == nil {
		if id < 0 {
			return &tg.InputPeerChannel{
				ChannelID: id,
			}, nil
		}
		return &tg.InputPeerUser{
			UserID: id,
		}, nil
	}

	return nil, fmt.Errorf("cannot resolve peer: %s", chatID)
}

func generateMessageID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (c *TelegramClient) IsAuthorized() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.authorized
}
