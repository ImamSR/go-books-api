package books

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ImamSR/go-books-api/internal/util"
)

type pgStore struct {
	pool *pgxpool.Pool
}

func NewPGStore(pool *pgxpool.Pool) Store {
	return &pgStore{pool: pool}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (p *pgStore) Create(b *Book) (string, error) {
	if strings.TrimSpace(b.Name) == "" {
		return "", ErrInvalidName
	}
	if b.ReadPage > b.PageCount {
		return "", ErrReadPageTooBig
	}

	now := time.Now()
	finished := b.PageCount == b.ReadPage

	var id string
	var err error

	for attempt := 0; attempt < 3; attempt++ {
		id = util.RandomID()
		_, err = p.pool.Exec(context.Background(),
			`INSERT INTO books
			   (id, name, author, publisher, page_count, read_page, reading, finished, inserted_at, updated_at)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			id, b.Name, b.Author, b.Publisher, b.PageCount, b.readPageOrZero(),
			b.Reading, finished, now, now,
		)
		if err == nil {
			return id, nil
		}
		if !isUniqueViolation(err) {
			return "", err
		}
		
	}

	return "", err
}

func (p *pgStore) Get(id string) (*Book, error) {
	row := p.pool.QueryRow(context.Background(),
		`SELECT id, name, author, publisher, page_count, read_page, reading, finished, inserted_at, updated_at
         FROM books WHERE id = $1`, id)

	var b Book
	err := row.Scan(&b.ID, &b.Name, &b.Author, &b.Publisher, &b.PageCount, &b.ReadPage, &b.Reading, &b.Finished, &b.InsertedAt, &b.UpdatedAt)
	if err != nil {
		return nil, ErrNotFound
	}
	return &b, nil
}

func (p *pgStore) List(f Filter) ([]Book, int, error) {
  // base filter
  where := "WHERE 1=1"
  args := []any{}
  i := 1
  if strings.TrimSpace(f.Name) != "" {
    where += " AND lower(name) LIKE $" + strconv.Itoa(i)
    args = append(args, "%"+strings.ToLower(strings.TrimSpace(f.Name))+"%")
    i++
  }
  if f.Reading != nil {
    where += " AND reading = $" + strconv.Itoa(i)
    args = append(args, *f.Reading)
    i++
  }
  if f.Finished != nil {
    where += " AND finished = $" + strconv.Itoa(i)
    args = append(args, *f.Finished)
    i++
  }

  // total
  var total int
  if err := p.pool.QueryRow(context.Background(),
    "SELECT COUNT(*) FROM books "+where, args...,
  ).Scan(&total); err != nil {
    return nil, 0, err
  }

  // page
  limit := f.Limit
  if limit <= 0 { limit = 10 }
  offset := f.Offset
  if offset < 0 { offset = 0 }

  q := `
    SELECT id, name, author, publisher, page_count, read_page, reading, finished, inserted_at, updated_at
    FROM books ` + where + `
    ORDER BY inserted_at DESC
    LIMIT $` + strconv.Itoa(i) + ` OFFSET $` + strconv.Itoa(i+1)

  rows, err := p.pool.Query(context.Background(), q, append(args, limit, offset)...)
  if err != nil { return nil, 0, err }
  defer rows.Close()

  var out []Book
  for rows.Next() {
    var b Book
    if err := rows.Scan(&b.ID, &b.Name, &b.Author, &b.Publisher, &b.PageCount, &b.ReadPage, &b.Reading, &b.Finished, &b.InsertedAt, &b.UpdatedAt); err != nil {
      return nil, 0, err
    }
    out = append(out, b)
  }
  return out, total, nil
}

func (p *pgStore) Update(id string, patch Book) error {
	if strings.TrimSpace(patch.Name) == "" {
		return ErrInvalidName
	}
	if patch.ReadPage > patch.PageCount {
		return ErrReadPageTooBig
	}
	now := time.Now()
	finished := patch.PageCount == patch.ReadPage

	ct, err := p.pool.Exec(context.Background(),
		`UPDATE books
		   SET name=$1, author=$2, publisher=$3,
		       page_count=$4, read_page=$5,
		       reading=$6, finished=$7, updated_at=$8
		 WHERE id=$9`,
		patch.Name, patch.Author, patch.Publisher,
		patch.PageCount, patch.readPageOrZero(),
		patch.Reading, finished, now, id,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (p *pgStore) Delete(id string) error {
	ct, err := p.pool.Exec(context.Background(), `DELETE FROM books WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// helpers
func itoa(i int) string { return strconv.Itoa(i) }

func (b *Book) readPageOrZero() int {
	if b.ReadPage < 0 {
		return 0
	}
	return b.ReadPage
}
