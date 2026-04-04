package telegram_service

//func TestService_SendMessage(t *testing.T) {
//	logger, _ := zap.NewDevelopment()
//	strg := storage.NewMemoryStorage()
//
//	// Сохраняем соединение
//	strg.SaveConnection(&storage.ConnectionInfo{
//		ID:        "test_conn_auth",
//		Status:    "authorized",
//		CreatedAt: time.Now().Unix(),
//	})
//
//	testCases := map[string]struct {
//		connectionID  string
//		chatID        string
//		text          string
//		mockClient    *client.MockTelegramClient
//		expectedMsgID string
//		expectedError string
//	}{
//		"success send message": {
//			connectionID: "test_conn_auth",
//			chatID:       "@username",
//			text:         "Hello, World!",
//			mockClient: &client.MockTelegramClient{
//				IsAuthorizedFunc: func() bool { return true },
//				SendMessageFunc: func(ctx context.Context, chatID, text string) (string, error) {
//					return "msg_12345", nil
//				},
//			},
//			expectedMsgID: "msg_12345",
//			expectedError: "",
//		},
//		"connection not found": {
//			connectionID:  "non_existent",
//			chatID:        "@username",
//			text:          "Hello",
//			mockClient:    nil,
//			expectedMsgID: "",
//			expectedError: "connection non_existent not found",
//		},
//		"connection not authorized": {
//			connectionID: "test_conn_unauth",
//			chatID:       "@username",
//			text:         "Hello",
//			mockClient: &client.MockTelegramClient{
//				IsAuthorizedFunc: func() bool { return false },
//				SendMessageFunc: func(ctx context.Context, chatID, text string) (string, error) {
//					return "", fmt.Errorf("not authorized")
//				},
//			},
//			expectedMsgID: "",
//			expectedError: "connection test_conn_unauth not authorized",
//		},
//		"empty chat id": {
//			connectionID: "test_conn_auth",
//			chatID:       "",
//			text:         "Hello",
//			mockClient: &client.MockTelegramClient{
//				IsAuthorizedFunc: func() bool { return true },
//				SendMessageFunc: func(ctx context.Context, chatID, text string) (string, error) {
//					return "", fmt.Errorf("chat_id is required")
//				},
//			},
//			expectedMsgID: "",
//			expectedError: "chat_id is required",
//		},
//		"empty text": {
//			connectionID: "test_conn_auth",
//			chatID:       "@username",
//			text:         "",
//			mockClient: &client.MockTelegramClient{
//				IsAuthorizedFunc: func() bool { return true },
//				SendMessageFunc: func(ctx context.Context, chatID, text string) (string, error) {
//					return "", fmt.Errorf("text is required")
//				},
//			},
//			expectedMsgID: "",
//			expectedError: "text is required",
//		},
//	}
//
//	for name, tt := range testCases {
//		t.Run(name, func(t *testing.T) {
//			s := &Service{
//				clients:     NewExternalServicesClients(),
//				cancelFuncs: make(map[string]context.CancelFunc),
//				storage:     strg,
//				logger:      logger,
//				apiID:       123456,
//				apiHash:     "test_hash",
//				sessionDir:  "./test_sessions",
//				qrTimeout:   60 * time.Second,
//			}
//
//			// Добавляем мок-клиент если нужно
//			if tt.mockClient != nil {
//				s.clients.telegram[tt.connectionID] = tt.mockClient
//			} else if tt.connectionID == "test_conn_unauth" {
//				// Добавляем неавторизованный клиент
//				unauthClient := &client.MockTelegramClient{
//					IsAuthorizedFunc: func() bool { return false },
//				}
//				s.clients.telegram[tt.connectionID] = unauthClient
//				strg.SaveConnection(&storage.ConnectionInfo{
//					ID:        tt.connectionID,
//					Status:    "pending",
//					CreatedAt: time.Now().Unix(),
//				})
//			} else if tt.connectionID == "test_conn_auth" {
//				// Уже есть в storage, но нужно добавить в clients
//				if _, exists := s.clients.telegram[tt.connectionID]; !exists && tt.mockClient != nil {
//					s.clients.telegram[tt.connectionID] = tt.mockClient
//				}
//			}
//
//			ctx := context.Background()
//			msgID, err := s.SendMessage(ctx, tt.connectionID, tt.chatID, tt.text)
//
//			if tt.expectedError != "" {
//				if err == nil {
//					t.Errorf("expected error, got nil")
//				} else if err.Error() != tt.expectedError {
//					t.Errorf("expected error %q, got %q", tt.expectedError, err.Error())
//				}
//			} else {
//				if err != nil {
//					t.Errorf("unexpected error: %v", err)
//				}
//				if msgID != tt.expectedMsgID {
//					t.Errorf("expected message ID %s, got %s", tt.expectedMsgID, msgID)
//				}
//			}
//		})
//	}
//}
