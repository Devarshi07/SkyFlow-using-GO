package airports

import "time"

type Airport struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CityID    string    `json:"city_id"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateAirportRequest struct {
	Name   string `json:"name"`
	CityID string `json:"city_id"`
	Code   string `json:"code"`
}

type UpdateAirportRequest struct {
	Name   *string `json:"name,omitempty"`
	CityID *string `json:"city_id,omitempty"`
	Code   *string `json:"code,omitempty"`
}
