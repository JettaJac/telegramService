package browser

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	utlLB "net/url"

	"github.com/skip2/go-qrcode"
	"go.uber.org/zap"
)

type QRServer struct {
	port     int
	logger   *zap.Logger
	server   *http.Server
	qrData   string
	readyCh  chan struct{}
	closedCh chan struct{}
}

func NewQRServer(logger *zap.Logger) *QRServer {
	return &QRServer{
		port:     0,
		logger:   logger,
		readyCh:  make(chan struct{}),
		closedCh: make(chan struct{}),
	}
}

// Start запускает HTTP сервер для показа QR кода
func (s *QRServer) Start(qrData string) (string, error) {
	if qrData == "" {
		return "", fmt.Errorf("qrData cannot be empty")
	}

	s.qrData = qrData

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		png, err := qrcode.Encode(qrData, qrcode.Medium, 300)
		if err != nil {
			http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(png)
	})

	// Страница со статусом
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-s.readyCh:
			_, _ = w.Write([]byte("authorized"))
		default:
			_, _ = w.Write([]byte("pending"))
		}
	})

	s.server = &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("failed to find free port: %w", err)
	}

	s.port = listener.Addr().(*net.TCPAddr).Port

	// Запускаем сервер
	go func() {
		if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			s.logger.Error("QR server error", zap.Error(err))
		}
	}()

	url := fmt.Sprintf("http://localhost:%d", s.port)
	return url, nil
}

// OpenBrowser открывает URL в браузере по умолчанию
func OpenBrowser(url string) error {
	err := validUrl(url)
	if err != nil {
		return err
	}
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin": // macOS
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if err != nil {
		return fmt.Errorf("failed to open browser: %w", err)
	}
	return nil
}

// Close закрывает сервер и браузер
func (s *QRServer) Close() error {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.server.Shutdown(ctx); err != nil {
			s.logger.Error("failed to shutdown QR server", zap.Error(err))
		}
	}
	close(s.closedCh)
	return nil
}

func (s *QRServer) SetAuthorized() {
	select {
	case <-s.readyCh:
		// уже закрыт
	default:
		close(s.readyCh)
	}
}

// waitForScan ожидает сканирования QR кода или таймаута
func (s *QRServer) waitForScan(ctx context.Context, timeout time.Duration) error {
	select {
	case <-s.readyCh:
		s.logger.Info("QR code scanned successfully")
		return nil
	case <-time.After(timeout):
		s.logger.Warn("QR code scan timeout")
		return fmt.Errorf("QR code scan timeout after %v", timeout)
	case <-ctx.Done():
		return ctx.Err()
	}
}

func validUrl(url string) error {
	if url == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	if strings.TrimSpace(url) == "" {
		return fmt.Errorf("URL contains only whitespace")
	}

	parsedURL, err := utlLB.Parse(url)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Scheme == "" {
		return fmt.Errorf("URL missing scheme (http://, https://, etc.)")
	}

	supportedSchemes := map[string]bool{
		"http":  true,
		"https": true,
		"file":  true,
		"ftp":   true,
	}

	if !supportedSchemes[parsedURL.Scheme] {
		return fmt.Errorf("unsupported URL scheme: %s (supported: http, https, file, ftp)", parsedURL.Scheme)
	}

	if parsedURL.Scheme == "http" || parsedURL.Scheme == "https" {
		if parsedURL.Host == "" {
			return fmt.Errorf("URL missing host for scheme %s", parsedURL.Scheme)
		}
	}
	return nil
}
