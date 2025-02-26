package database

import (
	"casbin-demo/models"
	"database/sql"
	"fmt"
	"time"
)

var db *sql.DB

func CreateUser(username, password string) error {
	_, err := db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", username, password)
	return err
}

func GetPasswordForUser(username string) (string, error) {
	var hashedPassword string
	err := db.QueryRow("SELECT password FROM users WHERE username=$1 AND deleted_at IS NULL", username).Scan(&hashedPassword)
	return hashedPassword, err
}

func GetUserByUsername(username string) (models.User, error) {
	var user models.User
	err := db.QueryRow("SELECT id, username, password FROM users WHERE username=$1 AND deleted_at IS NULL", username).Scan(&user.ID, &user.Username, &user.Password)
	return user, err
}

func SoftDeleteUser(username string) error {
	_, err := db.Exec("UPDATE users SET deleted_at = $1 WHERE username = $2 AND deleted_at IS NULL", time.Now(), username)
	return err
}

// AddProductStock increases product stock
func AddProductStock(op models.Operation) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
        UPDATE products 
        SET quantity = quantity + $1, 
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $2`, op.Quantity, op.ProductID)
	if err != nil {
		return fmt.Errorf("failed to update stock: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	if err := recordOperation(tx, op.ProductID, models.OperationAddStock, "Import stock", op.UserID); err != nil {
		return err
	}

	return tx.Commit()
}

// RemoveProductStock decreases product stock
func RemoveProductStock(op models.Operation) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	var currentQuantity int
	err = tx.QueryRow("SELECT quantity FROM products WHERE id = $1", op.ProductID).Scan(&currentQuantity)
	if err == sql.ErrNoRows {
		return fmt.Errorf("product not found")
	}
	if err != nil {
		return fmt.Errorf("database error: %v", err)
	}

	if currentQuantity < op.Quantity {
		return fmt.Errorf("insufficient stock")
	}

	_, err = tx.Exec(`
        UPDATE products 
        SET quantity = quantity - $1,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $2`, op.Quantity, op.ProductID)
	if err != nil {
		return fmt.Errorf("failed to update stock: %v", err)
	}

	if err := recordOperation(tx, op.ProductID, models.OperationRemoveStock, "Deliver stock", op.UserID); err != nil {
		return err
	}

	return tx.Commit()
}

func UpdateProduct(op models.Operation) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Build dynamic UPDATE query
	query := "UPDATE products SET updated_at = CURRENT_TIMESTAMP"
	params := make([]interface{}, 0)
	paramCount := 1

	if op.Name != "" {
		query += fmt.Sprintf(", name = $%d", paramCount)
		params = append(params, op.Name)
		paramCount++
	}

	if op.UnitPrice > 0 {
		query += fmt.Sprintf(", unit_price = $%d", paramCount)
		params = append(params, op.UnitPrice)
		paramCount++
	}

	if op.Quantity >= 0 {
		query += fmt.Sprintf(", quantity = $%d", paramCount)
		params = append(params, op.Quantity)
		paramCount++
	}

	query += fmt.Sprintf(" WHERE id = $%d", paramCount)
	params = append(params, op.ProductID)

	result, err := tx.Exec(query, params...)
	if err != nil {
		return fmt.Errorf("failed to update product: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	if err := recordOperation(tx, op.ProductID, models.OperationAdjustProduct, op.Reason, op.UserID); err != nil {
		return err
	}

	return tx.Commit()
}

func GetAllProducts() ([]models.Product, error) {
	query := `SELECT id, name, unit_price, quantity
              FROM products WHERE deleted_at IS NULL`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		err := rows.Scan(&p.ID, &p.Name, &p.UnitPrice, &p.Quantity)
		if err != nil {
			return nil, err

		}
		products = append(products, p)
	}
	return products, nil
}

func GetProductByID(id int) (models.Product, error) {
	var product models.Product
	query := `SELECT id, name, unit_price, quantity 
              FROM products WHERE id = $1 AND deleted_at IS NULL`
	err := db.QueryRow(query, id).Scan(
		&product.ID, &product.Name, &product.UnitPrice, &product.Quantity)
	return product, err
}

func CreateProduct(op models.Operation) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	var productID int
	err = tx.QueryRow(`
        INSERT INTO products (name, unit_price, quantity, created_at, updated_at) 
        VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
        RETURNING id`,
		op.Name, op.UnitPrice, op.Quantity).Scan(&productID)
	if err != nil {
		return fmt.Errorf("failed to add product: %v", err)
	}

	// Record the operation
	if err := recordOperation(tx, productID, op.Type, op.Reason, op.UserID); err != nil {
		return err
	}

	return tx.Commit()
}

func DeleteProduct(op models.Operation) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Soft delete the product
	result, err := tx.Exec(`
        UPDATE products 
        SET deleted_at = CURRENT_TIMESTAMP 
        WHERE id = $1 AND deleted_at IS NULL`,
		op.ProductID)
	if err != nil {
		return fmt.Errorf("failed to delete product: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %v", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	// Record the operation
	if err := recordOperation(tx, op.ProductID, op.Type, op.Reason, op.UserID); err != nil {
		return err
	}

	return tx.Commit()
}

func GetProductsReport() ([]models.OperationReport, error) {
	query := `
        SELECT 
            o.id,
            o.product_id,
            p.name as product_name,
            o.type,
            o.reason,
            o.created_by,
            u.username as created_by_username,
            o.created_at
        FROM operations o
        JOIN products p ON o.product_id = p.id
        JOIN users u ON o.created_by = u.id
        ORDER BY o.created_at DESC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query operations: %v", err)
	}
	defer rows.Close()

	var reports []models.OperationReport
	for rows.Next() {
		var report models.OperationReport
		err := rows.Scan(
			&report.ID,
			&report.ProductID,
			&report.ProductName,
			&report.Type,
			&report.Reason,
			&report.CreatedBy,
			&report.CreatedByUsername,
			&report.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan operation row: %v", err)
		}
		reports = append(reports, report)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating operation rows: %v", err)
	}

	return reports, nil
}

// recordOperation records an operation in the database
func recordOperation(tx *sql.Tx, productID int, opType models.OperationType, reason string, userID int) error {
	_, err := tx.Exec(`
        INSERT INTO operations (product_id, type, reason, created_by)
        VALUES ($1, $2, $3, $4)`,
		productID, opType, reason, userID)
	if err != nil {
		return fmt.Errorf("failed to record operation: %v", err)
	}
	return nil
}
