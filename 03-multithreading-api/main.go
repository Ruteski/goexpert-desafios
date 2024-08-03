package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type response struct {
	site string
	res  string
}

func main() {
	http.HandleFunc("/", buscaCEP)
	fmt.Println("Server online na porta 8000 🚀")
	http.ListenAndServe(":8000", nil)
}

func buscaCEP(w http.ResponseWriter, r *http.Request) {
	cepParam := r.URL.Query().Get("cep")
	if cepParam == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	c1 := make(chan response)
	c2 := make(chan response)

	go func() {
		//time.Sleep(1 * time.Second)
		req, err := http.Get("http://viacep.com.br/ws/" + cepParam + "/json/")
		if err != nil {
			log.Fatal(err)
		}
		defer req.Body.Close()

		res, err := io.ReadAll(req.Body)
		if err != nil {
			panic(err)
		}
		c2 <- response{"https://viacep.com.br", string(res)}
	}()

	go func() {
		//time.Sleep(2 * time.Second)
		req, err := http.Get("https://brasilapi.com.br/api/cep/v1/" + cepParam)
		if err != nil {
			log.Fatal(err)
		}
		defer req.Body.Close()

		res, err := io.ReadAll(req.Body)
		if err != nil {
			panic(err)
		}
		c1 <- response{"https://brasilapi.com.br", string(res)}
	}()

	select {
	case msg := <-c1:
		fmt.Printf("Site de consulta: %s.\nDados do endereço: %s\n\n", msg.site, msg.res)
	case msg := <-c2:
		fmt.Printf("Site de consulta: %s.\nDados do endereço: %s\n\n", msg.site, msg.res)
	case <-time.After(time.Second * 1):
		println("timeout")
	}
}
