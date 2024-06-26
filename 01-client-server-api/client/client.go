package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	serverURL  = "http://localhost:8080/cotacao"
	timeoutReq = 300 * time.Millisecond
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutReq)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", serverURL, nil)
	if err != nil {
		log.Fatalf("Erro ao criar requisição: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			log.Fatal("Tempo excedido ao tentar fazer a requisição!")
		}

		log.Fatalf("Erro ao fazer requisição: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Erro na resposta do servidor: %v.", resp.Status)
	}

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler a resposta: %v\n", err)
	}

	err = gravarArquivo(string(res))
	if err != nil {
		log.Fatalf("Erro ao gravarar arquivo: %v", err)
	}

	fmt.Println("Cotação salva em cotacao.txt")
}

func gravarArquivo(bid string) error {
	f, err := os.Create("cotacao.txt")
	if err != nil {
		return err
	}

	tamanho, err := f.Write([]byte(fmt.Sprintf("Dólar: %s", bid)))
	if err != nil {
		return err
	}

	fmt.Printf("Arquivo criado com sucesso! Tamanho: %d bytes\n", tamanho)
	f.Close()

	return nil
}
