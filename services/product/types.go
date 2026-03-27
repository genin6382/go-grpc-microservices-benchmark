package product

import "time"

type Product struct {
	Id string `json:"id"`
	Name string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Price float64 `json:"price,omitempty"`
	Stock int `json:"stock"`
	CreatedAt time.Time `json:"created_at"`
}