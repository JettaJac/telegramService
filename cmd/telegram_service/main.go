package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"telegramservice/internal/config"

	grpcservice "telegramservice/internal/controllers/grpc/telegram_service"
	tservice "telegramservice/internal/service/telegram_service"
	"telegramservice/internal/storage"
	desc "telegramservice/pb/go"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// прпогнать по линту

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	log := initLogger(cfg)
	defer func() {
		_ = log.Sync()
	}()

	log.Info("starting Telegram Service",
		zap.String("version", "1.0.0"),
		zap.String("environment", cfg.Environment),
		zap.Int("api_id", cfg.Clients.TelegramClient.APIID),
		zap.String("grpc_port", cfg.Server.Port),
		zap.String("session_dir", cfg.Clients.TelegramClient.SessionDir),
		zap.String("log_level", cfg.Log.Level))

	storageService := storage.NewMemoryStorage()

	tService := tservice.NewService(
		cfg.Clients.TelegramClient.APIID,
		cfg.Clients.TelegramClient.APIHash,
		cfg.Clients.TelegramClient.SessionDir,
		storageService,
		cfg.Clients.TelegramClient.QRTimeout,
	)

	grpcServer := grpcservice.NewGRPCServer(tService)

	lis, err := net.Listen("tcp", ":"+cfg.Server.Port)
	if err != nil {
		log.Fatal("failed to listen",
			zap.String("error", err.Error()),
			zap.String("address", cfg.Server.Port))
	}

	s := grpc.NewServer()
	desc.RegisterTelegramServiceServer(s, grpcServer)
	reflection.Register(s)

	go func() {
		log.Info("starting gRPC server", zap.String("port", cfg.Server.Port))
		if err := s.Serve(lis); err != nil {
			log.Fatal("failed to serve", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")
	s.GracefulStop()
	log.Info("server stopped")
}

func initLogger(cfg *config.Config) *zap.Logger {
	var level zapcore.Level
	switch cfg.Log.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	cnf := zap.Config{
		Level:       zap.NewAtomicLevelAt(level),
		Development: cfg.Environment == "development",
		Sampling:    nil,
		Encoding:    cfg.Log.Format,
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{cfg.Log.Output},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := cnf.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}

	return logger
}
