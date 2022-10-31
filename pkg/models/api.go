package models

type Body struct {
	Url string  `json:"url"`
	TTL *uint16 `json:"TTL,omitempty"`
}

type Header struct {
	ApiKey string `header:"api_key"`
}

type Curt struct {
	Url       string  `json:"url"`
	Curt      string  `json:"curt"`
	Key       string  `json:"key"`
	TTL       *uint16 `json:"TTL,omitempty"`
	ExpiresAt *uint64 `json:"expiresAt,omitempty"`
}
