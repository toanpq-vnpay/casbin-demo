package models

type OperationType string

const (
	OperationAddStock      OperationType = "ADD_STOCK"
	OperationRemoveStock   OperationType = "REMOVE_STOCK"
	OperationAdjustProduct OperationType = "ADJUST_PRODUCT"
	OperationAddProduct    OperationType = "CREATE_PRODUCT"
	OperationDelProduct    OperationType = "DELETE_PRODUCT"
)

type Operation struct {
	ProductID int
	Quantity  int
	Name      string
	UnitPrice float64
	Reason    string
	Type      OperationType
	UserID    int
}

// First, add this struct to your models package
type OperationReport struct {
	ID                int           `json:"id"`
	ProductID         int           `json:"product_id"`
	ProductName       string        `json:"product_name"`
	Type              OperationType `json:"operation_type"`
	Reason            string        `json:"reason"`
	CreatedBy         int           `json:"created_by"`
	CreatedByUsername string        `json:"created_by_username"`
	CreatedAt         string        `json:"created_at"`
}
