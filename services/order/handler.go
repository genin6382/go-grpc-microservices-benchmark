package order

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/genin6382/go-grpc-microservices-benchmark/internal/config"
	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
)

type OrderHandler struct {
	DB     *sql.DB
	Config *config.Config
}

func (h *OrderHandler) HandleListOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := ListOrders(h.DB, r.Context())
	if err != nil {
		log.Errorf("Failed to list orders: %v", err)
		http.Error(w, "Failed to fetch orders", http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		http.Error(w, "No orders found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)
}

func (h *OrderHandler) HandleGetOrderByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	order, err := ListOrderByID(h.DB, r.Context(), id)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}


	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) HandleGetOrdersByUserID(w http.ResponseWriter, r *http.Request) {
	// Extract user_id from context (set by VerifyToken middleware)
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	orders, err := ListOrdersByUserID(h.DB, r.Context(), userID)
	if err != nil {
		log.Errorf("Failed to fetch orders for user %s: %v", userID, err)
		http.Error(w, "Failed to fetch orders", http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		http.Error(w, "No orders found for this user", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)
}

func (h *OrderHandler) HandleUpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	order, err := UpdateOrderStatus(h.DB, r.Context(), id, req.Status)
	if err != nil {
		log.Errorf("Failed to update order status: %v", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) HandleCreateOrder(w http.ResponseWriter, r *http.Request) {
    var req struct {
        ProductID string `json:"product_id"`
        Quantity  int    `json:"quantity"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    userID, ok := r.Context().Value("user_id").(string)
    if !ok || userID == "" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    if req.Quantity <= 0 {
        http.Error(w, "Quantity must be greater than zero", http.StatusBadRequest)
        return
    }

    order, err := CreateOrder(h.DB, r.Context(), userID, req.ProductID, req.Quantity)
    if err != nil {
        log.Errorf("ERROR: Order Error: %v\n", err) 
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) HandleDeleteOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	order, err := DeleteOrder(h.DB, r.Context(), id)
	if err != nil {
		log.Errorf("Failed to delete order: %v", err)
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}
