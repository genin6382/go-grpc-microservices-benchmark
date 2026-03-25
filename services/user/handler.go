// Handler for user-related requests
package user

import (
	"database/sql"
	"encoding/json"
	"net/http"
	internaljwt "github.com/genin6382/go-grpc-microservices-benchmark/internal/jwt"
	"github.com/genin6382/go-grpc-microservices-benchmark/internal/config"
	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	DB *sql.DB
	Config *config.Config
}
func (h *UserHandler) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := ListUsers(h.DB, r.Context())

	if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
	if len(users) == 0 {
		http.Error(w, "No users found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) HandleGetUserByID(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL
	id := chi.URLParam(r, "id")

	if id == "" || len(id) > 255 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	user, err := ListUserByID(h.DB, r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name=="" || req.Password=="" {
		http.Error(w, "Name and Password are required",  http.StatusBadRequest)
		return 
	}
	user, err := CreateUser(h.DB, r.Context(), req.Name, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	if id == "" || len(id) > 255 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user,err := DeleteUser(h.DB, r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, err := VerifyPassword(h.DB, r.Context(), req.Name, req.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	tokenString, err := internaljwt.GenerateJWT(userID, []byte(h.Config.JWTSecretKey))
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
	})
}
