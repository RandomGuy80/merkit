package model

import "time"

type OrderStatus string

const (
	OrderPending   OrderStatus = "pending"
	OrderPaid      OrderStatus = "paid"
	OrderShipped   OrderStatus = "shipped"
	OrderDelivered OrderStatus = "delivered"
	OrderCancelled OrderStatus = "cancelled"
	OrderRefunded  OrderStatus = "refunded"
)

type Order struct {
	ID              string      `json:"id"`
	ListingID       string      `json:"listing_id"`
	BuyerID         string      `json:"buyer_id"`
	SellerID        string      `json:"seller_id"`
	Amount          float64     `json:"amount"`
	PlatformFee     float64     `json:"platform_fee"`
	Status          OrderStatus `json:"status"`
	StripePaymentID *string     `json:"stripe_payment_id,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`

	Listing *Listing `json:"listing,omitempty"`
	Buyer   *User    `json:"buyer,omitempty"`
	Seller  *User    `json:"seller,omitempty"`
}

type CreateOrderRequest struct {
	ListingID string `json:"listing_id"`
}

type UpdateOrderStatusRequest struct {
	Status OrderStatus `json:"status"`
}
