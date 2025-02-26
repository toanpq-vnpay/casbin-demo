package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"casbin-demo/database"
	"casbin-demo/middlewares"
	"casbin-demo/models"

	"github.com/gorilla/mux"
)

// AddStock handles increasing product stock
func AddStock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.Atoi(vars["productId"])
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var req models.StockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Quantity <= 0 {
		http.Error(w, "Quantity must be positive", http.StatusBadRequest)
		return
	}

	claims, ok := r.Context().Value(middlewares.ClaimsKey).(*models.Claims)
	if !ok {
		http.Error(w, "Error retrieving user info", http.StatusInternalServerError)
		return
	}

	op := models.Operation{
		ProductID: productID,
		Quantity:  req.Quantity,
		UserID:    claims.UserID,
	}

	if err := database.AddProductStock(op); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Stock added successfully"})
}

// RemoveStock handles decreasing product stock
func RemoveStock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.Atoi(vars["productId"])
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var req models.StockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Quantity <= 0 {
		http.Error(w, "Quantity must be positive", http.StatusBadRequest)
		return
	}

	claims, ok := r.Context().Value(middlewares.ClaimsKey).(*models.Claims)
	if !ok {
		http.Error(w, "Error retrieving user info", http.StatusInternalServerError)
		return
	}

	op := models.Operation{
		ProductID: productID,
		Quantity:  req.Quantity,
		UserID:    claims.UserID,
	}

	if err := database.RemoveProductStock(op); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Stock removed successfully"})
}

// AdjustStocks handles adjusting product
func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.Atoi(vars["productId"])
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var req models.ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	claims, ok := r.Context().Value(middlewares.ClaimsKey).(*models.Claims)
	if !ok {
		http.Error(w, "Error retrieving user info", http.StatusInternalServerError)
		return
	}

	op := models.Operation{
		ProductID: productID,
		Quantity:  req.Quantity,
		Reason:    req.Reason,
		UserID:    claims.UserID,
	}

	if err := database.UpdateProduct(op); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Stock adjusted successfully"})
}

func GetAllProducts(w http.ResponseWriter, r *http.Request) {
	products, err := database.GetAllProducts()
	if err != nil {
		if err == sql.ErrNoRows {
			// Return empty array instead of null when no products found
			json.NewEncoder(w).Encode([]models.Product{})
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response []models.ProductResponse
	for _, p := range products {
		response = append(response, models.ProductResponse{
			ID:        p.ID,
			Name:      p.Name,
			Quantity:  p.Quantity,
			UnitPrice: p.UnitPrice,
		})
	}

	json.NewEncoder(w).Encode(response)
}

func CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req models.ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	claims, ok := r.Context().Value(middlewares.ClaimsKey).(*models.Claims)
	if !ok {
		http.Error(w, "Error retrieving user info", http.StatusInternalServerError)
		return
	}

	// Validate request
	if req.Name == "" {
		http.Error(w, "Product name is required", http.StatusBadRequest)
		return
	}
	if req.UnitPrice <= 0 {
		http.Error(w, "Unit price must be greater than 0", http.StatusBadRequest)
		return
	}
	if req.Quantity < 0 {
		http.Error(w, "Quantity cannot be negative", http.StatusBadRequest)
		return
	}

	op := models.Operation{
		Name:      req.Name,
		UnitPrice: req.UnitPrice,
		Quantity:  req.Quantity,
		UserID:    claims.UserID,
		Type:      models.OperationAddProduct,
		Reason:    "Initial product creation",
	}

	if err := database.CreateProduct(op); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.Atoi(vars["productId"])
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	claims, ok := r.Context().Value(middlewares.ClaimsKey).(*models.Claims)
	if !ok {
		http.Error(w, "Error retrieving user info", http.StatusInternalServerError)
		return
	}

	op := models.Operation{
		ProductID: productID,
		UserID:    claims.UserID,
		Reason:    "Product deletion",
		Type:      models.OperationDelProduct,
	}

	if err := database.DeleteProduct(op); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func GetProductByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.Atoi(vars["productId"])
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	product, err := database.GetProductByID(productID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := models.ProductResponse{
		ID:        product.ID,
		Name:      product.Name,
		Quantity:  product.Quantity,
		UnitPrice: product.UnitPrice,
	}

	json.NewEncoder(w).Encode(response)
}

func GetProductsReport(w http.ResponseWriter, r *http.Request) {
	reports, err := database.GetProductsReport()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(reports); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
