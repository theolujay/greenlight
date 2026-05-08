package data

import (
	"database/sql"
	"errors"
	"time"

	"github.com/theolujay/greenlight/internal/validator"

	"github.com/lib/pq"
)

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitzero"`    // Release year
	Runtime   Runtime   `json:"runtime,omitzero"` // (in minutes)
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version"` // Starts at 1 and increments ech htime the movie information is updated
}

type MovieModel struct {
	DB *sql.DB
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")
	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")
	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}

// The Insert() method accepts a pointer to a movie struct,
// which should contain the data for the new record.
func (m MovieModel) Insert(movie *Movie) error {
	// The RETURNING clause is PostgreSQL-specific
	// (it's not part of the SQL standard)
	query := `
		INSERT INTO movies (title, year, runtime, genres)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version
	`

	// Create an args slice containing the values for the placeholder paramters
	// from the movie struct. Declaring this slice immediately next to the SQL
	// query helps to make it nice and clear *what values are being used where*
	// in the query.
	//
	// NOTE:
	//
	// Behind the scenes, the pq.Array() adapter takes our []string slice and
	// converts it to a pq.StringArray type. In turn, the pq.StringArray type
	// implements the driver.Valuer and sql.Scanner interfaces necessary to
	// translate our native []string slice to and from a value that the
	// PostgreSQL database can understand and store in a text[] array column.
	// It can also be used in the same way with []bool, []byte, []int32,
	// []int64, []float32 and []float64 slices.
	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	// Use the QueryRow() method instead, since the query (because of RETURNING)
	// is returning a single row of data... then call Scan to update the returned
	// values (the system-generated ones) at the location the paramters point to.
	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	var movie Movie
	query := `
		SELECT id, created_at, title, year, runtime, genres, version
		FROM movies
		WHERE id = $1
	`
	err := m.DB.QueryRow(query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &movie, nil
}

func (m MovieModel) Update(movie *Movie) error {
	query := `
		UPDATE movies
		SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
		WHERE id = $5
		RETURNING version
	`
	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
	}
	return m.DB.QueryRow(query, args...).Scan(&movie.Version)
}

func (m MovieModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := "DELETE FROM movies WHERE id = $1"

	res, err := m.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
