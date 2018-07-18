package model

import (
	_ "github.com/go-sql-driver/mysql"
)

type ClientDB struct {
	ID        string      `json:"id"`
	Key       string      `json:"key"`
	Name      string      `json:"name"`
	Status    string      `json:"status"`
	CreatedAt string      `json:"created_at"`
	UpdatedAt string      `json:"updated_at"`
	DeletedAt interface{} `json:"deleted_at"`
}

type ClientES struct {
	ID        string      `json:"id"`
	Key       string      `json:"key"`
	Name      string      `json:"name"`
	Status    string      `json:"status"`
	CreatedAt string      `json:"created_at"`
	UpdatedAt string      `json:"updated_at"`
	DeletedAt interface{} `json:"deleted_at"`
}

var ClientStatusEnum = []string{"active", "inactive"}
