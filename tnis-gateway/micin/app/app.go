package app

import (
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"tnis/tnis-gateway/micin/config"
)

func (a *App) Initialize(config *config.Config) {
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

	// --- AUTH ---
	// Client
	a.Post("/v100/auth/create", a.AuthCreate100)
	// Token Auth
	a.Post("/v100/auth/get_token", a.AuthGetToken100)
	a.Post("/v100/auth/check_token", a.AuthCheckToken100)

	// --- CUSTOMER ---
	// Customer
	a.Post("/v100/customer/create", a.CustomerCreate100)
	a.Put("/v100/customer/update/{id}", a.CustomerUpdate100)

	// --- TRANSACTION ---
	// Transaction
	a.Post("/v100/transaction/save", a.TransactionSaveWithdraw100)
	a.Post("/v100/transaction/withdraw", a.TransactionSaveWithdraw100)
	a.Post("/v100/transaction/history/{account_number}", a.TransactionHistory100)
	a.Get("/v100/transaction/balance/{account_number}", a.TransactionBalance100)

	// --- NOTIF ---
	// Notif
	a.Get("/v100/notif/send_email/{id}", a.NotifSendEmail100)
}

/*
	1. Request
		a. URL + Headers + Body
	2. Gateway
		a. Cek service yang dipanggil, butuh token gak
			- Jika butuh token, cek service valid gak (auth/check_token)
		b. Call API service
	3. Respond
*/

// --- AUTH ---
func (a *App) AuthCreate100(w http.ResponseWriter, r *http.Request) {
	service := "auth/create"
	AuthAndCallAPI(w, r, service, "post", "v100")
}

func (a *App) AuthGetToken100(w http.ResponseWriter, r *http.Request) {
	service := "auth/get_token"
	CallAPI(w, r, service, "post", "v100")
}

func (a *App) AuthCheckToken100(w http.ResponseWriter, r *http.Request) {
	service := "auth/check_token"
	CallAPI(w, r, service, "post", "v100")
}

// --- CUSTOMER ---
func (a *App) CustomerCreate100(w http.ResponseWriter, r *http.Request) {
	service := "customer/create"
	AuthAndCallAPI(w, r, service, "post", "v100")
}

func (a *App) CustomerUpdate100(w http.ResponseWriter, r *http.Request) {
	service := "customer/update"
	AuthAndCallAPI(w, r, service, "put", "v100")
}

// --- TRANSACTION ---
func (a *App) TransactionSaveWithdraw100(w http.ResponseWriter, r *http.Request) {
	service := "transaction/save"
	AuthAndCallAPI(w, r, service, "post", "v100")
}

func (a *App) TransactionHistory100(w http.ResponseWriter, r *http.Request) {
	service := "transaction/history"
	AuthAndCallAPI(w, r, service, "post", "v100")
}

func (a *App) TransactionBalance100(w http.ResponseWriter, r *http.Request) {
	service := "transaction/balance"
	AuthAndCallAPI(w, r, service, "get", "v100")
}

// --- NOTIF ---
func (a *App) NotifSendEmail100(w http.ResponseWriter, r *http.Request) {
	service := "notif/send_email"
	AuthAndCallAPI(w, r, service, "get", "v100")
}

func (a *App) Run(host string) {
	loggedRouter := handlers.LoggingHandler(os.Stdout, a.Router)
	http.ListenAndServe(host, loggedRouter)
}
