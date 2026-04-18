package order

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/genin6382/go-grpc-microservices-benchmark/internal/config"
	internalmiddleware "github.com/genin6382/go-grpc-microservices-benchmark/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

type OrderHandler struct {
	DB     *sql.DB
	Config *config.Config
	UserClient *UserServiceClient
	ProductClient *ProductServiceClient
	CacheClient *redis.Client
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

	cacheKey := "order:" + id
	// Get from cache
	val , err := h.CacheClient.Get(context.Background(), cacheKey).Result()
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(val))
		return
	}
	//If cache miss, fetch from DB
	order, err := ListOrderByID(h.DB, r.Context(), id)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}
	//Encode order to JSON 
	orderJSON, err := json.Marshal(order)
	if err != nil {
		log.Errorf("Failed to marshal order: %v", err)
		http.Error(w, "Failed to process order data", http.StatusInternalServerError)
		return
	}
	// Set cache as seperate goroutine to avoid blocking response
	go func() {
		err := h.CacheClient.Set(context.Background(), cacheKey, orderJSON, 6*time.Hour).Err()
		if err != nil {
			log.Errorf("Failed to set cache for order %s: %v", id, err)
		}
	}()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(orderJSON)
}

func (h *OrderHandler) HandleGetOrdersByUserID(w http.ResponseWriter, r *http.Request) {
	// Extract user_id from context (set by VerifyToken middleware)
	userID, ok := r.Context().Value(internalmiddleware.UserIDKey).(string)
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

    // Check if the user exists
    userExists, err := h.UserClient.CheckUserExists(r.Context(), userID)
    if err != nil {
        log.Errorf("Error checking user exists: %v", err)
        http.Error(w, "Failed to check user", http.StatusInternalServerError)
        return
    }
    if !userExists {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }

	// Check if the product exists and get its price
	productInfo, err := h.ProductClient.GetProductInfo(r.Context(), req.ProductID)
	if err != nil {
		log.Errorf("Error fetching product info: %v", err)
		http.Error(w, "Failed to fetch product info", http.StatusInternalServerError)
		return
	}
	if productInfo == nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	if int(productInfo.Stock) < req.Quantity {
		http.Error(w, "Insufficient stock", http.StatusBadRequest)
		return
	}

	totalCost := float64(req.Quantity) * float64(productInfo.Price)

	// Update product stock
	success , err := h.ProductClient.UpdateStock(r.Context(),req.ProductID,int32(req.Quantity))

	if err != nil || !success {
		log.Errorf("Error updating stock: %v", err)
		http.Error(w, "Failed to update stock", http.StatusInternalServerError)
		return
	}
	
    order, err := CreateOrder(h.DB, r.Context(), userID, req.ProductID, req.Quantity, totalCost)
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
	//Invalidate cache asynchronously
	go func() {
		cacheKey := "order:" + id
		err := h.CacheClient.Del(context.Background(), cacheKey).Err()
		if err != nil {
			log.Errorf("Failed to invalidate cache for key %s: %v", cacheKey, err)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}
