package model

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Icon string `json:"icon,omitempty"`
}

type CreateCategoryRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
	Icon string `json:"icon"`
}
