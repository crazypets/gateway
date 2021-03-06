package main

import (
	"context"
	"flag"
	"fmt"
	"gateway/internal/app"
	"gateway/internal/config"
	zapLogger "gateway/internal/infrastructure/logger"
	"gateway/internal/infrastructure/tracer"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rtsoftSG/tgbot"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

func main() {
	configFile := flag.String("config", "./configs/config.yaml", "configuration file path")
	flag.Parse()

	cfg, err := config.Load(*configFile)
	if err != nil {
		panic(err)
	}

	logger := zapLogger.New(&cfg.Logger)
	tgSDK := tgbot.NewSDK(&http.Client{Timeout: time.Second * 30}, cfg.TgBotAddr)

	closer, err := tracer.InitGlobalTracer(cfg.Jaeger.ServiceName, cfg.Jaeger.AgentAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer closer.Close()

	svcID := uuid.NewV4().String()
	application := app.New(cfg,
		app.WithLogger(logger),
		app.WithServiceID(svcID))

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		if err = application.Run(); err != nil {
			logger.Fatal("launch application", zap.Error(err))
		}
	}()

	err = tgSDK.Send(context.Background(), time.Now().UTC(), "INFO",
		fmt.Sprintf("auth-service: %s is launched!", svcID))
	if err != nil {
		logger.Fatal("tg service is unreachable!", zap.Error(err))
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-done

	if err = application.Stop(); err != nil {
		logger.Fatal("stop application", zap.Error(err))
	}
	wg.Wait()

	err = tgSDK.Send(context.Background(), time.Now().UTC(), "INFO",
		fmt.Sprintf("auth-service: %s is gracefully stopped!", svcID))
	if err != nil {
		logger.Fatal("tg service is unreachable!", zap.Error(err))
	}
}
