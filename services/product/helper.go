package product

import "fmt"

func ValidateProductInput(name string, description string, price float64, stock int) error {
	if name == "" || len(name) > 255 {
		return fmt.Errorf("Invalid product name")
	}
	if description == "" || len(description) > 1000 {
		return fmt.Errorf("Invalid product description")
	}
	if price <= 5 {
		return fmt.Errorf("Invalid product price")
	}
	if stock < 0 {
		return fmt.Errorf("Invalid product stock")
	}
	return nil
}
