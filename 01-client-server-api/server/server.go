package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
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

func main() {
	http.HandleFunc("/cotacao", handlerCotacao)

	fmt.Println("Server online na porta 8080 🚀")
	http.ListenAndServe(":8080", nil)
}

func handlerCotacao(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "cotacoes.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	cotacao, err := cotacaoDolar()
	if err != nil {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			log.Println("Tempo excedido ao realizar a busca da cotação!")
		} else {
			log.Println(err)
		}

		http.Error(w, "Erro ao coletar cotação do dólar", http.StatusInternalServerError)
		return
	}

	if err := persistirCotacao(db, cotacao); err != nil {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			log.Println("Tempo excedido ao tentar salvar a cotação no banco de dados!")
		} else {
			log.Println(err)
		}

		http.Error(w, "Erro ao persistir cotação do dólar", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(cotacao.USDBRL.Bid)
}

func cotacaoDolar() (*Cotacao, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var cotacao Cotacao
	err = json.Unmarshal(body, &cotacao)

	if err != nil {
		return nil, err
	}

	return &cotacao, nil
}

func persistirCotacao(db *sql.DB, cotacao *Cotacao) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	stmt, err := db.PrepareContext(ctx, "INSERT INTO cotacoes(code,codein,name,high,low,varBid,pctChange,bid,ask,timestamp,create_date)VALUES (?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx,
		&cotacao.USDBRL.Code,
		&cotacao.USDBRL.Codein,
		&cotacao.USDBRL.Name,
		&cotacao.USDBRL.High,
		&cotacao.USDBRL.Low,
		&cotacao.USDBRL.VarBid,
		&cotacao.USDBRL.PctChange,
		&cotacao.USDBRL.Bid,
		&cotacao.USDBRL.Ask,
		&cotacao.USDBRL.Timestamp,
		&cotacao.USDBRL.CreateDate)

	if err != nil {
		return err
	}

	return nil
}
