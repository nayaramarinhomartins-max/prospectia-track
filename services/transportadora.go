package services

import "fmt"

// ResultadoRastreio é o formato unificado que todas as transportadoras devolvem.
// Não importa se é Correios, Jadlog ou uma desconhecida — a resposta sempre tem este formato.
type ResultadoRastreio struct {
	Codigo      string  `json:"codigo"`
	Status      string  `json:"status"`
	Descricao   string  `json:"descricao"`
	Localizacao string  `json:"localizacao"`
}

// Transportadora é o contrato. Toda transportadora — conhecida ou não — precisa implementar estas duas funções.
type Transportadora interface {
	ConsultarRastreio(codigo string) (*ResultadoRastreio, error)
	NomeDaTransportadora() string
}

// ConfigTransportadora são os dados que vêm do banco para montar a transportadora certa.
type ConfigTransportadora struct {
	ID          int
	Nome        string
	Tipo        string // "correios", "jadlog", "totalexpress", "custom"
	EndpointURL string
	AuthTipo    string // "none", "bearer", "api_key"
	AuthValor   string
	JsonPath    string // caminho do status na resposta JSON, ex: "data.status"
}

// CriarTransportadora é a fábrica — recebe a config do banco e devolve a implementação certa.
// Se for uma transportadora desconhecida (tipo "custom"), usa o adaptador genérico.
func CriarTransportadora(config ConfigTransportadora) (Transportadora, error) {
	switch config.Tipo {
	case "correios":
		return &CorreiosTransportadora{}, nil
	case "jadlog":
		return &JadlogTransportadora{Token: config.AuthValor}, nil
	case "custom":
		return &CustomTransportadora{
			Nome:        config.Nome,
			EndpointURL: config.EndpointURL,
			AuthTipo:    config.AuthTipo,
			AuthValor:   config.AuthValor,
			JsonPath:    config.JsonPath,
		}, nil
	default:
		return nil, fmt.Errorf("tipo de transportadora desconhecido: %s", config.Tipo)
	}
}
