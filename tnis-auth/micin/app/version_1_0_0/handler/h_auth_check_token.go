package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/olivere/elastic"
	"github.com/spf13/cast"

	"tnis/tnis-auth/micin/app/version_1_0_0/model"
	"tnis/tnis-auth/micin/config"
)

func AuthCheckToken(w http.ResponseWriter, r *http.Request, db *sql.DB, es *elastic.Client) {
	e := Execution{name: "POST /v100/auth/chech_token"}
	e.Start()

	config := config.GetConfig()
	ctx := context.Background()

	var token model.TokenES

	// Params
	token_param := r.FormValue("token")
	service := r.FormValue("service")

	// Validation
	errors := map[string]interface{}{}
	is_valid := 1
	if !ValidateRequired(token_param) {
		is_valid = 0
		errors["token"] = "Token is required"
	}
	if !ValidateRequired(service) {
		is_valid = 0
		errors["service"] = "Service is required"
	}

	if is_valid == 0 {
		RespondInvalidRequestParam(w, r, e, errors)
		return
	}

	// Check valid token
	now := now()
	elastic_source := `	{
						  	"from": 0,
							"size": 1,
						 	"query": {
							  	"bool": {
							    	"must": [
							    		{
								      		"match_phrase": {
								        		"id": {
								          			"query": "` + token_param + `"
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
	search_result, err := es.Search().
		Index(config.ES.IndiceToken).
		Type("data_type").
		Source(elastic_source).
		Do(ctx)
	if err != nil {
		errors["error"] = err
		RespondApiError(w, r, e, errors)
		return
	}

	if search_result.Hits.TotalHits == 0 {
		errors["token"] = "Invalid token"
		RespondInvalidToken(w, r, e, errors)
		return
	} else {
		for _, hit := range search_result.Hits.Hits {
			err = json.Unmarshal(*hit.Source, &token)
			if err != nil {
				errors["error"] = err
				RespondApiError(w, r, e, errors)
				return
			}
		}
	}

	// Check valid service user
	elastic_source = `	{
						  	"from": 0,
							"size": 1,
						 	"query": {
							  	"bool": {
							    	"must": [
							    		{
								      		"match_phrase": {
								        		"id": {
								          			"query": "` + cast.ToString(token.User.ID) + `"
								        		}
								      		}
								      	},
								      	{
								      		"match_phrase": {
								        		"services.service": {
								          			"query": "` + service + `"
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
		errors["service"] = "Unauthorized"
		RespondUnathorize(w, r, e, errors)
		return
	}

	data := map[string]interface{}{
		"token": token,
	}

	RespondSuccess(w, r, e, data)
	return
}
