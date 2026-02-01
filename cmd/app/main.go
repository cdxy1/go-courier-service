package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/cdxy1/go-courier-service/internal/app"
	"github.com/cdxy1/go-courier-service/internal/infra/postgres"
	"github.com/cdxy1/go-courier-service/internal/pprofserver"
	"github.com/cdxy1/go-courier-service/pkg/config"
)

func main() {
	cfg := config.GetEnv()

	pool, err := postgres.NewPostgresConnection(cfg)
	if err != nil {
		log.Fatalln("postgres not started:", err.Error())
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	a := app.NewApp(ctx, pool, cfg)

	a.Start(ctx)

	var pprofSrv *pprofserver.Server
	if cfg.Pprof != nil && cfg.Pprof.Enabled {
		pprofSrv = pprofserver.New(cfg.Pprof)
		go func() {
			if err := pprofSrv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Printf("pprof server stopped: %v", err)
			}
		}()
	}

	go func() {
		if err := a.Echo.Start(fmt.Sprintf(":%s", cfg.Port)); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.Echo.Logger.Fatal(err)
		}
	}()

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer a.Echo.Logger.Print("Shutting down service-courier")
	defer cancel()
	defer pool.Close()
	defer func() {
		if err := a.OrderGateway.Close(); err != nil {
			a.Echo.Logger.Printf("close order gateway: %v", err)
		}
	}()
	defer func() {
		if pprofSrv == nil {
			return
		}
		if err := pprofSrv.Shutdown(ctx); err != nil {
			log.Printf("shutdown pprof server: %v", err)
		}
	}()
	if a.EventConsumer != nil {
		defer func() {
			if err := a.EventConsumer.Close(); err != nil {
				a.Echo.Logger.Printf("close event consumer: %v", err)
			}
		}()
	}

	if err := a.Echo.Shutdown(ctx); err != nil {
		a.Echo.Logger.Fatal(err)
	}
}
