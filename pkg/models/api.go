package models

type Body struct {
	Url string  `json:"url" validate:"required"`
	TTL *uint16 `json:"TTL,omitempty"`
}

type Header struct {
	XApiKey string `header:"X-API-Key"`
}

type Curt struct {
	Url       string  `json:"url,omitempty"`
	Curt      string  `json:"curt,omitempty"`
	Key       string  `json:"key"`
	TTL       *uint16 `json:"TTL,omitempty"`
	ExpiresAt *uint64 `json:"expiresAt,omitempty"`
}

type StatusInternalServerError struct {
	ErrorCode    int
	ErrorMessage string
}

type GenericError struct {
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}
