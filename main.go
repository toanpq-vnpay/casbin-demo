package main

import (
	"fmt"
	"log"
	"net/http"

	"casbin-demo/database"
	"casbin-demo/handlers"
	"casbin-demo/middlewares"

	"casbin-demo/enforcer"

	"github.com/gorilla/mux"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
}

func main() {
	err := database.InitializeDatabase()
	if err != nil {
		log.Fatal(err)
	}

	err = enforcer.InitializeEnforcer()
	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()

	// Public route
	router.HandleFunc("/login", handlers.LoginHandler).Methods("POST")
	// router.HandleFunc("/users", handlers.RegisterHandler).Methods("POST")

	protected := router.PathPrefix("").Subrouter()
	protected.Use(middlewares.Authenticate())
	protected.Use(middlewares.Authorize(enforcer.GetEnforcer()))

	// Users management
	protected.HandleFunc("/users/me", handlers.GetCurrentUserInfo).Methods("GET")
	protected.HandleFunc("/users/{username}", handlers.GetUserByUsername).Methods("GET")
	protected.HandleFunc("/users/{username}", handlers.SoftDeleteUser).Methods("DELETE")
	protected.HandleFunc("/users", handlers.RegisterHandler).Methods("POST")
	protected.HandleFunc("/users/{username}/groups", handlers.GetUserGroups).Methods("GET")

	// Products management
	protected.HandleFunc("/products", handlers.GetAllProducts).Methods("GET")
	protected.HandleFunc("/products", handlers.CreateProduct).Methods("POST")
	protected.HandleFunc("/products/{productId}", handlers.UpdateProduct).Methods("PATCH")
	protected.HandleFunc("/products/{productId}", handlers.GetProductByID).Methods("GET")
	protected.HandleFunc("/products/{productId}", handlers.DeleteProduct).Methods("DELETE")
	protected.HandleFunc("/products/{productId}/stocks/in", handlers.AddStock).Methods("PATCH")
	protected.HandleFunc("/products/{productId}/stocks/out", handlers.RemoveStock).Methods("PATCH")

	// Report
	protected.HandleFunc("/reports/products", handlers.GetProductsReport).Methods("GET")

	// Group management
	protected.HandleFunc("/groups/{groupname}/users/{username}", handlers.AddUserToGroup).Methods("POST")
	protected.HandleFunc("/groups/{groupname}/users/{username}", handlers.RemoveUserFromGroup).Methods("DELETE")
	protected.HandleFunc("/groups/{groupname}/users", handlers.GetGroupUsers).Methods("GET")
	protected.HandleFunc("/groups/{groupname}", handlers.DeleteGroup).Methods("DELETE")

	// Permissions management
	protected.HandleFunc("/permissions/{name}", handlers.DeletePermissions).Methods("DELETE")
	protected.HandleFunc("/permissions", handlers.GrantPermission).Methods("POST")

	fmt.Println("Server started on port 8080")

	log.Fatal(http.ListenAndServe(":8080", router))
}
