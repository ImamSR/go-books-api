package books

import "time"

type Book struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Author    string    `json:"author"`
	Publisher string    `json:"publisher"`
	PageCount int       `json:"pageCount"`
	ReadPage  int       `json:"readPage"`
	Reading   bool      `json:"reading"`
	Finished  bool      `json:"finished"`
	InsertedAt time.Time `json:"insertedAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
