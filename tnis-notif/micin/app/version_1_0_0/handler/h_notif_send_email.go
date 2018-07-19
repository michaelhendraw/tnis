package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/olivere/elastic"
	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/spf13/cast"

	"tnis/tnis-notif/micin/app/version_1_0_0/model"
	"tnis/tnis-notif/micin/config"
)

func NotifSendEmail(w http.ResponseWriter, r *http.Request, db *sql.DB, es *elastic.Client) {
	e := Execution{name: "GET /v100/notif/send_email/{id}"}
	e.Start()

	config := config.GetConfig()
	ctx := context.Background()

	var transaction model.TransactionES

	// Params
	vars := mux.Vars(r)
	id := vars["id"]

	authorization := r.Header.Get("authorization")

	// Validation
	errors := map[string]interface{}{}

	token := ""
	s := strings.Split(authorization, " ")
	if len(s) >= 2 {
		token = s[1]
	} else {
		errors["token"] = "Token is required"
		RespondInvalidToken(w, r, e, errors)
		return
	}
	var token_detail model.TokenES
	token_detail = GetTokenDetail(es, token)
	if token_detail.User.ID == 0 {
		errors["token"] = "Invalid token"
		RespondInvalidToken(w, r, e, errors)
		return
	}

	is_valid := 1
	if !ValidateRequired(id) {
		is_valid = 0
		errors["id"] = "Transaction ID is required"
	}

	if is_valid == 0 {
		RespondInvalidRequestParam(w, r, e, errors)
		return
	}

	// Check valid account number
	elastic_source := `	{
						  	"from": 0,
							"size": 1,
						 	"query": {
							  	"bool": {
							    	"must": [ 
							    		{
								      		"match_phrase": {
								        		"id": {
								          			"query": "` + id + `"
								        		}
								      		}
								      	}
								    ]
							  	}
							}
						}`
	search_result, err := es.Search().
		Index(config.ES.IndiceTransaction).
		Type("data_type").
		Source(elastic_source).
		Do(ctx)
	if err != nil {
		errors["error"] = err
		RespondApiError(w, r, e, errors)
		return
	}

	if search_result.Hits.TotalHits == 0 {
		errors["transaction"] = "Transaction not found"
		RespondDataNotFound(w, r, e, errors)
		return
	} else {
		for _, hit := range search_result.Hits.Hits {
			err = json.Unmarshal(*hit.Source, &transaction)
			if err != nil {
				errors["error"] = err
				RespondApiError(w, r, e, errors)
				return
			}
		}
	}

	// Send Email
	// TODO : to make custom subject and content from database
	subject := ""
	content := ""
	if transaction.Category == "in" {
		subject = "OCBCNISP Notification Success Cash Deposit"
		content = "Congratulations " + transaction.User.Name + ",<br><br>Your cash deposit of " + StringToCurrency(cast.ToString(transaction.Total)) + " has been successfully done.<br><br>Best Regards,<br>OCBCNISP"
	} else if transaction.Category == "out" {
		subject = "OCBCNISP Notification Success Withdrawn"
		content = "Congratulations " + transaction.User.Name + ",<br><br>Your withdrawn of " + StringToCurrency(cast.ToString(transaction.Total)) + " has been successfully done.<br><br>Best Regards,<br>OCBCNISP"
	}
	from_name := config.EMAIL.FromName
	from_address := config.EMAIL.FromAddress
	to_name := transaction.User.Name
	to_address := transaction.User.Email

	from := mail.NewEmail(from_name, from_address)
	to := mail.NewEmail(to_name, to_address)
	message := mail.NewSingleEmail(from, subject, to, subject, content)
	client := sendgrid.NewSendClient(config.EMAIL.EmailAPI)
	_, err = client.Send(message)

	if err != nil {
		errors["error"] = "can't send email"
		status := "failed"
		fmt.Println(status, errors)
		respon := Respon{
			Status:  status,
			Code:    200,
			Message: "failed",
			ExeTime: e.End(),
			Data:    errors,
			Error:   Empty{},
		}
		RespondJson(w, http.StatusOK, respon)
		return
	} else {
		data := map[string]interface{}{
			"transaction": transaction,
		}
		status := "success"
		fmt.Println(status, errors)
		respon := Respon{
			Status:  status,
			Code:    200,
			Message: "Success",
			ExeTime: e.End(),
			Data:    data,
			Error:   Empty{},
		}
		RespondJson(w, http.StatusOK, respon)
		return
	}
}
