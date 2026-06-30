package services

import (
	"fmt"
	"os"

	"github.com/resend/resend-go/v3"
)

func EnviarEmailStatusAtualizado(destinatarioEmail, destinatarioNome, codigo, status, localizacao string) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY não configurada")
	}

	client := resend.NewClient(apiKey)

	corpo := fmt.Sprintf(`
	<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
		<h2 style="color: #1a1a2e;">📦 Atualização do seu pedido</h2>
		<p>Olá, <strong>%s</strong>!</p>
		<p>Seu pedido foi atualizado:</p>
		<div style="background: #f4f4f4; padding: 16px; border-radius: 8px; margin: 16px 0;">
			<p><strong>Código:</strong> %s</p>
			<p><strong>Status:</strong> %s</p>
			<p><strong>Localização:</strong> %s</p>
		</div>
		<p style="color: #888; font-size: 12px;">Você recebe este e-mail pois cadastrou rastreamento automático.</p>
	</div>
	`, destinatarioNome, codigo, status, localizacao)

	params := &resend.SendEmailRequest{
		From:    "Rastreio <onboarding@resend.dev>",
		To:      []string{destinatarioEmail},
		Subject: fmt.Sprintf("Pedido %s — %s", codigo, status),
		Html:    corpo,
	}

	_, err := client.Emails.Send(params)
	return err
}
