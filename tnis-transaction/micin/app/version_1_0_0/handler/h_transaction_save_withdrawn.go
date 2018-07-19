package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/olivere/elastic"
	"github.com/spf13/cast"

	"tnis/tnis-transaction/micin/app/version_1_0_0/model"
	"tnis/tnis-transaction/micin/config"
)

func TransactionSaveWithdraw(w http.ResponseWriter, r *http.Request, db *sql.DB, es *elastic.Client) {
	e := Execution{name: "POST /v100/transaction/{save/withdraw}"}
	e.Start()

	config := config.GetConfig()
	ctx := context.Background()

	var transaction model.TransactionES
	var transaction_customer model.TransactionCustomerES
	var customer model.CustomerES

	// Params
	customer_id := r.FormValue("customer_id")
	category := r.FormValue("category")
	total := cast.ToInt(r.FormValue("total"))

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
	if !ValidateRequired(customer_id) {
		is_valid = 0
		errors["customer_id"] = "Customer is required"
	}
	if !ValidateRequired(category) {
		is_valid = 0
		errors["category"] = "Category is required"
	} else {
		if !ValidateInArray(category, model.TransactionCategoryEnum) {
			is_valid = 0
			errors["category"] = "Category is not in list"
		}
	}

	if is_valid == 0 {
		RespondInvalidRequestParam(w, r, e, errors)
		return
	}

	// Check valid ID
	elastic_source := `	{
						  	"from": 0,
							"size": 1,
						 	"query": {
							  	"bool": {
							    	"must": [ 
							    		{
								      		"match_phrase": {
								        		"id": {
								          			"query": "` + cast.ToString(cast.ToInt(customer_id)) + `"
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
			transaction_customer.ID = customer.ID
			transaction_customer.AccountNumber = customer.AccountNumber
			transaction_customer.IdentityCard = customer.IdentityCard
			transaction_customer.Name = customer.Name
			transaction_customer.Email = customer.Email
			transaction_customer.PhoneNumber = customer.PhoneNumber
		}
	}

	if category == "in" {
		customer.Total += total
	} else if category == "out" {
		if (customer.Total - total) <= 0 { // TODO: next customer can have level (Gold, Member, Platinum) to check the minimum saldo that cann't be withdrawn
			errors["error"] = "Your balance is not sufficient to withdraw"
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
		}
		customer.Total -= total
	}

	// Insert MySQL
	date := now()

	query := "	INSERT INTO transaction (`date`,`customer_id`,`user_id`,`category`,`total`) VALUES (?,?,?,?,?)"
	stmt, err := db.Prepare(query)
	if err != nil {
		errors["error"] = "Error prepare INSERT INTO transaction " + err.Error()
		RespondApiError(w, r, e, errors)
		return
	}
	defer stmt.Close()
	res, err := stmt.Exec(date, customer_id, token_detail.User.ID, category, total)
	if err != nil {
		errors["error"] = "Error exec INSERT INTO transaction " + err.Error()
		RespondApiError(w, r, e, errors)
		return
	}

	// Get Inserted Data
	id, _ := res.LastInsertId()

	var user_id int
	rows, _ := db.Query("SELECT `id`,`date`,`customer_id`,`user_id`,`category`,`total` FROM `transaction` WHERE id = ?", id)
	for rows.Next() {
		err = rows.Scan(&id, &date, &customer_id, &user_id, &category, &total)

		transaction.ID = cast.ToInt(id)
		transaction.Date = date
		transaction.Customer = transaction_customer
		transaction.User = token_detail.User
		transaction.Category = category
		transaction.Total = total
	}

	go InsertUpdateTransactionElastic(es, transaction)
	go InsertUpdateCustomerElastic(es, customer)
	go NotifSendEmail(authorization, cast.ToString(id))

	data := map[string]interface{}{
		"transaction": transaction,
		"customer":    customer,
	}

	RespondSuccess(w, r, e, data)
	return
}

func InsertUpdateTransactionElastic(es *elastic.Client, transaction model.TransactionES) {
	// Insert/Update Elastic Transaction
	config := config.GetConfig()
	ctx := context.Background()

	// insert_update, err := es.Index().
	_, _ = es.Index().
		Index(config.ES.IndiceTransaction).
		Type("data_type").
		Id(cast.ToString(transaction.ID)).
		BodyJson(transaction).
		Refresh("true").
		Do(ctx)
	// if err != nil {
	// 	fmt.Println("Error InsertUpdateTransactionElastic " + err.Error())
	// }
	// fmt.Println("Insert Transaction : ", insert_update.Id)
}

func InsertUpdateCustomerElastic(es *elastic.Client, customer model.CustomerES) {
	// Insert/Update Elastic Customer
	config := config.GetConfig()
	ctx := context.Background()

	// insert_update, err := es.Index().
	_, _ = es.Index().
		Index(config.ES.IndiceCustomer).
		Type("data_type").
		Id(cast.ToString(customer.ID)).
		BodyJson(customer).
		Refresh("true").
		Do(ctx)
	// if err != nil {
	// 	fmt.Println("Error InsertUpdateCustomerElastic " + err.Error())
	// }
	// fmt.Println("Insert Customer : ", insert_update.Id)
}
