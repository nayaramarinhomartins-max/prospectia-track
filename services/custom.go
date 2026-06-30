package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// CustomTransportadora funciona com qualquer API de rastreio.
// O cliente informa a URL, o tipo de autenticação e onde fica o status na resposta JSON.
type CustomTransportadora struct {
	Nome        string
	EndpointURL string // ex: "https://api.xyz.com/track/{codigo}"
	AuthTipo    string // "none", "bearer", "api_key"
	AuthValor   string // o token ou chave
	JsonPath    string // ex: "data.status" — onde está o status na resposta
}

func (c *CustomTransportadora) NomeDaTransportadora() string { return c.Nome }

func (c *CustomTransportadora) ConsultarRastreio(codigo string) (*ResultadoRastreio, error) {
	// Substitui {codigo} na URL pelo código real
	url := strings.ReplaceAll(c.EndpointURL, "{codigo}", codigo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao montar requisição: %w", err)
	}

	// Aplica autenticação conforme configurado
	switch c.AuthTipo {
	case "bearer":
		req.Header.Set("Authorization", "Bearer "+c.AuthValor)
	case "api_key":
		req.Header.Set("X-Api-Key", c.AuthValor)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao consultar %s: %w", c.Nome, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %w", err)
	}

	// Transforma o JSON em mapa genérico
	var dados map[string]any
	if err := json.Unmarshal(body, &dados); err != nil {
		return nil, fmt.Errorf("resposta não é um JSON válido: %w", err)
	}

	// Navega pelo JSON usando o caminho configurado (ex: "data.status")
	status := extrairCampo(dados, c.JsonPath)
	if status == "" {
		status = "Status não identificado"
	}

	return &ResultadoRastreio{
		Codigo:      codigo,
		Status:      status,
		Descricao:   fmt.Sprintf("Consultado via %s", c.Nome),
		Localizacao: extrairCampo(dados, "localizacao"),
	}, nil
}

// extrairCampo navega num JSON usando notação de ponto: "data.tracking.status"
func extrairCampo(dados map[string]any, caminho string) string {
	partes := strings.Split(caminho, ".")
	var atual any = dados

	for _, parte := range partes {
		mapa, ok := atual.(map[string]any)
		if !ok {
			return ""
		}
		atual = mapa[parte]
	}

	if atual == nil {
		return ""
	}
	return fmt.Sprintf("%v", atual)
}
