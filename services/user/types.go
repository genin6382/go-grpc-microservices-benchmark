package user

import (
	"time"
)

// User struct represents a user in the system
type User struct {
	Id string `json:"id"`
	Name string `json:"name"`
	Password string `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}