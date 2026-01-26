package models

import "time"

type Order struct {
	ID		 		string	   		`json:"id"`
	CustomerID 		string	  		`json:"customer_id"`
	RestaurantID 	string	   		`json:"restaurant_id"`
	Items      		[]string 		`json:"items"`
	Status			OrderStatus 	`json:"status"`
	CreatedAt  		time.Time  		`json:"created_at"`
	UpdatedAt  		time.Time  		`json:"updated_at"`
}


type OrderStatus string

const (
	StatusPending    OrderStatus = "pending"
	StatusProcessing OrderStatus = "processing"
	StatusCompleted  OrderStatus = "completed"
	StatusFailed     OrderStatus = "failed"
)


// OrderRequest represents the incoming order creation request
type OrderRequest struct {
	CustomerID   	string   	`json:"customer_id"`
	RestaurantID 	string   	`json:"restaurant_id"`
	Items        	[]string 	`json:"items"`
}