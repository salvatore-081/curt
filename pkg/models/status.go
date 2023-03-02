package models

type Module struct {
	Path    string  `json:"path"`
	Version string  `json:"info"`
	Sum     string  `json:"sum,omitempty"`
	Replace *Module `json:"replace,omitempty"`
}
