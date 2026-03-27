// Performs database operations for the product service
package product

import (
	"database/sql"
	"context"
	"fmt"
)

func ListProducts(db *sql.DB, ctx context.Context) ([]Product,error){
	records , err := db.QueryContext(ctx, "SELECT * FROM products")

	if err!= nil {
		return nil , fmt.Errorf("Failed to Query Products: %v",err)
	}

	defer records.Close()

	var products []Product
	for records.Next(){
		var prod Product
		err := records.Scan(&prod.Id, &prod.Name, &prod.Description , &prod.Price , &prod.Stock , &prod.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product record: %w", err)
		}
		products = append(products, prod)
	}
	return products, records.Err()
}

func ListProductByID(db *sql.DB, ctx context.Context, id string) (*Product,error){
	record := db.QueryRowContext(ctx , "SELECT * FROM products WHERE id = $1",id)

	var prod Product
	if err := record.Scan(&prod.Id, &prod.Name, &prod.Description , &prod.Price , &prod.Stock , &prod.CreatedAt); err != nil {
		return nil, fmt.Errorf("Failed to scan product record: %w", err)
	}
	return &prod, nil
}

func CreateProduct(db *sql.DB , ctx context.Context , name string, description string, price float64, stock int) (*Product,error){
	record := db.QueryRowContext(ctx,"INSERT INTO products(name, description, price, stock) VALUES ($1, $2, $3, $4) RETURNING id, name , description , price , stock , created_at", name, description, price, stock)
	var prod Product
	if err := record.Scan(&prod.Id, &prod.Name, &prod.Description, &prod.Price, &prod.Stock, &prod.CreatedAt); err != nil {
		return nil, fmt.Errorf("Failed to scan product record: %w", err)
	}
	return &prod, nil
}

func UpdateProductDetails(db *sql.DB, ctx context.Context, id string, name string, description string, price float64, stock int) (*Product,error){
	record := db.QueryRowContext(ctx,"UPDATE products SET name = $1, description = $2, price = $3, stock = $4 WHERE id = $5 RETURNING id, name , description , price , stock , created_at", name, description, price, stock, id)
	var prod Product
	if err := record.Scan(&prod.Id, &prod.Name, &prod.Description, &prod.Price, &prod.Stock, &prod.CreatedAt); err != nil {
		return nil, fmt.Errorf("Failed to scan product record: %w", err)
	}
	return &prod, nil
}

func UpdateProductStock(db *sql.DB, ctx context.Context, id string, delta int) (*Product,error){
	record := db.QueryRowContext(ctx,"UPDATE products SET stock = stock + $1 WHERE id = $2 AND (stock + $1) >= 0 RETURNING id, stock , created_at", delta, id)
	var prod Product
	if err := record.Scan(&prod.Id, &prod.Stock, &prod.CreatedAt); err != nil {
		return nil, fmt.Errorf("Insufficient stock or Product not found")
	}
	return &prod, nil
}

func DeleteProduct(db *sql.DB , ctx context.Context , id string) (*Product,error){
	record := db.QueryRowContext(ctx,"DELETE FROM products WHERE id = $1 RETURNING id, name, description, price, stock, created_at", id)

	var prod Product
	if err := record.Scan(&prod.Id, &prod.Name, &prod.Description , &prod.Price , &prod.Stock , &prod.CreatedAt); err != nil {
		return nil, fmt.Errorf("Failed to scan product record: %w", err)
	}
	return &prod, nil
}

