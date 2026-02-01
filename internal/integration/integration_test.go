package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	ipostgres "github.com/cdxy1/go-courier-service/internal/infra/postgres"
	"github.com/cdxy1/go-courier-service/internal/model"
	courierrepo "github.com/cdxy1/go-courier-service/internal/repository/courier"
	deliveryrepo "github.com/cdxy1/go-courier-service/internal/repository/delivery"
	courierusecase "github.com/cdxy1/go-courier-service/internal/usecase/courier"
	deliveryusecase "github.com/cdxy1/go-courier-service/internal/usecase/delivery"
	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestIntegration_CourierDeliveryFlow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pool, terminate := startPostgres(ctx, t)
	defer terminate()

	if err := runMigrations(ctx, pool); err != nil {
		t.Fatalf("failed to prepare schema: %v", err)
	}

	courierRepo := courierrepo.NewCourierRepository(pool)
	deliveryRepo := deliveryrepo.NewDeliveryRepository(pool)
	txManager := ipostgres.NewTxManager(pool)

	courierUC := courierusecase.NewCourierUsecase(courierRepo)
	timeFactory := model.NewDeliveryTimeFactory(time.Minute*30, time.Minute*15, time.Minute*5)
	deliveryUC := deliveryusecase.NewDeliveryUsecase(courierRepo, deliveryRepo, txManager, timeFactory, model.UTCNow)

	courierID, err := courierUC.Create(ctx, &model.CourierModel{
		Name:          "Alice",
		Phone:         "+79990000001",
		Status:        model.CourierStatusAvailable,
		TransportType: model.TransportCar,
	})
	if err != nil {
		t.Fatalf("create courier: %v", err)
	}

	courier, err := courierUC.GetOneById(ctx, courierID)
	if err != nil {
		t.Fatalf("get courier: %v", err)
	}
	if courier.Status != model.CourierStatusAvailable {
		t.Fatalf("expected status available, got %s", courier.Status)
	}

	allCouriers, err := courierUC.GetAll(ctx)
	if err != nil {
		t.Fatalf("get all couriers: %v", err)
	}
	if len(allCouriers) != 1 {
		t.Fatalf("expected 1 courier, got %d", len(allCouriers))
	}

	orderID := "order-integration-1"
	delivery, assignedCourier, err := deliveryUC.Assign(ctx, orderID)
	if err != nil {
		t.Fatalf("assign delivery: %v", err)
	}
	if delivery == nil || assignedCourier == nil {
		t.Fatalf("expected non-nil delivery and courier")
	}
	if assignedCourier.ID != courierID {
		t.Fatalf("expected courier id %d, got %d", courierID, assignedCourier.ID)
	}
	if delivery.CourierId != courierID {
		t.Fatalf("delivery courier id mismatch: %d", delivery.CourierId)
	}
	if delivery.OrderId != orderID {
		t.Fatalf("expected order id %s, got %s", orderID, delivery.OrderId)
	}
	if delivery.Deadline.IsZero() {
		t.Fatalf("deadline should be set")
	}

	var cntAfterAssign int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM delivery WHERE order_id=$1`, orderID).Scan(&cntAfterAssign); err != nil {
		t.Fatalf("query delivery count: %v", err)
	}
	if cntAfterAssign != 1 {
		t.Fatalf("expected 1 delivery row, got %d", cntAfterAssign)
	}

	courierAfterAssign, err := courierUC.GetOneById(ctx, courierID)
	if err != nil {
		t.Fatalf("get courier after assign: %v", err)
	}
	if courierAfterAssign.Status != model.CourierStatusBusy {
		t.Fatalf("expected busy status after assign, got %s", courierAfterAssign.Status)
	}

	unassignResult, err := deliveryUC.Unassign(ctx, orderID)
	if err != nil {
		t.Fatalf("unassign delivery: %v", err)
	}
	if unassignResult == nil {
		t.Fatalf("expected unassign result")
	}
	if unassignResult.CourierId != courierID {
		t.Fatalf("expected courier id %d in unassign result, got %d", courierID, unassignResult.CourierId)
	}

	var cntAfterUnassign int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM delivery WHERE order_id=$1`, orderID).Scan(&cntAfterUnassign); err != nil {
		t.Fatalf("query delivery count after unassign: %v", err)
	}
	if cntAfterUnassign != 0 {
		t.Fatalf("expected 0 delivery rows after unassign, got %d", cntAfterUnassign)
	}

	courierAfterUnassign, err := courierUC.GetOneById(ctx, courierID)
	if err != nil {
		t.Fatalf("get courier after unassign: %v", err)
	}
	if courierAfterUnassign.Status != model.CourierStatusAvailable {
		t.Fatalf("expected available status after unassign, got %s", courierAfterUnassign.Status)
	}
}

func startPostgres(ctx context.Context, t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	const (
		user     = "test"
		password = "test"
		dbName   = "testdb"
	)

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     user,
			"POSTGRES_PASSWORD": password,
			"POSTGRES_DB":       dbName,
		},
		WaitingFor: wait.ForSQL("5432/tcp", "pgx", func(host string, port nat.Port) string {
			return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port.Port(), dbName)
		}).WithStartupTimeout(90 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("start container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("get container host: %v", err)
	}
	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatalf("get container port: %v", err)
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port.Port(), dbName)
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping database: %v", err)
	}

	cleanup := func() {
		pool.Close()
		terminateCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := container.Terminate(terminateCtx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}

	return pool, cleanup
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS couriers (
            id BIGSERIAL PRIMARY KEY,
            name TEXT NOT NULL,
            phone TEXT NOT NULL UNIQUE,
            status TEXT NOT NULL,
            transport_type TEXT NOT NULL DEFAULT 'on_foot',
            assignments_count BIGINT NOT NULL DEFAULT 0,
            created_at TIMESTAMP DEFAULT NOW(),
            updated_at TIMESTAMP DEFAULT NOW()
        );`,
		`CREATE TABLE IF NOT EXISTS delivery (
            id BIGSERIAL PRIMARY KEY,
            courier_id BIGINT NOT NULL,
            order_id VARCHAR(255) NOT NULL,
            assigned_at TIMESTAMP NOT NULL DEFAULT NOW(),
            deadline TIMESTAMP NOT NULL
        );`,
	}

	for _, stmt := range statements {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("exec migration: %w", err)
		}
	}

	return nil
}
