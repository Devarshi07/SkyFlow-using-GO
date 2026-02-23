package cities

import "time"

type City struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Country   string    `json:"country"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateCityRequest struct {
	Name    string `json:"name"`
	Country string `json:"country"`
	Code    string `json:"code"`
}

type UpdateCityRequest struct {
	Name    *string `json:"name,omitempty"`
	Country *string `json:"country,omitempty"`
	Code    *string `json:"code,omitempty"`
}
