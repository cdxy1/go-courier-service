package app

import (
	"context"
	"fmt"
	"time"

	"github.com/cdxy1/go-courier-service/internal/gateway/order"
	"github.com/cdxy1/go-courier-service/internal/gateway/orderhttp"
	hc "github.com/cdxy1/go-courier-service/internal/handler/courier"
	hd "github.com/cdxy1/go-courier-service/internal/handler/delivery"
	ipostgres "github.com/cdxy1/go-courier-service/internal/infra/postgres"
	"github.com/cdxy1/go-courier-service/internal/model"
	"github.com/cdxy1/go-courier-service/internal/observability"
	"github.com/cdxy1/go-courier-service/internal/ratelimit"
	rc "github.com/cdxy1/go-courier-service/internal/repository/courier"
	rd "github.com/cdxy1/go-courier-service/internal/repository/delivery"
	"github.com/cdxy1/go-courier-service/internal/routes"
	"github.com/cdxy1/go-courier-service/internal/transport/kafka"
	ucc "github.com/cdxy1/go-courier-service/internal/usecase/courier"
	ucd "github.com/cdxy1/go-courier-service/internal/usecase/delivery"
	"github.com/cdxy1/go-courier-service/internal/usecase/order_event"
	"github.com/cdxy1/go-courier-service/internal/worker"
	"github.com/cdxy1/go-courier-service/pkg/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type App struct {
	Echo             *echo.Echo
	Worker           *worker.OrderAssigner
	DeliveryMonitor  *worker.DeliveryMonitor
	OrderGateway     *order.OrderGateway
	OrderHTTPGateway *orderhttp.OrderGateway
	EventConsumer    *kafka.Consumer
	cfg              *config.Сonfig
}

func NewApp(ctx context.Context, conn *pgxpool.Pool, cfg *config.Сonfig) *App {
	e := echo.New()
	e.Use(observability.MetricsAndLogging())
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	crepo := rc.NewCourierRepository(conn)
	cuc := ucc.NewCourierUsecase(crepo)
	ch := hc.NewCourierHandler(cuc)

	tm := ipostgres.NewTxManager(conn)
	drepo := rd.NewDeliveryRepository(conn)
	timeFactory := model.NewDeliveryTimeFactory(
		cfg.Delivery.OnFootDuration,
		cfg.Delivery.ScooterDuration,
		cfg.Delivery.CarDuration,
	)
	duc := ucd.NewDeliveryUsecase(crepo, drepo, tm, timeFactory, model.UTCNow)
	cd := hd.NewDeliveryHandler(duc)
	deliveryMonitor := worker.NewDeliveryMonitor(duc, cfg.Delivery.MonitorInterval, nil)

	apiLimiter := ratelimit.NewTokenBucketLimiter(5, 5, time.Minute)
	apiRateLimitMiddleware := ratelimit.Middleware(apiLimiter, nil)
	r := routes.NewRoutes(ch, cd, apiRateLimitMiddleware)
	r.Register(e)

	orderGateway, err := order.NewOrderGateway(cfg.OrderServiceGRPC)
	if err != nil {
		panic(fmt.Sprintf("failed to create order gateway: %v", err))
	}
	orderAssigner := worker.NewOrderAssigner(orderGateway, duc)

	orderHTTPGateway := orderhttp.NewOrderGateway(cfg.OrderServiceHTTP)
	eventFactory := order_event.NewHandlerFactory(duc)
	eventProcessor := order_event.NewProcessor(eventFactory, orderHTTPGateway)

	var eventConsumer *kafka.Consumer
	if cfg.Kafka != nil && cfg.Kafka.Enabled {
		consumer, err := kafka.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID, cfg.Kafka.Topic, cfg.Kafka.Version, eventProcessor)
		if err != nil {
			panic(fmt.Sprintf("failed to create kafka consumer: %v", err))
		}
		eventConsumer = consumer
	}

	return &App{
		Echo:             e,
		Worker:           orderAssigner,
		DeliveryMonitor:  deliveryMonitor,
		OrderGateway:     orderGateway,
		OrderHTTPGateway: orderHTTPGateway,
		EventConsumer:    eventConsumer,
		cfg:              cfg,
	}
}

func (a *App) Start(ctx context.Context) {
	if a.cfg.OrderPolling {
		go a.Worker.Start(ctx)
	}
	if a.DeliveryMonitor != nil {
		go a.DeliveryMonitor.Start(ctx)
	}
	if a.EventConsumer != nil {
		go func() {
			if err := a.EventConsumer.Start(ctx); err != nil {
				a.Echo.Logger.Printf("kafka consumer stopped: %v", err)
			}
		}()
	}
}
