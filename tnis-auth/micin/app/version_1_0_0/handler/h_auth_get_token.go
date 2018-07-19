package handler

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/olivere/elastic"

	"tnis/tnis-auth/micin/app/version_1_0_0/model"
	"tnis/tnis-auth/micin/config"
)

func AuthGetToken(w http.ResponseWriter, r *http.Request, db *sql.DB, es *elastic.Client) {
	e := Execution{name: "POST /v100/auth/get_token"}
	e.Start()

	config := config.GetConfig()
	ctx := context.Background()

	var client model.TokenClientES
	var user model.TokenUserES
	var token model.TokenES

	// Params
	client_id := r.FormValue("client_id")
	client_key := r.FormValue("client_key")
	user_code := r.FormValue("user_code")
	user_password := r.FormValue("user_password")

	// Validation
	errors := map[string]interface{}{}
	is_valid := 1
	if !ValidateRequired(client_id) {
		is_valid = 0
		errors["client_id"] = "Client ID is required"
	}
	if !ValidateRequired(client_key) {
		is_valid = 0
		errors["client_key"] = "Client Key is required"
	}
	if !ValidateRequired(user_code) {
		is_valid = 0
		errors["user_code"] = "User Code is required"
	}
	if !ValidateRequired(user_password) {
		is_valid = 0
		errors["user_password"] = "User Password is required"
	}

	if is_valid == 0 {
		RespondInvalidRequestParam(w, r, e, errors)
		return
	}

	// Check valid client
	elastic_source := `	{
						  	"from": 0,
							"size": 1,
						 	"query": {
							  	"bool": {
							    	"must": [
							    		{
								      		"match_phrase": {
								        		"id": {
								          			"query": "` + client_id + `"
								        		}
								      		}
								      	},
								      	{
								      		"match_phrase": {
								        		"key": {
								          			"query": "` + client_key + `"
								        		}
								      		}
								      	},
								      	{
								      		"match_phrase": {
								        		"status": {
								          			"query": "active"
								        		}
								      		}
								      	}
							    	]
							  	}
							}
						}`
	search_result, err := es.Search().
		Index(config.ES.IndiceClient).
		Type("data_type").
		Source(elastic_source).
		Do(ctx)
	if err != nil {
		errors["error"] = err
		RespondApiError(w, r, e, errors)
		return
	}

	if search_result.Hits.TotalHits == 0 {
		errors["client"] = "Invalid client"
		RespondInvalidAuth(w, r, e, errors)
		return
	} else {
		for _, hit := range search_result.Hits.Hits {
			err = json.Unmarshal(*hit.Source, &client)
			if err != nil {
				errors["error"] = err
				RespondApiError(w, r, e, errors)
				return
			}
		}
	}

	// Check valid user
	user_password_md5_ := md5.New()
	user_password_md5_.Write([]byte(user_password))
	user_password_md5 := fmt.Sprintf("%x", user_password_md5_.Sum(nil))
	elastic_source = `	{
						  	"from": 0,
							"size": 1,
						 	"query": {
							  	"bool": {
							    	"must": [
							    		{
								      		"match_phrase": {
								        		"code": {
								          			"query": "` + user_code + `"
								        		}
								      		}
								      	},
								      	{
								      		"match_phrase": {
								        		"password": {
								          			"query": "` + user_password_md5 + `"
								        		}
								      		}
								      	},
								      	{
								      		"match_phrase": {
								        		"status": {
								          			"query": "active"
								        		}
								      		}
								      	}
							    	]
							  	}
							}
						}`
	search_result, err = es.Search().
		Index(config.ES.IndiceUser).
		Type("data_type").
		Source(elastic_source).
		Do(ctx)
	if err != nil {
		errors["error"] = err
		RespondApiError(w, r, e, errors)
		return
	}

	if search_result.Hits.TotalHits == 0 {
		errors["user"] = "Invalid user"
		RespondInvalidAuth(w, r, e, errors)
		return
	} else {
		for _, hit := range search_result.Hits.Hits {
			err = json.Unmarshal(*hit.Source, &user)
			if err != nil {
				errors["error"] = err
				RespondApiError(w, r, e, errors)
				return
			}
		}
	}

	// Old Token vs New Token
	now := now()
	elastic_source = `	{
						  	"from": 0,
							"size": 1,
						 	"query": {
							  	"bool": {
							    	"must": [
							    		{
								      		"match_phrase": {
								        		"client.id": {
								          			"query": "` + client_id + `"
								        		}
								      		}
								      	},
								      	{
								      		"match_phrase": {
								        		"client.key": {
								          			"query": "` + client_key + `"
								        		}
								      		}
								      	},
							    		{
								      		"match_phrase": {
								        		"user.code": {
								          			"query": "` + user_code + `"
								        		}
								      		}
								      	},
								      	{
									        "range": {
								          		"valid_until": {
								            		"gte": "` + now + `"
								          		}
								        	}
								      	}
							    	]
							  	}
							}
						}`
	search_result, err = es.Search().
		Index(config.ES.IndiceToken).
		Type("data_type").
		Source(elastic_source).
		Do(ctx)
	if err != nil {
		errors["error"] = err
		RespondApiError(w, r, e, errors)
		return
	}

	if search_result.Hits.TotalHits > 0 {
		for _, hit := range search_result.Hits.Hits {
			err = json.Unmarshal(*hit.Source, &token)
			if err != nil {
				errors["error"] = err
				RespondApiError(w, r, e, errors)
				return
			}

			fmt.Println("Get last token because token not expired yet")
		}
	} else {
		fmt.Println("Generate new token")

		token_id := strings.Replace(now, "-", "", -1)
		token_id = strings.Replace(token_id, " ", "", -1)
		token_id = strings.Replace(token_id, ":", "", -1)
		token_id += RandomText(114, "123456789")

		now2 := time.Now().Local()
		valid_until := now2.AddDate(1, 0, 1)
		valid_until_date_time := valid_until.Format(dateTimeFormat)

		query := "	INSERT INTO token (`id`,`user_id`,`client_id`,`created_at`,`valid_until`) VALUES (?,?,?,?,?)"
		stmt, err := db.Prepare(query)
		if err != nil {
			errors["error"] = "Error prepare INSERT INTO token " + err.Error()
			RespondApiError(w, r, e, errors)
			return
		}
		defer stmt.Close()
		_, err = stmt.Exec(token_id, user.ID, client.ID, now, valid_until_date_time)
		if err != nil {
			errors["error"] = "Error exec INSERT INTO token " + err.Error()
			RespondApiError(w, r, e, errors)
			return
		}

		// Get Inserted Data
		var user_id int
		rows, _ := db.Query("SELECT `id`,`user_id`,`client_id`,`created_at`,`valid_until` FROM token WHERE id = ?", token_id)
		for rows.Next() {
			err = rows.Scan(&token_id, &user_id, &client_id, &now, &valid_until_date_time)

			token.ID = token_id
			token.User = user
			token.Client = client
			token.CreatedAt = now
			token.ValidUntil = valid_until_date_time
		}

	}

	go InsertUpdateTokenElastic(es, token)

	data := map[string]interface{}{
		"token": token,
	}

	RespondSuccess(w, r, e, data)
	return
}

func InsertUpdateTokenElastic(es *elastic.Client, token model.TokenES) {
	// Insert/Update Elastic Token
	config := config.GetConfig()
	ctx := context.Background()

	insert_update, err := es.Index().
		// _, _ = es.Index().
		Index(config.ES.IndiceToken).
		Type("data_type").
		Id(token.ID).
		BodyJson(token).
		Refresh("true").
		Do(ctx)
	if err != nil {
		fmt.Println("Error InsertUpdateTokenElastic " + err.Error())
	}
	fmt.Println("Insert Token : ", insert_update.Id)
}
