package models

import "time"

type Cliente struct {
	ID       int       `json:"id"`
	Nome     string    `json:"nome"`
	Email    string    `json:"email"`
	ApiKey   string    `json:"api_key"`
	Plano    string    `json:"plano"`
	Ativo    bool      `json:"ativo"`
	CriadoEm time.Time `json:"criado_em"`
}
