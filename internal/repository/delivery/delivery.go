package delivery

import (
	"context"
	"errors"

	ipostgres "github.com/cdxy1/go-courier-service/internal/infra/postgres"
	"github.com/cdxy1/go-courier-service/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DeliveryRepository struct {
	conn *pgxpool.Pool
}

func NewDeliveryRepository(conn *pgxpool.Pool) *DeliveryRepository {
	return &DeliveryRepository{conn: conn}
}

func (d *DeliveryRepository) Create(ctx context.Context, delivery *model.DeliveryModel) error {
	db := ipostgres.DBFromContext(ctx, d.conn)
	query := `INSERT INTO delivery(courier_id,order_id,deadline) VALUES ($1,$2,$3)`
	if err := db.Exec(ctx, query, delivery.CourierId, delivery.OrderId, delivery.Deadline); err != nil {
		return ErrDatabaseInternal
	}
	return nil
}

func (d *DeliveryRepository) Delete(ctx context.Context, orderId string) (int, error) {
	db := ipostgres.DBFromContext(ctx, d.conn)
	var courierId int
	query := `DELETE FROM delivery WHERE order_id=$1 RETURNING courier_id`
	if err := db.QueryRow(ctx, query, orderId).Scan(&courierId); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrDeliveryNotFound
		}
		return 0, ErrDatabaseInternal
	}
	return courierId, nil
}

func (d *DeliveryRepository) GetCourierID(ctx context.Context, orderId string) (int, error) {
	db := ipostgres.DBFromContext(ctx, d.conn)
	var courierId int
	query := `SELECT courier_id FROM delivery WHERE order_id=$1`
	if err := db.QueryRow(ctx, query, orderId).Scan(&courierId); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrDeliveryNotFound
		}
		return 0, ErrDatabaseInternal
	}
	return courierId, nil
}

func (d *DeliveryRepository) ReleaseExpiredCouriers(ctx context.Context) (int, error) {
	db := ipostgres.DBFromContext(ctx, d.conn)
	query := `WITH expired AS (SELECT DISTINCT courier_id FROM delivery WHERE deadline < NOW())
			  UPDATE couriers SET status=$1
			  WHERE status=$2 AND id IN (SELECT courier_id FROM expired)
				AND NOT EXISTS(SELECT 1 FROM delivery WHERE courier_id = couriers.id AND deadline >= NOW())
			  RETURNING id`

	rows, err := db.Query(ctx, query, model.CourierStatusAvailable, model.CourierStatusBusy)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
			return 0, ErrDeliveryTableMissing
		}
		return 0, ErrDatabaseInternal
	}
	defer rows.Close()

	var updated int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return 0, ErrDatabaseInternal
		}
		updated++
	}
	if err := rows.Err(); err != nil {
		return 0, ErrDatabaseInternal
	}
	return updated, nil
}
