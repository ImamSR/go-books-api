package users

import "time"

type User struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    Username  string    `json:"username"`  // <--- TAMBAH
    Password  string    `json:"-"`         // hashed
    Roles     []string  `json:"roles"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}

type RegisterInput struct {
    Email    string   `json:"email"`
    Username string   `json:"username,omitempty"` // <--- TAMBAH
    Password string   `json:"password"`
    Roles    []string `json:"roles,omitempty"`
}


type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
