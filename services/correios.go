package services

import (
	"fmt"
	"time"
)

type CorreiosTransportadora struct{}

func (c *CorreiosTransportadora) NomeDaTransportadora() string { return "Correios" }

func (c *CorreiosTransportadora) ConsultarRastreio(codigo string) (*ResultadoRastreio, error) {
	if len(codigo) < 13 {
		return nil, fmt.Errorf("código de rastreio inválido: %s", codigo)
	}

	// Simulação realista enquanto não há credenciais de API dos Correios.
	// Em produção, substituir pela integração real (Linketrack pago, ou API oficial Correios com CNPJ).
	resultado := simularRastreioCorreios(codigo)
	return resultado, nil
}

func simularRastreioCorreios(codigo string) *ResultadoRastreio {
	hora := time.Now()

	// Varia o status baseado no último dígito do código para simular casos diferentes
	ultimo := string(codigo[len(codigo)-1])
	switch ultimo {
	case "1", "2", "3":
		return &ResultadoRastreio{
			Codigo:      codigo,
			Status:      "Em trânsito",
			Descricao:   fmt.Sprintf("Objeto encaminhado - %s", hora.Format("02/01/2006 15:04")),
			Localizacao: "CDD São Paulo / SP",
		}
	case "4", "5", "6":
		return &ResultadoRastreio{
			Codigo:      codigo,
			Status:      "Saiu para entrega",
			Descricao:   fmt.Sprintf("Objeto saiu para entrega ao destinatário - %s", hora.Format("02/01/2006 15:04")),
			Localizacao: "CDD Campinas / SP",
		}
	case "7", "8":
		return &ResultadoRastreio{
			Codigo:      codigo,
			Status:      "Entregue",
			Descricao:   fmt.Sprintf("Objeto entregue ao destinatário - %s", hora.Format("02/01/2006 15:04")),
			Localizacao: "Campinas / SP",
		}
	default:
		return &ResultadoRastreio{
			Codigo:      codigo,
			Status:      "Postado",
			Descricao:   fmt.Sprintf("Objeto postado - %s", hora.Format("02/01/2006 15:04")),
			Localizacao: "Agência dos Correios - São Paulo / SP",
		}
	}
}
