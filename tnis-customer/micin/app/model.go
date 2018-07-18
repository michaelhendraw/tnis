package app

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/olivere/elastic"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
	ES     *elastic.Client
}

type Empty struct {
}

type Data map[string]interface{}

type Respon struct {
	Status  string
	Code    int
	Message string
	ExeTime string
	Data    interface{}
	Error   interface{}
}

type Execution struct {
	name     string
	startExe time.Time
	endExe   time.Time
	exeTime  string
}
