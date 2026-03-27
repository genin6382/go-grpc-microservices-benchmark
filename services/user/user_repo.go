// Handles SQL queries related to users
package user

import (
	"database/sql"
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func ListUsers(db *sql.DB, ctx context.Context) ([]User, error) {
	records, err := db.QueryContext(ctx, "SELECT id, name , created_at FROM users")
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer records.Close()

	var users []User
	for records.Next() {
		var user User
		err := records.Scan(&user.Id, &user.Name, &user.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user record: %w", err)
		}
		users = append(users, user)
	}
	return users, records.Err() // Check for errors from iterating over rows
}

func ListUserByID(db *sql.DB, ctx context.Context, id string) (*User, error) {
	var user User
	record := db.QueryRowContext(ctx, "SELECT id, name , created_at FROM users WHERE id = $1", id)

	if err := record.Scan(&user.Id, &user.Name, &user.CreatedAt); err != nil {
        return nil, err 
    }
    return &user, nil
}

func CreateUser(db *sql.DB , ctx context.Context, name string, password string) (*User, error) {
	bytes , err := bcrypt.GenerateFromPassword([] byte(password), bcrypt.DefaultCost)

	if err != nil {
		return nil, fmt.Errorf("CreateUser hash: %w", err)
	}

	encryptedPassword := string(bytes)
	var user User
	record := db.QueryRowContext(ctx, "INSERT INTO users (name , password) VALUES ($1, $2) RETURNING id, name, created_at", name, encryptedPassword)

	if err := record.Scan(&user.Id, &user.Name, &user.CreatedAt); err != nil {
		return nil, fmt.Errorf("CreateUser scan: %w", err)
	}
	return &user, nil
}

func DeleteUser(db *sql.DB, ctx context.Context, id string) (*User, error) {
    var user User
    row := db.QueryRowContext(ctx,
        "DELETE FROM users WHERE id = $1 RETURNING id, name, created_at", id)
    if err := row.Scan(&user.Id, &user.Name, &user.CreatedAt); err != nil {
        return nil, err
    }
    return &user, nil
}

func VerifyPassword(db *sql.DB, ctx context.Context, name string, password string) (string, error) {
	var userID string
	var encryptedPassword string
	
	row := db.QueryRowContext(ctx, "SELECT id, password FROM users WHERE name = $1", name)
	if err := row.Scan(&userID, &encryptedPassword); err != nil {
		return "", fmt.Errorf("user not found")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(encryptedPassword), []byte(password)); err != nil {
		return "", fmt.Errorf("Password does not match") 
	}
	
	return userID, nil

}