package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/olivere/elastic"

	handler100 "tnis/tnis-transaction/micin/app/version_1_0_0/handler"
	"tnis/tnis-transaction/micin/config"
)

func (a *App) Initialize(config *config.Config) {
	// Connect MySQL
	var err error
	a.DB, err = sql.Open("mysql", config.DB.User+":"+config.DB.Password+"@"+config.DB.Host+"/"+config.DB.Database+"?autocommit=true&parseTime=false")
	a.DB.SetMaxIdleConns(0)
	a.DB.SetMaxOpenConns(1000)
	err = a.DB.Ping()
	if err != nil {
		fmt.Println("error app Initialize MySQL Connection : ", err.Error())
		return
	}
	fmt.Println("database connected")

	// Connect Elastic
	a.ES, err = elastic.NewClient(elastic.SetURL(config.ES.Host), elastic.SetRetrier(NewMyRetrier()))
	if err != nil {
		fmt.Println("error app Initialize Elasticsearch Connection : ", err.Error())
		return
	}
	fmt.Println("elasticsearch connected")

	// Check Indicate Elastic Transaction
	ctx := context.Background()
	exists, err := a.ES.IndexExists(config.ES.IndiceTransaction).Do(ctx)
	if err != nil {
		fmt.Println("error app Initialize Elasticsearch Indice Transaction : ", err)
		return
	}
	if !exists {
		fmt.Println("error app Initialize Elasticsearch Indice Transaction: ", err)
		return
	}

	// Check Indicate Elastic Customer
	exists, err = a.ES.IndexExists(config.ES.IndiceCustomer).Do(ctx)
	if err != nil {
		fmt.Println("error app Initialize Elasticsearch Indice Customer : ", err)
		return
	}
	if !exists {
		fmt.Println("error app Initialize Elasticsearch Indice Customer: ", err)
		return
	}

	// Check Indicate Elastic Token
	exists, err = a.ES.IndexExists(config.ES.IndiceToken).Do(ctx)
	if err != nil {
		fmt.Println("error app Initialize Elasticsearch Indice Token : ", err)
		return
	}
	if !exists {
		fmt.Println("error app Initialize Elasticsearch Indice Token : ", err)
		return
	}

	a.Router = mux.NewRouter()
	a.setRouters()
	a.Router.StrictSlash(true)
	a.Router.NotFoundHandler = http.HandlerFunc(RespondApiNotFound)
}

func (a *App) Get(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.Router.HandleFunc(path, f).Methods("GET")
}

func (a *App) Post(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.Router.HandleFunc(path, f).Methods("POST")
}

func (a *App) Put(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.Router.HandleFunc(path, f).Methods("PUT")
}

func (a *App) Delete(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.Router.HandleFunc(path, f).Methods("DELETE")
}

func (a *App) setRouters() {
	a.Get("/", Welcome)

	// Transaction
	a.Post("/v100/transaction/save", a.TransactionSaveWithdraw100)
	a.Post("/v100/transaction/withdraw", a.TransactionSaveWithdraw100)
	a.Post("/v100/transaction/history/{account_number}", a.TransactionHistory100)
	a.Get("/v100/transaction/balance/{account_number}", a.TransactionBalance100)
}

func (a *App) TransactionSaveWithdraw100(w http.ResponseWriter, r *http.Request) {
	handler100.TransactionSaveWithdraw(w, r, a.DB, a.ES)
}

func (a *App) TransactionHistory100(w http.ResponseWriter, r *http.Request) {
	handler100.TransactionHistory(w, r, a.DB, a.ES)
}

func (a *App) TransactionBalance100(w http.ResponseWriter, r *http.Request) {
	handler100.TransactionBalance(w, r, a.DB, a.ES)
}

func (a *App) Run(host string) {
	loggedRouter := handlers.LoggingHandler(os.Stdout, a.Router)
	http.ListenAndServe(host, loggedRouter)
}
