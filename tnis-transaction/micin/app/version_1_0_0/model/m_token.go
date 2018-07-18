package model

import (
	_ "github.com/go-sql-driver/mysql"
)

type TokenDB struct {
	ID         string `json:"id"`
	UserID     int    `json:"user_id"`
	ClientID   string `json:"client_id"`
	CreatedAt  string `json:"created_at"`
	ValidUntil string `json:"valid_until"`
}

type TokenES struct {
	ID         string        `json:"id"`
	User       TokenUserES   `json:"user"`
	Client     TokenClientES `json:"client"`
	CreatedAt  string        `json:"created_at"`
	ValidUntil string        `json:"valid_until"`
}

type TokenUserES struct {
	ID    int    `json:"id"`
	Code  string `json:"code"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type TokenUserShowES struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type TokenClientES struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}
