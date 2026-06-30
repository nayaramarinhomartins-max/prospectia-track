package services

import (
	"fmt"
	"log"
)

// ConfigNotificacao carrega do banco para cada cliente
type ConfigNotificacao struct {
	WhatsAppProvider string // "evolution", "360dialog", "none"
	WhatsAppKey      string // chave do provider escolhido
	WhatsAppInstance string // instância Evolution (só se provider = evolution)
}

// EnviarNotificacao decide qual canal usar e dispara
func EnviarNotificacao(cfg ConfigNotificacao, telefone, email, nome, codigo, status, localizacao string) {
	if email != "" {
		if err := EnviarEmailStatusAtualizado(email, nome, codigo, status, localizacao); err != nil {
			log.Printf("[Email] Falha para %s: %v", email, err)
		} else {
			log.Printf("[Email] Enviado para %s", email)
		}
	}

	if telefone == "" {
		return
	}

	switch cfg.WhatsAppProvider {
	case "evolution":
		if err := EnviarWhatsAppStatusAtualizado(telefone, nome, codigo, status, localizacao); err != nil {
			log.Printf("[Evolution] Falha para %s: %v", telefone, err)
		} else {
			log.Printf("[Evolution] Enviado para %s", telefone)
		}

	case "360dialog":
		if cfg.WhatsAppKey == "" {
			log.Printf("[360Dialog] API Key não configurada para este cliente")
			return
		}
		if err := EnviarWhatsApp360Dialog(cfg.WhatsAppKey, telefone, nome, codigo, status, localizacao); err != nil {
			log.Printf("[360Dialog] Falha para %s: %v", telefone, err)
		} else {
			log.Printf("[360Dialog] Enviado para %s", telefone)
		}

	case "none", "":
		log.Printf("[WhatsApp] Notificação desativada para este cliente")

	default:
		log.Printf("[WhatsApp] Provider desconhecido: %s", cfg.WhatsAppProvider)
		fmt.Println()
	}
}
