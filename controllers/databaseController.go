package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

type (
	IDatabaseController interface {
		SetData(w http.ResponseWriter, r *http.Request)
		GetData(w http.ResponseWriter, r *http.Request)
		DeleteData(w http.ResponseWriter, r *http.Request)
		SearchData(w http.ResponseWriter, r *http.Request)

		GetMetric(w http.ResponseWriter, r *http.Request)

		BackupData()
	}

	databaseController struct {
		controller
		FakeDatabase    map[string]interface{} `json:"fake_database"`
		FakeDatabaseTTL map[string]time.Time   `json:"fake_database_ttl"`
		Metrics         map[string]int         `json:"metrics"`
	}
)

func GetDatabaseController() IDatabaseController {
	c := &databaseController{
		FakeDatabase:    map[string]interface{}{},
		FakeDatabaseTTL: map[string]time.Time{},
		Metrics:         make(map[string]int),
	}

	//Check if data.txt is preset, if so load from that
	if fileInfo, err := os.Stat("./data.txt"); fileInfo != nil && err == nil {
		b, err := ioutil.ReadFile("./data.txt")
		if err != nil {
			panic("Could not initalize from file")
		}

		err = json.Unmarshal(b, c)
		if err != nil {
			panic("Could not load data!")
		}
	}

	return c
}

/*
A "Set" endpoint:
It should use the POST HTTP method
It should be available at the /set url path
The request should contain a JSON POST body consisting of a JSON entry of the following form: {"key": "<some key>", "value": "<string value>" }.
This endpoint should set the key value pair in memory.
The keys and values should be strings only
If successful, the response should be the JSON object that was just set.
*/
type dbData struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}
type setDataPayload struct {
	dbData
	TimeToLive int64 `json:"ttl"` //Number of seconds for data to last
}

func (this *databaseController) SetData(w http.ResponseWriter, r *http.Request) {
	payload := []setDataPayload{}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		this.WriteErrorMessageWithStatus(w, "Invalid payload", 400)
		return
	}

	for _, data := range payload {
		if data.Key == "" || data.Value == "" {
			this.WriteErrorMessageWithStatus(w, "Invalid payload", 400)
			return
		}

		this.FakeDatabase[data.Key] = data.Value

		//If TimeToLive is set, set it up
		if data.TimeToLive != 0 {
			this.FakeDatabaseTTL[data.Key] = time.Now().Add(time.Second * time.Duration(data.TimeToLive))
		} else {
			delete(this.FakeDatabaseTTL, data.Key)
		}
	}

	if this.Metrics == nil {
		this.Metrics = map[string]int{}
	}
	this.Metrics["Set"] = this.Metrics["Set"] + 1
	this.BackupData()
	this.ProcessSuccessWithResponseBody(w, payload, 200)
}

/*
A "Get" endpoint
It should use the GET HTTP method
It should be available at the /get url path
The request should be of the form: /get?key=someKey, which should return the value stored in memory for "someKey".
The response body should be a JSON object of the following form: {"key": "<some key>", "value": "<string value>" }.
If called with no parameters, it should return an error.
*/
func (this *databaseController) GetData(w http.ResponseWriter, r *http.Request) {
	keys := r.URL.Query()["keys"]
	if len(keys) < 1 {
		this.WriteErrorMessageWithStatus(w, "invalid_keys", 400)
		return
	}

	resp := []dbData{}
	errCount := 0
	for _, key := range keys {
		data, ok := this.FakeDatabase[key]
		if !ok {
			found := false
			for k, val := range this.FakeDatabase {
				if strings.HasPrefix(k, key) {
					found = true
					remaining := strings.TrimPrefix(k, key+".")
					subObjs := strings.Split(remaining, ".")

					prev := map[string]interface{}{}
					data = prev
					for idx, fieldName := range subObjs {
						if idx >= len(subObjs)-1 {
							prev[fieldName] = val
							break
						}
						t := map[string]interface{}{}
						prev[fieldName] = t
						prev = t
					}
				}
			}

			if !found {
				errCount++
				continue
			}
		}

		if expires, ok := this.FakeDatabaseTTL[key]; ok && !expires.IsZero() && time.Now().After(expires) {
			fmt.Println("Key is Expired")
			errCount++
			continue
		}

		resp = append(resp, dbData{
			Key:   key,
			Value: data,
		})
	}

	if errCount >= len(keys) {
		this.WriteErrorMessageWithStatus(w, "keys_not_found", 404)
		return
	}

	this.Metrics["Get"]++
	this.ProcessSuccessWithResponseBody(w, resp, 200)
}

/*
A "Delete" endpoint
It should use the POST HTTP method
It should be available at the /delete url path
The request should contain a JSON POST body consisting of a JSON entry of the following form: {"key": "<some key>"}.
If the key was successfully deleted, the response should be a JSON document including the key that was deleted.
If called with no parameters, should return an error.
If the key doesn't exist, it should respond with a 200, but should contain a message saying the key didn't exist.
*/
type deleteRequest struct {
	Key string `json:"key"`
}

func (this *databaseController) DeleteData(w http.ResponseWriter, r *http.Request) {
	payload := deleteRequest{}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil || payload.Key == "" {
		this.WriteErrorMessageWithStatus(w, "Invalid payload", 400)
		return
	}

	data, ok := this.FakeDatabase[payload.Key]
	if !ok {
		resp := map[string]interface{}{
			"message": "key_not_found",
		}
		this.ProcessSuccessWithResponseBody(w, resp, 200)
		return
	}
	delete(this.FakeDatabase, payload.Key)
	this.Metrics["Delete"]++

	this.BackupData()

	resp := dbData{
		Key:   payload.Key,
		Value: data,
	}
	this.ProcessSuccessWithResponseBody(w, resp, 200)
}

func (this *databaseController) SearchData(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")

	for key, val := range this.FakeDatabase {
		if strings.Contains(key, keyword) {
			resp := dbData{
				Key:   key,
				Value: val,
			}

			if expires, ok := this.FakeDatabaseTTL[key]; ok && time.Now().After(expires) {
				fmt.Println("Key is Expired")
				this.WriteErrorMessageWithStatus(w, "key_not_found", 404)
				return
			}

			this.ProcessSuccessWithResponseBody(w, resp, 200)
			return
		}
	}

	resp := map[string]interface{}{
		"message": "key_not_found",
	}

	this.Metrics["Search"]++
	this.ProcessSuccessWithResponseBody(w, resp, 404)
	return
}

func (this *databaseController) GetMetric(w http.ResponseWriter, r *http.Request) {
	metric := r.URL.Query().Get("metric")
	if metric == "" {
		this.WriteErrorMessageWithStatus(w, "invalid_metric", 400)
		return
	}

	data, ok := this.Metrics[metric]
	if !ok {
		this.WriteErrorMessageWithStatus(w, "metric_not_found", 404)
		return
	}

	resp := map[string]interface{}{
		"metric": metric,
		"count":  data,
	}

	this.ProcessSuccessWithResponseBody(w, resp, 200)
}

func (this *databaseController) BackupData() {
	//Delete file if it exists
	//Create file and write out fake database info
	file, err := os.Create("data.txt")
	if err != nil {
		panic(fmt.Sprintf("Could not Create file! err: %s", err))
	}

	b, err := json.Marshal(this)
	if err != nil {
		panic(fmt.Sprintf("Could not marshall data! err: %s", err))
	}

	_, err = file.Write(b)
	if err != nil {
		panic(fmt.Sprintf("Could not write to file! err: %s", err))
	}
}
