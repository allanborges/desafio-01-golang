package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type CotacaoResultado struct {
	Bid float64 `json:"bid"`
}

type CotacaoGravar struct {
	Dolar float64
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Println("falha ao criar o request")
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("falha ao receber a resposta")
	}
	defer res.Body.Close()
	resRead, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("falha ao receber o resultado da cotacao:", err)
	}
	var cotacao CotacaoResultado
	json.Unmarshal(resRead, &cotacao)
	arquivo, err := os.OpenFile("cotacao.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	defer arquivo.Close()
	if err != nil {
		panic(err)
	}
	encoder := json.NewEncoder(arquivo)
	encoder.Encode(CotacaoGravar{Dolar: cotacao.Bid})
}
