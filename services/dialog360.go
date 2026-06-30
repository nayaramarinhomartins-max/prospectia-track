package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Dialog360 usa a WhatsApp Business API oficial — sem risco de ban.
// O cliente precisa ter uma conta aprovada na 360Dialog com número verificado.
func EnviarWhatsApp360Dialog(apiKey, telefone, destinatarioNome, codigo, status, localizacao string) error {
	telefone = normalizarTelefone(telefone)

	mensagem := fmt.Sprintf(
		"📦 *Atualização do seu pedido*\n\nOlá, %s!\n\n*Código:* %s\n*Status:* %s\n*Localização:* %s\n\n_ProspectIA Track_",
		destinatarioNome, codigo, status, localizacao,
	)

	payload := map[string]any{
		"recipient_type": "individual",
		"to":             telefone,
		"type":           "text",
		"text":           map[string]string{"body": mensagem},
	}

	corpo, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "https://waba.360dialog.io/v1/messages", bytes.NewBuffer(corpo))
	if err != nil {
		return err
	}
	req.Header.Set("D360-API-KEY", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao chamar 360Dialog: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("360Dialog retornou status %d", resp.StatusCode)
	}

	return nil
}

// normalizarTelefone360 garante formato internacional sem + (ex: 5511999999999)
func normalizarTelefone360(tel string) string {
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
