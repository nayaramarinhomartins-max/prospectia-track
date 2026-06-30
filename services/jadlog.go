package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type JadlogTransportadora struct {
	Token string
}

func (j *JadlogTransportadora) NomeDaTransportadora() string { return "Jadlog" }

func (j *JadlogTransportadora) ConsultarRastreio(codigo string) (*ResultadoRastreio, error) {
	url := fmt.Sprintf("https://www.jadlog.com.br/embarcador/api/tracking/consultar?cte=%s", codigo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+j.Token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao consultar Jadlog: %w", err)
	}
	defer resp.Body.Close()

	var body struct {
		Tracking struct {
			Eventos []struct {
				Status    string `json:"status"`
				Descricao string `json:"descricao"`
				Cidade    string `json:"cidade"`
				UF        string `json:"uf"`
			} `json:"eventos"`
		} `json:"tracking"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("erro ao ler resposta da Jadlog: %w", err)
	}

	eventos := body.Tracking.Eventos
	if len(eventos) == 0 {
		return nil, fmt.Errorf("nenhum evento encontrado para o código %s", codigo)
	}

	ev := eventos[0]
	return &ResultadoRastreio{
		Codigo:      codigo,
		Status:      ev.Status,
		Descricao:   ev.Descricao,
		Localizacao: fmt.Sprintf("%s - %s", ev.Cidade, ev.UF),
	}, nil
}
