package courier

import (
	"context"
	"errors"
	"strings"

	ipostgres "github.com/cdxy1/go-courier-service/internal/infra/postgres"
	"github.com/cdxy1/go-courier-service/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CourierRepository struct {
	conn *pgxpool.Pool
}

func NewCourierRepository(conn *pgxpool.Pool) *CourierRepository {
	return &CourierRepository{conn: conn}
}

func (c *CourierRepository) Create(ctx context.Context, courier *model.CourierModel) (int, error) {
	db := ipostgres.DBFromContext(ctx, c.conn)
	var id int
	query := `INSERT INTO couriers(name,phone,status,transport_type) VALUES ($1,$2,$3,$4) RETURNING id`

	err := db.QueryRow(ctx, query, courier.Name, courier.Phone, courier.Status, courier.TransportType).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			return 0, ErrPhoneExists
		}
		return 0, ErrDatabaseInternal
	}

	return id, nil
}

func (c *CourierRepository) Update(ctx context.Context, courier *model.CourierModel) error {
	db := ipostgres.DBFromContext(ctx, c.conn)
	query := `UPDATE couriers SET name=$1, phone=$2, status=$3, transport_type=$4, updated_at=NOW() WHERE id = $5 RETURNING id`
	var returnedId int
	if err := db.QueryRow(ctx, query, courier.Name, courier.Phone, courier.Status, courier.TransportType, courier.ID).Scan(&returnedId); err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			return ErrPhoneExists
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrCourierNotFound
		}
		return ErrDatabaseInternal
	}
	return nil
}

func (c *CourierRepository) GetOneById(ctx context.Context, id int) (*model.CourierModel, error) {
	db := ipostgres.DBFromContext(ctx, c.conn)
	var courier model.CourierModel
	query := `SELECT id, name, phone, status, transport_type, assignments_count FROM couriers WHERE id=$1`

	err := db.QueryRow(ctx, query, id).Scan(
		&courier.ID,
		&courier.Name,
		&courier.Phone,
		&courier.Status,
		&courier.TransportType,
		&courier.AssignmentsCount,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCourierNotFound
		}
		return nil, ErrDatabaseInternal
	}

	return &courier, nil
}

func (c *CourierRepository) GetAll(ctx context.Context) ([]*model.CourierModel, error) {
	db := ipostgres.DBFromContext(ctx, c.conn)
	query := `SELECT id, name, phone, status, transport_type, assignments_count FROM couriers`

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, ErrDatabaseInternal
	}
	defer rows.Close()

	var couriers []*model.CourierModel
	for rows.Next() {
		var courier model.CourierModel

		err := rows.Scan(
			&courier.ID,
			&courier.Name,
			&courier.Phone,
			&courier.Status,
			&courier.TransportType,
			&courier.AssignmentsCount,
		)
		if err != nil {
			return nil, ErrReadingData
		}
		couriers = append(couriers, &courier)
	}

	if err = rows.Err(); err != nil {
		return nil, ErrDatabaseInternal
	}

	if couriers == nil {
		couriers = []*model.CourierModel{}
	}

	return couriers, nil
}

func (c *CourierRepository) GetByStatus(ctx context.Context, status model.CourierStatus) (*model.CourierModel, error) {
	db := ipostgres.DBFromContext(ctx, c.conn)
	var courier model.CourierModel
	query := `SELECT id, name, phone, status, transport_type, assignments_count FROM couriers WHERE status=$1`

	err := db.QueryRow(ctx, query, status).Scan(&courier.ID,
		&courier.Name,
		&courier.Phone,
		&courier.Status,
		&courier.TransportType,
		&courier.AssignmentsCount,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCourierNotFound
		}
		return nil, ErrDatabaseInternal
	}

	return &courier, nil
}

func (c *CourierRepository) GetAvailableLeastDelivered(ctx context.Context) (*model.CourierModel, error) {
	db := ipostgres.DBFromContext(ctx, c.conn)
	var courier model.CourierModel
	query := `SELECT c.id, c.name, c.phone, c.status, c.transport_type, c.assignments_count
	          FROM couriers c
	          WHERE c.status = $1
	          ORDER BY c.assignments_count ASC, c.id ASC
	          LIMIT 1
	          FOR UPDATE SKIP LOCKED`

	err := db.QueryRow(ctx, query, model.CourierStatusAvailable).Scan(
		&courier.ID,
		&courier.Name,
		&courier.Phone,
		&courier.Status,
		&courier.TransportType,
		&courier.AssignmentsCount,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCourierNotFound
		}
		return nil, ErrDatabaseInternal
	}
	return &courier, nil
}

func (c *CourierRepository) UpdateStatus(ctx context.Context, status model.CourierStatus, id int) error {
	db := ipostgres.DBFromContext(ctx, c.conn)
	query := `UPDATE couriers SET status=$1 WHERE id=$2 RETURNING id`
	var returnedId int
	if err := db.QueryRow(ctx, query, status, id).Scan(&returnedId); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrCourierNotFound
		}
		return ErrDatabaseInternal
	}

	return nil
}

func (c *CourierRepository) MarkAssigned(ctx context.Context, id int) error {
	db := ipostgres.DBFromContext(ctx, c.conn)
	query := `UPDATE couriers SET status=$1, assignments_count=assignments_count+1 WHERE id=$2 RETURNING id`
	var returnedId int
	if err := db.QueryRow(ctx, query, model.CourierStatusBusy, id).Scan(&returnedId); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrCourierNotFound
		}
		return ErrDatabaseInternal
	}
	return nil
}
