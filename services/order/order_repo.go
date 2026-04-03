// Performs database operations for the order service
package order

import (
	"context"
	"database/sql"
	"fmt"
)

func ListOrders(db *sql.DB, ctx context.Context) ([]Order, error) {
	records, err := db.QueryContext(ctx,
		`SELECT * FROM orders`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer records.Close()

	var orders []Order
	for records.Next() {
		var ord Order
		if err := records.Scan(&ord.ID, &ord.UserID, &ord.ProductID, &ord.Quantity,&ord.TotalCost, &ord.Status, &ord.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan order record: %w", err)
		}
		orders = append(orders, ord)
	}
	return orders, records.Err()
}

func ListOrderByID(db *sql.DB, ctx context.Context, id string) (*Order, error) {
	record := db.QueryRowContext(ctx,
		`SELECT * FROM orders WHERE id = $1`, id,
	)
	var ord Order
	if err := record.Scan(&ord.ID, &ord.UserID, &ord.ProductID, &ord.Quantity, &ord.TotalCost, &ord.Status, &ord.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to scan order record: %w", err)
	}
	return &ord, nil
}

func ListOrdersByUserID(db *sql.DB, ctx context.Context, userID string) ([]Order, error) {
	records, err := db.QueryContext(ctx,
		`SELECT * FROM orders WHERE user_id = $1`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer records.Close()

	var orders []Order
	for records.Next() {
		var ord Order
		if err := records.Scan(&ord.ID, &ord.UserID, &ord.ProductID, &ord.Quantity, &ord.TotalCost, &ord.Status, &ord.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan order record: %w", err)
		}
		orders = append(orders, ord)
	}
	return orders, records.Err()
}


func CreateOrder(db *sql.DB, ctx context.Context, userID string, productID string, quantity int, totalCost float64) (*Order, error) {
	//Place Order
	var ord Order
	err := db.QueryRowContext(ctx,
		`INSERT INTO orders (user_id, product_id, quantity, total_cost, status)
		 VALUES ($1, $2, $3, $4, 'confirmed')
		 RETURNING id, user_id, product_id, quantity, total_cost, status, created_at`,
		userID, productID, quantity, totalCost,
	).Scan(&ord.ID, &ord.UserID, &ord.ProductID, &ord.Quantity, &ord.TotalCost, &ord.Status, &ord.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}
	return &ord, nil
}


func UpdateOrderStatus(db *sql.DB, ctx context.Context, id string, status string) (*Order, error) {
	record := db.QueryRowContext(ctx,
		`UPDATE orders SET status = $1 WHERE id = $2
		 RETURNING id, user_id, product_id, quantity, total_cost, status, created_at`,
		status, id,
	)

	var ord Order
	if err := record.Scan(&ord.ID, &ord.UserID, &ord.ProductID, &ord.Quantity, &ord.TotalCost, &ord.Status, &ord.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to scan order record: %w", err)
	}
	return &ord, nil
}

func DeleteOrder(db *sql.DB, ctx context.Context, id string) (*Order, error) {
	record := db.QueryRowContext(ctx,
		`DELETE FROM orders WHERE id = $1
		 RETURNING id, user_id, product_id, quantity, total_cost, status, created_at`, id,
	)

	var ord Order
	if err := record.Scan(&ord.ID, &ord.UserID, &ord.ProductID, &ord.Quantity, &ord.TotalCost, &ord.Status, &ord.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to scan order record: %w", err)
	}
	return &ord, nil
}
