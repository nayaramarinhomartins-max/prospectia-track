package models

import "time"

type Rastreio struct {
	ID            int       `json:"id"`
	Codigo        string    `json:"codigo"`
	Status        string    `json:"status"`
	Descricao     string    `json:"descricao"`
	Localizacao   string    `json:"localizacao"`
	DataEvento    time.Time `json:"data_evento"`
	CriadoEm     time.Time `json:"criado_em"`
}

type WebhookRegistro struct {
	ID       int    `json:"id"`
	Codigo   string `json:"codigo"`
	URL      string `json:"url"`
}
