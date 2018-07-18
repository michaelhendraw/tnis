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
	"github.com/spf13/cast"

	"tnis-customer/micin/app/version_1_0_0/model"
	"tnis-customer/micin/config"
)

func CustomerUpdate(w http.ResponseWriter, r *http.Request, db *sql.DB, es *elastic.Client) {
	e := Execution{name: "PUT /v100/customer/update/{id}"}
	e.Start()

	config := config.GetConfig()
	ctx := context.Background()

	var customer model.CustomerES

	// Params
	vars := mux.Vars(r)
	id := vars["id"]
	identity_card := r.FormValue("identity_card")
	name := r.FormValue("name")
	birth_date := r.FormValue("birth_date")
	gender := r.FormValue("gender")
	address := r.FormValue("address")
	email := r.FormValue("email")
	phone_number := r.FormValue("phone_number")

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
		errors["id"] = "ID is required"
	}
	if !ValidateRequired(identity_card) {
		is_valid = 0
		errors["identity_card"] = "Identity Card is required"
	}
	if !ValidateRequired(name) {
		is_valid = 0
		errors["name"] = "Name is required"
	}
	if !ValidateRequired(birth_date) {
		is_valid = 0
		errors["birth_date"] = "Birth Date is required"
	}
	if !ValidateRequired(gender) {
		is_valid = 0
		errors["gender"] = "Gender is required"
	} else {
		if !ValidateInArray(gender, model.CustomerGenderEnum) {
			is_valid = 0
			errors["gender"] = "Gender is not in list"
		}
	}
	if !ValidateRequired(address) {
		is_valid = 0
		errors["address"] = "Address is required"
	}
	if !ValidateRequired(email) {
		is_valid = 0
		errors["email"] = "Email is required"
	}
	if !ValidateRequired(phone_number) {
		is_valid = 0
		errors["phone_number"] = "Phone Number is required"
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
								          			"query": "` + cast.ToString(cast.ToInt(id)) + `"
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

	// Check redudancy by identity card, name, birth date, gender
	elastic_source = `	{
						  	"from": 0,
							"size": 1,
						 	"query": {
							  	"bool": {
							    	"must": [ 
							    		{
								      		"match_phrase": {
								        		"identity_card": {
								          			"query": "` + identity_card + `"
								        		}
								      		}
								      	},
								      	{
								      		"match_phrase": {
								        		"name": {
								          			"query": "` + name + `"
								        		}
								      		}
								      	},
								      	{
								      		"match_phrase": {
								        		"birth_date": {
								          			"query": "` + birth_date + `"
								        		}
								      		}
								      	},
								      	{
								      		"match_phrase": {
								        		"gender": {
								          			"query": "` + gender + `"
								        		}
								      		}
								      	}
								    ],
								    "must_not": [ 
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
	search_result, err = es.Search().
		Index(config.ES.IndiceCustomer).
		Type("data_type").
		Source(elastic_source).
		Do(ctx)
	if err != nil {
		errors["error"] = err
		RespondApiError(w, r, e, errors)
		return
	}

	if search_result.Hits.TotalHits != 0 {
		errors["customer"] = "Customer already exist, same identity card, name, birth date, gender"
		RespondDataExist(w, r, e, errors)
		return
	}

	// Update MySQL
	updated_at := now()

	query := "	UPDATE customer SET `identity_card`=?,`name`=?,`birth_date`=?,`gender`=?,`address`=?,`email`=?,`phone_number`=?,`updated_at`=?,`updated_by`=? WHERE id=?"
	stmt, err := db.Prepare(query)
	if err != nil {
		errors["error"] = "Error prepare UPDATE customer " + err.Error()
		RespondApiError(w, r, e, errors)
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(identity_card, name, birth_date, gender, address, email, phone_number, updated_at, token_detail.User.ID, id)
	if err != nil {
		errors["error"] = "Error exec UPDATE customer " + err.Error()
		RespondApiError(w, r, e, errors)
		return
	}

	// Get Updated Data
	var account_number, status, created_at string
	var created_by, updated_by int
	rows, _ := db.Query("SELECT `id`,`account_number`,`identity_card`,`name`,`birth_date`,`gender`,`address`,`email`,`phone_number`,`status`,`created_at`,`created_by`,`updated_at`,`updated_by` FROM customer WHERE id = ?", id)
	for rows.Next() {
		err = rows.Scan(&id, &account_number, &identity_card, &name, &birth_date, &gender, &address, &email, &phone_number, &status, &created_at, &created_by, &updated_at, &updated_by)

		customer.ID = cast.ToInt(id)
		customer.AccountNumber = account_number
		customer.IdentityCard = identity_card
		customer.Name = name
		customer.BirthDate = birth_date
		customer.Gender = gender
		customer.Address = address
		customer.Email = email
		customer.PhoneNumber = phone_number
		customer.Status = status
		customer.CreatedAt = created_at
		customer.CreatedBy = created_by
		customer.UpdatedAt = updated_at
		customer.UpdatedBy = updated_by
	}

	go InsertUpdateCustomerElastic(es, customer)

	data := map[string]interface{}{
		"customer": customer,
	}

	RespondSuccess(w, r, e, data)
	return
}
