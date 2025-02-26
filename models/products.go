package models

type ProductRequest struct {
	Name      string  `json:"name"`
	UnitPrice float64 `json:"unit_price"`
	Quantity  int     `json:"quantity"`
	Reason    string  `json:"reason"`
}

type Product struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	UnitPrice float64 `json:"unit_price"`
	Quantity  int     `json:"quantity"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
	DeletedAt string  `json:"deleted_at"`
}

type StockRequest struct {
	Quantity int    `json:"quantity"`
	Reason   string `json:"reason"`
}

type ProductResponse struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	UnitPrice float64 `json:"unit_price"`
	Quantity  int     `json:"quantity"`
}
