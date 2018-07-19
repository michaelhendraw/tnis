package handler

import (
	"context"
	"database/sql"
	"net/http"
	"reflect"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/olivere/elastic"
	"github.com/spf13/cast"

	"tnis/tnis-auth/micin/app/version_1_0_0/model"
	"tnis/tnis-auth/micin/config"
)

func AuthCreate(w http.ResponseWriter, r *http.Request, db *sql.DB, es *elastic.Client) {
	e := Execution{name: "POST /v100/auth/create"}
	e.Start()

	config := config.GetConfig()
	ctx := context.Background()

	// Params
	name := r.FormValue("name")

	// Validation
	errors := map[string]interface{}{}
	is_valid := 1
	if !ValidateRequired(name) {
		is_valid = 0
		errors["name"] = "Name is required"
	}

	if is_valid == 0 {
		RespondInvalidRequestParam(w, r, e, errors)
		return
	}

	// Check redudancy
	elastic_source := `	{
						  	"from": 0,
							"size": 1,
						 	"query": {
							  	"bool": {
							    	"must": {
							      		"match_phrase": {
							        		"name": {
							          			"query": "` + name + `"
							        		}
							      		}
							    	}
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

	if search_result.Hits.TotalHits != 0 {
		errors["name"] = "Name already exist"
		RespondDataExist(w, r, e, errors)
		return
	}

	// Insert MySQL
	created_at := now()
	updated_at := created_at
	id := strings.Replace(created_at, "-", "", -1)
	id = strings.Replace(id, " ", "", -1)
	id = strings.Replace(id, ":", "", -1)
	id += RandomText(2, "123456789")
	key := RandomText(32, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	query := "	INSERT INTO client (`id`,`key`,`name`,`status`,`created_at`,`updated_at`,`deleted_at`) VALUES (?,?,?,?,?,?,?)"
	stmt, err := db.Prepare(query)
	if err != nil {
		errors["error"] = "Error prepare INSERT INTO client " + err.Error()
		RespondApiError(w, r, e, errors)
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(id, key, name, "active", created_at, updated_at, nil)
	if err != nil {
		errors["error"] = "Error exec INSERT INTO client " + err.Error()
		RespondApiError(w, r, e, errors)
		return
	}

	// Get Inserted Data
	var client model.ClientDB
	data := map[string]interface{}{
		"client": client,
	}

	var status string
	var deleted_at interface{}
	rows, _ := db.Query("SELECT `id`,`key`,`name`,`status`,`created_at`,`updated_at`,`deleted_at` FROM client WHERE id = ?", id)
	for rows.Next() {
		err = rows.Scan(&id, &key, &name, &status, &created_at, &updated_at, &deleted_at)

		client.ID = id
		client.Key = key
		client.Name = name
		client.Status = status
		client.CreatedAt = created_at
		client.UpdatedAt = updated_at
		if reflect.TypeOf(deleted_at) != nil {
			client.DeletedAt = cast.ToString(deleted_at)
		} else {
			client.DeletedAt = deleted_at
		}
	}

	go InsertUpdateClientElastic(es, client)

	data = map[string]interface{}{
		"client": client,
	}

	RespondSuccess(w, r, e, data)
	return
}

func InsertUpdateClientElastic(es *elastic.Client, client model.ClientDB) {
	// Insert/Update Elastic Client
	config := config.GetConfig()
	ctx := context.Background()

	// insert_update, err := es.Index().
	_, _ = es.Index().
		Index(config.ES.IndiceClient).
		Type("data_type").
		Id(client.ID).
		BodyJson(client).
		Refresh("true").
		Do(ctx)
	// if err != nil {
	// 	fmt.Println("Error InsertUpdateClientElastic " + err.Error())
	// }
	// fmt.Println("Insert Client : ", insert_update.Id)
}
