package model

import "time"

type ListingStatus string

const (
	ListingActive   ListingStatus = "active"
	ListingSold     ListingStatus = "sold"
	ListingArchived ListingStatus = "archived"
)

type Listing struct {
	ID          string        `json:"id"`
	SellerID    string        `json:"seller_id"`
	CategoryID  *int          `json:"category_id,omitempty"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Price       float64       `json:"price"`
	Currency    string        `json:"currency"`
	Status      ListingStatus `json:"status"`
	Location    *string       `json:"location,omitempty"`
	Images      []string      `json:"images"`
	Tags        []string      `json:"tags"`
	Views       int           `json:"views"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type CreateListingRequest struct {
	CategoryID  *int    `json:"category_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
	Location    *string `json:"location"`
	Tags        []string `json:"tags"`
}

type UpdateListingRequest struct {
	CategoryID  *int          `json:"category_id"`
	Title       *string       `json:"title"`
	Description *string       `json:"description"`
	Price       *float64      `json:"price"`
	Location    *string       `json:"location"`
	Tags        []string      `json:"tags"`
	Status      *ListingStatus `json:"status"`
}

type ListingsFilter struct {
	Query      string
	CategoryID *int
	MinPrice   *float64
	MaxPrice   *float64
	SellerID   string
	Cursor     string
	Limit      int
}

type ListingsPage struct {
	Items      []Listing `json:"items"`
	NextCursor string    `json:"next_cursor,omitempty"`
}
