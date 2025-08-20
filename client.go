package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	serverUrl     = "http://localhost:8080/cotacao"
	clientTimeout = 300 * time.Millisecond
	outFile       = "cotacao.txt"
)

type response struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, serverUrl, nil)

	if err != nil {
		log.Fatalf("erro ao criar request: %v", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			log.Fatalf("[client] contexto excedido/cancelado ao fazer request: %v", err)
		}
		log.Fatalf("[client] erro ao fazer request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("[client] servidor retornou status: %s", resp.Status)
	}

	var result response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatalf("[client] erro ao decodificar resposta: %v", err)
	}

	content := fmt.Sprintf("Cotação do dólar: %s", result.Bid)
	if err := os.WriteFile(outFile, []byte(content), 0644); err != nil {
		log.Fatalf("[client] erro ao escrever no arquivo %s: %v", outFile, err)
	}

	log.Printf("cotação salva em %s: %s", outFile, result.Bid)
}
