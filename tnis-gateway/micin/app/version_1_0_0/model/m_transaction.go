package model

import (
	_ "github.com/go-sql-driver/mysql"
)

type TransactionDB struct {
	ID         int    `json:"id"`
	Date       string `json:"date"`
	CustomerID int    `json:"customer_id"`
	UserID     int    `json:"user_id"`
	Category   string `json:"category"`
	Total      int    `json:"total"`
}

type TransactionES struct {
	ID       int                   `json:"id"`
	Date     string                `json:"date"`
	Customer TransactionCustomerES `json:"customer"`
	User     TokenUserES           `json:"user"`
	Category string                `json:"category"`
	Total    int                   `json:"total"`
}

type TransactionShowES struct {
	ID       int             `json:"id"`
	Date     string          `json:"date"`
	User     TokenUserShowES `json:"user"`
	Category string          `json:"category"`
	Total    int             `json:"total"`
}

type TransactionCustomerES struct {
	ID            int    `json:"id"`
	AccountNumber string `json:"account_number"`
	IdentityCard  string `json:"identity_card"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	PhoneNumber   string `json:"phone_number"`
}

type CustomerES struct {
	ID            int    `json:"id"`
	AccountNumber string `json:"account_number"`
	IdentityCard  string `json:"identity_card"`
	Name          string `json:"name"`
	BirthDate     string `json:"birth_date"`
	Gender        string `json:"gender"`
	Address       string `json:"address"`
	Email         string `json:"email"`
	PhoneNumber   string `json:"phone_number"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
	CreatedBy     int    `json:"created_by"`
	UpdatedAt     string `json:"updated_at"`
	UpdatedBy     int    `json:"updated_by"`
	Total         int    `json:"total"`
}

type CustomerShowES struct {
	ID            int    `json:"id"`
	AccountNumber string `json:"account_number"`
	Name          string `json:"name"`
	Total         int    `json:"total"`
}

var TransactionCategoryEnum = []string{"in", "out"}
