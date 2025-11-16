package users

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrEmailTaken = errors.New("email already used")
	ErrUserNotFound   = errors.New("user not found")
)

type Repo interface {
	Create(u *User) error
	FindByEmail(email string) (*User, error)
}

type pgRepo struct {
	pool *pgxpool.Pool
}

func NewPGRepo(pool *pgxpool.Pool) Repo {
	return &pgRepo{pool: pool}
}

func (r *pgRepo) Create (u *User) error {
	now := time.Now()
	u.CreatedAt, u.UpdatedAt = now, now
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO users (id, email, username, password, roles, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		u.ID, u.Email, u.Username, u.Password, u.Roles, u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()),"unique") {
			return ErrEmailTaken
		}
		return err
	}
	return nil
}

func (r *pgRepo) FindByEmail(email string) (*User, error) {
	row := r.pool.QueryRow(context.Background(),
		`SELECT id, email, username, password, roles, created_at, updated_at
		FROM users WHERE lower(email)=lower($1)`, email)

	var u User
	if err := row.Scan(&u.ID, &u.Email, &u.Username, &u.Password, &u.Roles, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, ErrUserNotFound
	}
	return &u, nil
}