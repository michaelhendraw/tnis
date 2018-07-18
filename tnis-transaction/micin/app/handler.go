package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
	"reflect"
	"strconv"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/olivere/elastic"
)

// Versioning (Versi Client : Versi API)
var app_api_version = map[string]map[string]string{
	// android
	"1583516695817679": map[string]string{
		"100": "100",
		"101": "100",
		"102": "100",
		"200": "200",
	},
	// ios
	"1894514824971982": map[string]string{
		"100": "100",
		"101": "100",
		"200": "200",
	},
	// web
	"1938221925812416": map[string]string{
		"100": "100",
		"200": "200",
	},
}

var clients = map[string]map[string][]string{
	// android
	"1583516695817679": map[string][]string{
		"last_version":        {"200"},
		"must_update_version": {"102"},
	},
	// ios
	"1894514824971982": map[string][]string{
		"last_version":        {"200"},
		"must_update_version": {"101"},
	},
	// web
	"1938221925812416": map[string][]string{
		"last_version":        {"200"},
		"must_update_version": {},
	},
}

// Execution Time
var dateTimeExecutionFormat = "2006-01-02 15:04:05.000000"

func (e *Execution) Start() {
	e.startExe = time.Now()
	fmt.Println("\n--- START", e.name, "on", e.startExe.Format(dateTimeExecutionFormat), "---\n")
}

func (e *Execution) End() string {
	e.endExe = time.Now()
	e.exeTime = strconv.FormatInt(int64(time.Since(e.startExe)/time.Millisecond), 10) + "ms"
	fmt.Println("\n--- END", e.name, "on", e.endExe.Format(dateTimeExecutionFormat), "time execute", e.exeTime, "---")
	return e.exeTime
}

// Respon JSON
func RespondJson(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*\n")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, X-Auth-Token\n")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.WriteHeader(status)
	w.Write([]byte(response))
}

func RespondApiNotFound(w http.ResponseWriter, r *http.Request) {
	respon := Respon{
		Status:  "api_not_found",
		Code:    404,
		Message: "API not found",
		ExeTime: "",
		Data:    Empty{},
		Error:   "API not found",
	}
	RespondJson(w, http.StatusOK, respon)
	return
}

func RespondApiError(w http.ResponseWriter, r *http.Request, text interface{}) {
	respon := Respon{
		Status:  "api_error",
		Code:    500,
		Message: "API Error",
		ExeTime: "",
		Data:    Empty{},
		Error:   text,
	}
	RespondJson(w, http.StatusOK, respon)
	return
}

// String
func InsertEveryCharacter(s string, n int, c rune) string {
	var buffer bytes.Buffer
	var n_1 = n - 1
	var l_1 = len(s) - 1
	for i, rune := range s {
		buffer.WriteRune(rune)
		if i%n == n_1 && i != l_1 {
			buffer.WriteRune(c)
		}
	}
	return buffer.String()
}

// Validate
func InArray(val interface{}, array interface{}) (exists bool) {
	exists = false

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)
		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				exists = true
				return
			}
		}
	}
	return
}

// ElasticSearch
type MyRetrier struct {
	backoff elastic.Backoff
}

func NewMyRetrier() *MyRetrier {
	return &MyRetrier{
		backoff: elastic.NewExponentialBackoff(10*time.Millisecond, 8*time.Second),
	}
}

func (r *MyRetrier) Retry(ctx context.Context, retry int, req *http.Request, resp *http.Response, err error) (time.Duration, bool, error) {
	if err == syscall.ECONNREFUSED {
		return 0, false, errors.New("Elasticsearch or network down")
	}

	if retry >= 2 {
		return 0, false, nil
	}

	wait, stop := r.backoff.Next(retry)
	return wait, stop, nil
}

// Others
func Recovery(full_func string) {
	if r := recover(); r != nil {
		fmt.Println("recovered: You did not create a function:", full_func, r)
	}
}

func Welcome(w http.ResponseWriter, r *http.Request) {
	ex, err := os.Executable()
	if err != nil {
		RespondApiError(w, r, err.Error())
	}

	dir := path.Dir(ex)
	var viewWelcome = template.Must(template.ParseFiles(dir + "/static/welcome.html"))
	err = viewWelcome.Execute(w, nil)
	if err != nil {
		RespondApiNotFound(w, r)
	}
}
