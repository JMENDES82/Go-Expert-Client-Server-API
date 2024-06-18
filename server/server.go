package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/valyala/fastjson"
)

type Cotacao struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

func Server() {
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", CotacaoDolar)
	http.ListenAndServe(":8080", mux)
}

func inicializaBanco() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./server/cotacoes.db")
	if err != nil {
		log.Printf("Falha ao abrir banco de dados: %v", err)
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS cotacoes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		valor TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil {
		log.Printf("Falha ao criar a tabela: %v", err)
		return nil, err
	}
	return db, nil
}

func CotacaoDolar(w http.ResponseWriter, r *http.Request) {
	db, err := inicializaBanco()
	if err != nil {
		log.Fatalf("Falha ao inicializar o banco de dados: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
	defer cancel()

	cotacao, err := buscaCotacao(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Falha ao buscar a cotacao: %v", err), http.StatusInternalServerError)
		log.Println("Falha ao buscar a cotacao: %v", err)
		return
	}

	ctxDB, cancelDB := context.WithTimeout(r.Context(), 10*time.Millisecond)
	defer cancelDB()

	err = salvaCotacao(ctxDB, db, cotacao.USDBRL.Bid)
	if err != nil {
		log.Printf("Falha ao salvar cotacao: %v", err)
	}

	response := struct {
		Bid string `json:"bid"`
	}{
		Bid: cotacao.USDBRL.Bid,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func buscaCotacao(ctx context.Context) (*Cotacao, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var p fastjson.Parser
	value, err := p.ParseBytes(body)
	if err != nil {
		return nil, err
	}

	var result Cotacao
	if err := json.Unmarshal([]byte(value.String()), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func salvaCotacao(ctx context.Context, db *sql.DB, valor string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, err := db.ExecContext(ctx, `INSERT INTO cotacoes (valor) VALUES (?)`, valor)
		return err
	}
}
