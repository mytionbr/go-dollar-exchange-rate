package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	_ "modernc.org/sqlite"
)

const (
	apiURL     = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	serverPort = ":8080"
	dbUrl      = "file:cotacoes.db?cache=shared&mode=rwc&_pragma=busy_timeout(5000)"
	apiTimeout = 200 * time.Millisecond
	dbTimeout  = 10 * time.Millisecond
)

type quote struct {
	Bid string `json:"bid"`
}

type apiResponse struct {
	USDBRL quote `json:"USDBRL"`
}

type server struct {
	db     *sql.DB
	client *http.Client
}

func main() {
	db, err := sql.Open("sqlite", dbUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.SetMaxOpenConns(1)
	db.SetConnMaxIdleTime(30 * time.Second)

	if err := ensureSchema(db); err != nil {
		log.Fatalf("erro ao criar schema: %v", err)
	}

	s := &server{
		db:     db,
		client: &http.Client{},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", s.handleCotacao)

	srv := &http.Server{
		Addr:              serverPort,
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
	}

	log.Printf("Servidor iniciado na porta %s", serverPort)

	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}

func ensureSchema(db *sql.DB) error {
	const ddl = `CREATE TABLE IF NOT EXISTS cotacoes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		bid TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := db.Exec(ddl)
	return err
}

func (s *server) handleCotacao(w http.ResponseWriter, r *http.Request) {
	q, err := s.fetchUSD(r.Context())
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			log.Printf("[fetchUSD] contexto excedido/cancelado: %v", err)
			http.Error(w, "Tempo limite excedido", http.StatusGatewayTimeout)
			return
		}

		log.Printf("[fetchUSD] erro ao buscar cotação: %v", err)
		http.Error(w, "erro ao consultar API externa", http.StatusInternalServerError)
		return
	}

	if err := s.saveQuote(r.Context(), q); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			log.Printf("[saveQuote] contexto excedido/cancelado ao salvar cotação: %v", err)
		} else {
			log.Printf("[saveQuote] erro ao salvar cotação: %v", err)
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]string{"bid": q.Bid})
}

func (s *server) fetchUSD(parent context.Context) (quote, error) {
	ctx, cancel := context.WithTimeout(parent, apiTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return quote{}, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return quote{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return quote{}, errors.New("API externa retornou status diferente de 200")
	}

	var ar apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		return quote{}, err
	}

	return ar.USDBRL, nil
}

func (s *server) saveQuote(ctx context.Context, q quote) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	_, err := s.db.ExecContext(ctx, "INSERT INTO cotacoes (bid) VALUES (?)", q.Bid)
	return err
}
