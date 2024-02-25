package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
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

const URL_COTATION = "https://economia.awesomeapi.com.br/json/last/USD-BRL"

func main() {
	db, err := sql.Open("sqlite3", "file.db")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS ExchangeRate (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code TEXT NOT NULL,
		codein TEXT NOT NULL,
		name TEXT NOT NULL,
		high REAL NOT NULL,
		low REAL NOT NULL,
		varBid REAL NOT NULL,
		pctChange REAL NOT NULL,
		bid REAL NOT NULL,
		ask REAL NOT NULL,
		timestamp INTEGER NOT NULL,
		create_date TEXT NOT NULL
	);
	`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		panic(err.Error())
	}
	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		cotacao := manipulaResposta(w)
		insereNovaCotacao(db, cotacao)
	})
	http.ListenAndServe(":8080", nil)
}

func manipulaResposta(w http.ResponseWriter) *Cotacao {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*200)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", URL_COTATION, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}
	resp, err := http.DefaultClient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Erro ao ler o corpo da resposta:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}
	var cotacao Cotacao
	err = json.Unmarshal(body, &cotacao)
	if err != nil {
		log.Println("Erro ao fazer parse da resposta JSON:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}
	log.Println("status", resp.StatusCode)
	w.WriteHeader(http.StatusCreated)
	w.Header().Add("Content-Type:", "application/json")

	w.Write([]byte("{\"bid\":" + cotacao.USDBRL.Bid + "}"))
	return &cotacao
}

func insereNovaCotacao(db *sql.DB, cotacao *Cotacao) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*10)
	defer cancel()
	sqlStatement := `INSERT INTO ExchangeRate
	(code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, create_date)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	stmt, err := db.Prepare(sqlStatement)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	cotacaoUSDBRL := cotacao.USDBRL
	_, err = stmt.ExecContext(ctx, cotacaoUSDBRL.Code, cotacaoUSDBRL.Codein, cotacaoUSDBRL.Name,
		cotacaoUSDBRL.High, cotacaoUSDBRL.Low, cotacaoUSDBRL.VarBid,
		cotacaoUSDBRL.PctChange, cotacaoUSDBRL.Bid, cotacaoUSDBRL.Ask, cotacaoUSDBRL.Timestamp, cotacaoUSDBRL.CreateDate)
	if err != nil {
		panic(err)
	}
}
