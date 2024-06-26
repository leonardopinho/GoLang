package main

import (
	"context"
	"encoding/json"
	"github.com/leonardopinho/GoLang/1.a-Client-Server-API/server/internal/middleware"
	"github.com/leonardopinho/GoLang/1.a-Client-Server-API/server/internal/utils"
	"github.com/leonardopinho/GoLang/1.a-Client-Server-API/server/pkg/db"
	"github.com/leonardopinho/GoLang/1.a-Client-Server-API/server/pkg/db/models"
	"io"
	"log"
	"net/http"
	"time"
)

type BidResponse struct {
	Value string `json:"value"`
}

func main() {
	err := db.InitDb()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Starting server...")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /cotacao", getDollarPriceHandle)
	if err := http.ListenAndServe(":8080", middleware.RecoveryMiddleware(mux)); err != nil {
		log.Fatal(err)
	}

}

func getDollarPriceHandle(w http.ResponseWriter, _ *http.Request) {
	price, err := GetUSDBRL()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if price == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	// save in database
	_, err = db.SaveUSDBRL(price)
	if err != nil {
		http.Error(w, err.Error(), http.StatusRequestTimeout)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// result
	err = json.NewEncoder(w).Encode(BidResponse{Value: price.Bid})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func GetUSDBRL() (*models.USDBRL, error) {
	url := "https://economia.awesomeapi.com.br/json/last/USD-BRL"

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var current models.USDBRL
	err = utils.ParseJson(body, "USDBRL", &current)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &current, nil
}
