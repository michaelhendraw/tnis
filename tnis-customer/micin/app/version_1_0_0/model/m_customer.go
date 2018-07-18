package model

import (
	_ "github.com/go-sql-driver/mysql"
)

type CustomerDB struct {
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

var CustomerStatusEnum = []string{"active", "inactive", "pending"}
var CustomerGenderEnum = []string{"man", "woman"}
