package models

type Body struct {
	Url string  `json:"url"`
	TTL *uint16 `json:"TTL,omitempty"`
}
