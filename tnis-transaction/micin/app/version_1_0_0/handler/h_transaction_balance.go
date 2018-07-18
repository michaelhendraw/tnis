package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/olivere/elastic"

	"tnis-transaction/micin/app/version_1_0_0/model"
	"tnis-transaction/micin/config"
)

func TransactionBalance(w http.ResponseWriter, r *http.Request, db *sql.DB, es *elastic.Client) {
	e := Execution{name: "POST /v100/transaction/balance/{account_number}"}
	e.Start()

	config := config.GetConfig()
	ctx := context.Background()

	var customer model.CustomerShowES

	// Params
	vars := mux.Vars(r)
	account_number := vars["account_number"]

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
	if !ValidateRequired(account_number) {
		is_valid = 0
		errors["account_number"] = "Account Number is required"
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
								        		"account_number": {
								          			"query": "` + account_number + `"
								        		}
								      		}
								      	}
								    ]
							  	}
							}
						}`
	search_result, err := es.Search().
		Index(config.ES.IndiceCustomer).
		Type("data_type").
		Source(elastic_source).
		Do(ctx)
	if err != nil {
		errors["error"] = err
		RespondApiError(w, r, e, errors)
		return
	}

	if search_result.Hits.TotalHits == 0 {
		errors["customer"] = "Customer not found"
		RespondDataNotFound(w, r, e, errors)
		return
	} else {
		for _, hit := range search_result.Hits.Hits {
			err = json.Unmarshal(*hit.Source, &customer)
			if err != nil {
				errors["error"] = err
				RespondApiError(w, r, e, errors)
				return
			}
		}
	}

	data := map[string]interface{}{
		"customer": customer,
	}

	RespondSuccess(w, r, e, data)
	return
}
