package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func WhatsAppQR(w http.ResponseWriter, r *http.Request) {
	url := os.Getenv("EVOLUTION_URL")
	key := os.Getenv("EVOLUTION_KEY")
	instance := os.Getenv("EVOLUTION_INSTANCE")

	if url == "" || key == "" || instance == "" {
		http.Error(w, "Evolution API não configurada no .env", http.StatusServiceUnavailable)
		return
	}

	client := &http.Client{Timeout: 15 * time.Second}

	// Busca o QR Code da instância
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/instance/connect/%s", url, instance), nil)
	req.Header.Set("apikey", key)

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Erro ao conectar na Evolution API: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func WhatsAppStatus(w http.ResponseWriter, r *http.Request) {
	url := os.Getenv("EVOLUTION_URL")
	key := os.Getenv("EVOLUTION_KEY")
	instance := os.Getenv("EVOLUTION_INSTANCE")

	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/instance/fetchInstances", url), nil)
	req.Header.Set("apikey", key)

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Erro ao verificar status", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	var instancias []map[string]any
	json.NewDecoder(resp.Body).Decode(&instancias)

	// Filtra a instância configurada
	for _, inst := range instancias {
		if inst["name"] == instance {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(inst)
			return
		}
	}

	http.Error(w, "Instância não encontrada", http.StatusNotFound)
}
