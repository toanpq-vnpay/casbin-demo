package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"casbin-demo/models"

	"casbin-demo/database"

	"casbin-demo/enforcer"
	"casbin-demo/middlewares"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

var (
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	json.NewDecoder(r.Body).Decode(&user)

	if len(user.Username) < 4 || len(user.Username) > 32 {
		http.Error(w, "Username must be at least 4 characters and at most 32 characters", http.StatusBadRequest)
		return
	}

	if len(user.Password) < 6 || len(user.Password) > 32 {
		http.Error(w, "Password must be at least 6 characters and at most 32 characters", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error encrypting password", http.StatusInternalServerError)
		return
	}

	err = database.CreateUser(user.Username, string(hashedPassword))
	if err != nil {
		fmt.Println("Error registering user", err)
		http.Error(w, "Error registering user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	json.NewDecoder(r.Body).Decode(&user)

	fmt.Println("Logging in user", user.Username)
	dbUser, err := database.GetUserByUsername(user.Username)

	if err != nil {
		fmt.Println("Error getting user", err)
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password))
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(time.Hour * 24)
	claims := &models.Claims{
		Username: user.Username,
		UserID:   dbUser.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

// getUserInfo is a helper function that gets user info and roles
func getUserInfo(username string) (*models.UserResponse, error) {
	user, err := database.GetUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	// Get the enforcer instance and roles
	e := enforcer.GetEnforcer()
	roles, err := e.GetRolesForUser(username)
	if err != nil {
		return nil, fmt.Errorf("error getting user groups: %v", err)
	}

	return &models.UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Groups:   roles,
	}, nil
}

func GetCurrentUserInfo(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middlewares.ClaimsKey).(*models.Claims)
	if !ok {
		http.Error(w, "Error retrieving user info", http.StatusInternalServerError)
		return
	}

	response, err := getUserInfo(claims.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(response)
}

func GetUserByUsername(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	response, err := getUserInfo(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(response)
}

func SoftDeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	// Check if user exists first
	_, err := database.GetUserByUsername(username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Get the enforcer instance
	e := enforcer.GetEnforcer()

	// Remove user from Casbin (this removes all roles and policies related to the user)
	_, err = e.DeleteUser(username)
	if err != nil {
		http.Error(w, "Error removing user from authorization system", http.StatusInternalServerError)
		return
	}

	err = database.SoftDeleteUser(username)
	if err != nil {
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func GetUserGroups(w http.ResponseWriter, r *http.Request) {
	// Get username from URL parameters
	vars := mux.Vars(r)
	username := vars["username"]

	// Get the enforcer instance
	e := enforcer.GetEnforcer()

	// Get all roles/groups for the user
	roles, err := e.GetImplicitRolesForUser(username)
	if err != nil {
		http.Error(w, "Error getting user groups: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create response object
	response := struct {
		Username string   `json:"username"`
		Groups   []string `json:"groups"`
	}{
		Username: username,
		Groups:   roles,
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
