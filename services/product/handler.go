// Handles the product service requests
package product

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/genin6382/go-grpc-microservices-benchmark/internal/config"
	"github.com/go-chi/chi/v5"
)

type ProductHandler struct {
	DB *sql.DB
	Config *config.Config
}

func (h *ProductHandler) HandleListProducts(w http.ResponseWriter, r *http.Request){
	products , err := ListProducts(h.DB,r.Context())

	if err!= nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
        return
	}

	if len(products) == 0 {
		http.Error(w, "No Products Available", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(products)
}

func (h * ProductHandler) HandleGetProductByID(w http.ResponseWriter, r *http.Request){
	id := chi.URLParam(r, "id")

	if id == "" || len(id) > 255 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}
	product, err := ListProductByID(h.DB, r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(product)
}

func (h *ProductHandler) HandleCreateProduct(w http.ResponseWriter, r *http.Request){
	var product Product
	err := json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	//Helper function to validate the product input - present in helper.go
	validationError := ValidateProductInput(product.Name, product.Description, product.Price, product.Stock)

	if validationError != nil {
		http.Error(w, validationError.Error(), http.StatusBadRequest)
		return
	}
	newProduct , err := CreateProduct(h.DB, r.Context(), product.Name, product.Description, product.Price, product.Stock)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newProduct)
}

func (h *ProductHandler) HandleUpdateProductDetails(w http.ResponseWriter, r *http.Request){
	id := chi.URLParam(r, "id")
	var product Product
	err := json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if id == "" || len(id) > 255 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}
	//Helper function to validate the product input - present in helper.go
	validationError := ValidateProductInput(product.Name, product.Description, product.Price, product.Stock)

	if validationError != nil {
		http.Error(w, validationError.Error(), http.StatusBadRequest)
		return
	}

	updatedProduct , err := UpdateProductDetails(h.DB, r.Context(), id, product.Name, product.Description, product.Price, product.Stock)
	
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedProduct)
}

func (h *ProductHandler) HandleUpdateProductStock(w http.ResponseWriter, r *http.Request){
	id := chi.URLParam(r, "id")
	var req struct {
		Delta *int `json:"delta"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if id == "" || len(id) > 255 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}
	if req.Delta == nil {
		http.Error(w, "Delta value is required", http.StatusBadRequest)
		return
	}
	updatedProduct, err := UpdateProductStock(h.DB, r.Context(), id, *req.Delta)
	if err != nil {
		if err.Error() == "Insufficient stock or Product not found" {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedProduct)
}


func (h *ProductHandler) HandleDeleteProduct(w http.ResponseWriter, r *http.Request){
	id := chi.URLParam(r, "id")
	
	if id == "" || len(id) > 255 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}
	
	product,err := DeleteProduct(h.DB, r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(product)
}
