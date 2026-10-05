package model

import "time"

type Review struct {
	ID         string    `json:"id"`
	OrderID    string    `json:"order_id"`
	ReviewerID string    `json:"reviewer_id"`
	SellerID   string    `json:"seller_id"`
	Rating     int       `json:"rating"`
	Body       string    `json:"body,omitempty"`
	CreatedAt  time.Time `json:"created_at"`

	Reviewer *User `json:"reviewer,omitempty"`
}

type CreateReviewRequest struct {
	OrderID string `json:"order_id"`
	Rating  int    `json:"rating"`
	Body    string `json:"body"`
}
