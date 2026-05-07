package data

import (
	"time"
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
