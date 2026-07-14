package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/s-usmonalizoda25/userServiceCinemaProject/internal/models"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateUser(ctx context.Context, u *models.User) (int64, error) {
	var id int64
	query := `INSERT INTO users (name, email, phone, password, age) 
              VALUES ($1, $2, $3, $4, $5) RETURNING id`

	err := r.pool.QueryRow(ctx, query, u.Name, u.Email, u.Phone, u.Password, u.Age).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, fmt.Errorf("email already exists")
		}
		return 0, err
	}

	return id, nil
}

func (r *Repository) GetUser(ctx context.Context, id int64) (*models.User, error) {
	u := &models.User{ID: id}
	query := `SELECT name, email, phone, age FROM users WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(&u.Name, &u.Email, &u.Phone, &u.Age)
	return u, err
}

func (r *Repository) UpdateUser(ctx context.Context, u *models.User) error {
	query := `UPDATE users SET name = $1, phone = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, u.Name, u.Phone, u.ID)
	return err
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	u := &models.User{}
	query := `SELECT id, name, email, phone, password, age FROM users WHERE email = $1`
	err := r.pool.QueryRow(ctx, query, email).Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.Password, &u.Age)
	return u, err
}
