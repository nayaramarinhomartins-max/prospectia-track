package services

import (
	"fmt"
	"log"
	"time"
)

// RetryConfig define quantas tentativas e quanto esperar entre elas
type RetryConfig struct {
	MaxTentativas int
	EsperaInicial time.Duration
}

var PadraoRetry = RetryConfig{
	MaxTentativas: 3,
	EsperaInicial: 1 * time.Second,
}

// ComRetry executa uma função e repete automaticamente se falhar.
// Cada falha dobra o tempo de espera: 1s → 2s → 4s (exponential backoff)
func ComRetry(config RetryConfig, nome string, fn func() (*ResultadoRastreio, error)) (*ResultadoRastreio, error) {
	espera := config.EsperaInicial

	for tentativa := 1; tentativa <= config.MaxTentativas; tentativa++ {
		resultado, err := fn()
		if err == nil {
			if tentativa > 1 {
				log.Printf("[%s] Sucesso na tentativa %d", nome, tentativa)
			}
			return resultado, nil
		}

		if tentativa == config.MaxTentativas {
			return nil, fmt.Errorf("falhou após %d tentativas: %w", config.MaxTentativas, err)
		}

		log.Printf("[%s] Tentativa %d falhou: %v — aguardando %s antes de tentar novamente", nome, tentativa, err, espera)
		time.Sleep(espera)
		espera *= 2 // dobra o tempo a cada falha
	}

	return nil, fmt.Errorf("retry esgotado")
}
