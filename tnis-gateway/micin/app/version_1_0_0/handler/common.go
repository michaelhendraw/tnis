package handler

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	rand2 "math/rand"
	"net/http"
	"reflect"
	"strconv"
	"time"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/leekchan/accounting"
	"github.com/olivere/elastic"
	"github.com/spf13/cast"
)

// Status Code
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

type MainParam struct {
	W  http.ResponseWriter
	R  *http.Request
	DB *sql.DB
	ES *elastic.Client
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

func RespondApiError(w http.ResponseWriter, r *http.Request, e Execution, text interface{}) {
	status := "api_error"
	fmt.Println(status, text)
	respon := Respon{
		Status:  status,
		Code:    500,
		Message: "API Error",
		ExeTime: e.End(),
		Data:    Empty{},
		Error:   text,
	}
	RespondJson(w, http.StatusOK, respon)
	return
}

func RespondDataNotFound(w http.ResponseWriter, r *http.Request, e Execution, text interface{}) {
	status := "data_not_found"
	fmt.Println(status, text)
	respon := Respon{
		Status:  status,
		Code:    404,
		Message: "Data not found",
		ExeTime: e.End(),
		Data:    Empty{},
		Error:   text,
	}
	RespondJson(w, http.StatusOK, respon)
	return
}

func RespondDataExist(w http.ResponseWriter, r *http.Request, e Execution, text interface{}) {
	status := "data_already_exist"
	fmt.Println(status, text)
	respon := Respon{
		Status:  status,
		Code:    404,
		Message: "Data already exist",
		ExeTime: e.End(),
		Data:    Empty{},
		Error:   text,
	}
	RespondJson(w, http.StatusOK, respon)
	return
}

func RespondInvalidAuth(w http.ResponseWriter, r *http.Request, e Execution, text interface{}) {
	status := "invalid_auth"
	fmt.Println(status, text)
	respon := Respon{
		Status:  status,
		Code:    401,
		Message: "Invalid Auth",
		ExeTime: e.End(),
		Data:    Empty{},
		Error:   text,
	}
	RespondJson(w, http.StatusOK, respon)
	return
}

func RespondInvalidToken(w http.ResponseWriter, r *http.Request, e Execution, text interface{}) {
	status := "invalid_token"
	fmt.Println(status, text)
	respon := Respon{
		Status:  status,
		Code:    401,
		Message: "Invalid Token",
		ExeTime: e.End(),
		Data:    Empty{},
		Error:   text,
	}
	RespondJson(w, http.StatusOK, respon)
	return
}

func RespondUnathorize(w http.ResponseWriter, r *http.Request, e Execution, text interface{}) {
	status := "unauthorized"
	fmt.Println(status, text)
	respon := Respon{
		Status:  status,
		Code:    401,
		Message: "Unauthorized",
		ExeTime: e.End(),
		Data:    Empty{},
		Error:   text,
	}
	RespondJson(w, http.StatusOK, respon)
	return
}

func RespondInvalidRequestParam(w http.ResponseWriter, r *http.Request, e Execution, text interface{}) {
	status := "invalid_request_param"
	fmt.Println(status, text)
	respon := Respon{
		Status:  status,
		Code:    400,
		Message: "Invalid request param",
		ExeTime: e.End(),
		Data:    Empty{},
		Error:   text,
	}
	RespondJson(w, http.StatusOK, respon)
	return
}

func RespondSuccess(w http.ResponseWriter, r *http.Request, e Execution, text interface{}) {
	status := "success"
	fmt.Println(status, text)
	respon := Respon{
		Status:  status,
		Code:    200,
		Message: "Success",
		ExeTime: e.End(),
		Data:    text,
		Error:   Empty{},
	}
	RespondJson(w, http.StatusOK, respon)
	return
}

// Encrypt and Decrypt
func Encrypt(key []byte, text string) string {
	plaintext := []byte(text)
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
	return base64.URLEncoding.EncodeToString(ciphertext)
}

func Decrypt(key []byte, secure_mess string) (decoded_mess string, err error) {
	cipherText, err := base64.URLEncoding.DecodeString(secure_mess)
	if err != nil {
		return
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	if len(cipherText) < aes.BlockSize {
		err = errors.New("Ciphertext block size is too short!")
		return
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	decoded_mess = string(cipherText)
	return
}

// String
var seededRand *rand2.Rand = rand2.New(rand2.NewSource(time.Now().UnixNano()))

func RandomText(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func StringToCurrency(text string) string {
	ac := accounting.Accounting{Symbol: "", Precision: 2}
	return ac.FormatMoney(cast.ToInt(text))
}

// Date Time
var dateTimeFormat = "2006-01-02 15:04:05"
var dateFormat = "2006-01-02"

// func now(format string) string {
// 	if format != "" {
// 		dateTimeFormat = format
// 	}
func now() string {
	now := time.Now().Local()
	return now.Format(dateTimeFormat)
}

// Validate
func ValidateInArray(val interface{}, array interface{}) (exists bool) {
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

func ValidateRequired(text string) (response bool) {
	response = false

	if len(text) != 0 {
		response = true
		return
	}
	return
}

// random text for save photo and video
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand2.Intn(len(letterBytes))]
	}
	return string(b)
}
