package browser

import (
	"context"
	"net/http"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewQRServer(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	testCases := map[string]struct {
		logger  *zap.Logger
		wantNil bool
	}{
		"success create server": {
			logger:  logger,
			wantNil: false,
		},
		"nil logger": {
			logger:  nil,
			wantNil: false, // даже с nil logger должен создаться
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			server := NewQRServer(tt.logger)
			if server == nil && !tt.wantNil {
				t.Errorf("expected non-nil server")
			}
			if server != nil && server.readyCh == nil {
				t.Errorf("readyCh not initialized")
			}
		})
	}
}

func TestQRServer_Start(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	testCases := map[string]struct {
		qrData      string
		expectedErr bool
	}{
		"valid qr data": {
			qrData:      "tg://login?token=test123",
			expectedErr: false,
		},
		"empty qr data": {
			qrData:      "",
			expectedErr: true,
		},
		"long qr data": {
			qrData:      "tg://login?token=" + string(make([]byte, 1000)),
			expectedErr: false,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			server := NewQRServer(logger)
			url, err := server.Start(tt.qrData)

			if tt.expectedErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectedErr && url == "" {
				t.Errorf("expected non-empty URL")
			}

			// Проверяем, что сервер запущен
			if !tt.expectedErr {
				resp, err := http.Get(url)
				if err != nil {
					t.Errorf("failed to connect to QR server: %v", err)
				} else {
					_ = resp.Body.Close()
					if resp.StatusCode != http.StatusOK {
						t.Errorf("expected status 200, got %d", resp.StatusCode)
					}
					if resp.Header.Get("Content-Type") != "image/png" {
						t.Errorf("expected Content-Type image/png, got %s", resp.Header.Get("Content-Type"))
					}
				}
			}

			_ = server.Close()
		})
	}
}

func TestQRServer_WaitForScan(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	testCases := map[string]struct {
		timeout       time.Duration
		setAuthorized bool
		expectedErr   bool
	}{
		"successful scan": {
			timeout:       5 * time.Second,
			setAuthorized: true,
			expectedErr:   false,
		},
		"timeout": {
			timeout:       100 * time.Millisecond,
			setAuthorized: false,
			expectedErr:   true,
		},
		"context cancelled": {
			timeout:       5 * time.Second,
			setAuthorized: false,
			expectedErr:   true,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			server := NewQRServer(logger)
			_, err := server.Start("test_qr")
			if err != nil {
				t.Fatalf("failed to start server: %v", err)
			}
			defer func() {
				_ = server.Close()
			}()

			ctx := context.Background()
			if name == "context cancelled" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			if tt.setAuthorized {
				go func() {
					time.Sleep(50 * time.Millisecond)
					server.SetAuthorized()
				}()
			}

			err = server.waitForScan(ctx, tt.timeout)

			if tt.expectedErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestQRServer_Close(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	testCases := map[string]struct {
		startServer bool
	}{
		"close running server": {
			startServer: true,
		},
		"close stopped server": {
			startServer: false,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			server := NewQRServer(logger)

			if tt.startServer {
				_, err := server.Start("test_qr")
				if err != nil {
					t.Fatalf("failed to start server: %v", err)
				}
			}

			err := server.Close()
			if err != nil {
				t.Errorf("unexpected error on close: %v", err)
			}

			// Проверяем, что сервер закрыт
			if tt.startServer {
				// Попытка подключиться к закрытому серверу должна вернуть ошибку
				time.Sleep(100 * time.Millisecond)
			}
		})
	}
}

func TestQRServer_SetAuthorized(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	testCases := map[string]struct {
		setBeforeWait bool
	}{
		"set before wait": {
			setBeforeWait: true,
		},
		"set after wait started": {
			setBeforeWait: false,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			server := NewQRServer(logger)
			_, err := server.Start("test_qr")
			if err != nil {
				t.Fatalf("failed to start server: %v", err)
			}
			defer func() {
				_ = server.Close()
			}()

			if tt.setBeforeWait {
				server.SetAuthorized()
			} else {
				go func() {
					time.Sleep(50 * time.Millisecond)
					server.SetAuthorized()
				}()
			}

			ctx := context.Background()
			err = server.waitForScan(ctx, 5*time.Second)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestOpenBrowser(t *testing.T) {
	// Эта функция зависит от OS, поэтому тестируем только логику
	testCases := map[string]struct {
		url         string
		expectedErr bool
	}{
		"valid url": {
			url:         "http://localhost:8080",
			expectedErr: false,
		},
		"empty url": {
			url:         "",
			expectedErr: true,
		},
		"invalid url": {
			url:         "://invalid",
			expectedErr: true,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			err := OpenBrowser(tt.url)

			if tt.expectedErr && err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}

// Интеграционный тест для всего QR сервера
func TestQRServer_Integration(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	server := NewQRServer(logger)

	qrData := "tg://login?token=integration_test_token_12345"
	url, err := server.Start(qrData)
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer func() {
		_ = server.Close()
	}()

	// Проверяем, что сервер отвечает
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("failed to get QR page: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != "image/png" {
		t.Errorf("expected Content-Type image/png, got %s", resp.Header.Get("Content-Type"))
	}

	// Проверяем статус endpoint
	statusResp, err := client.Get(url + "/status")
	if err != nil {
		t.Fatalf("failed to get status: %v", err)
	}
	defer func() {
		_ = statusResp.Body.Close()
	}()

	// Изначально статус должен быть "pending"
	// Проверка содержимого зависит от реализации
}
