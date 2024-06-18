package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/valyala/fastjson"
)

type Cotacao struct {
	Bid string `json:"bid"`
}

func Client() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	cotacao, err := buscaCotacao(ctx)
	if err != nil {
		log.Printf("Falha ao buscar cotacao: %v", err)
	}

	err = salvaEmArquivo(cotacao.Bid)
	if err != nil {
		log.Printf("Falha ao salvar em arquivo: %v", err)
	}

}

func buscaCotacao(ctx context.Context) (*Cotacao, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
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

func salvaEmArquivo(cotacao string) error {
	data := fmt.Sprintf("Dólar: %s", cotacao)
	return os.WriteFile("./client/cotacao.txt", []byte(data), 0644)
}
