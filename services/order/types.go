package order

type Order struct {
	ID   string `json:"id"`
	UserID string `json:"user_id"`
	ProductID string `json:"product_id"`
	Quantity int `json:"quantity"`
	TotalCost float64 `json:"total_cost"`
	Status string `json:"status"`
	CreatedAt string `json:"created_at"`
}