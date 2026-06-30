package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

func EnviarWhatsAppStatusAtualizado(telefone, destinatarioNome, codigo, status, localizacao string) error {
	evolutionURL := os.Getenv("EVOLUTION_URL")
	evolutionKey := os.Getenv("EVOLUTION_KEY")
	evolutionInstance := os.Getenv("EVOLUTION_INSTANCE")

	if evolutionURL == "" || evolutionKey == "" || evolutionInstance == "" {
		return fmt.Errorf("Evolution API não configurada no .env")
	}

	// Normaliza o telefone — remove tudo que não é número
	telefone = normalizarTelefone(telefone)

	mensagem := fmt.Sprintf(
		"📦 *Atualização do seu pedido*\n\nOlá, %s!\n\n*Código:* %s\n*Status:* %s\n*Localização:* %s\n\n_Rastreamento automático_",
		destinatarioNome, codigo, status, localizacao,
	)

	payload := map[string]any{
		"number":  telefone,
		"text":    mensagem,
		"options": map[string]any{"delay": 1000},
	}

	corpo, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/message/sendText/%s", evolutionURL, evolutionInstance)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(corpo))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", evolutionKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao chamar Evolution API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Evolution API retornou status %d", resp.StatusCode)
	}

	return nil
}

// normalizarTelefone garante formato 5511999999999
func normalizarTelefone(tel string) string {
	var sb strings.Builder
	for _, c := range tel {
		if c >= '0' && c <= '9' {
			sb.WriteRune(c)
		}
	}
	numero := sb.String()
	if !strings.HasPrefix(numero, "55") {
		numero = "55" + numero
	}
	return numero
}
