package controllers

import (
	"encoding/json"
	"net/http"
)

type (
	IDatabaseController interface {
		SetData(w http.ResponseWriter, r *http.Request)
		GetData(w http.ResponseWriter, r *http.Request)
		DeleteData(w http.ResponseWriter, r *http.Request)
	}

	databaseController struct {
		controller
		fakeDatabase map[string]string
	}
)

func GetDatabaseController() IDatabaseController {
	return &databaseController{
		fakeDatabase: map[string]string{},
	}
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
	Key   string `json:"key"`
	Value string `json:"value"`
}
func (this *databaseController) SetData(w http.ResponseWriter, r *http.Request) {
	payload := dbData{}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil || payload.Key == "" || payload.Value == "" {
		this.WriteErrorMessageWithStatus(w, "Invalid payload", 400)
		return
	}

	this.fakeDatabase[payload.Key] = payload.Value

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
	key := r.URL.Query().Get("key")
	if key == "" {
		this.WriteErrorMessageWithStatus(w, "invalid_key", 400)
		return
	}

	data, ok := this.fakeDatabase[key]
	if !ok {
		this.WriteErrorMessageWithStatus(w, "key_not_found", 404)
		return
	}

	resp := dbData{
		Key: key,
		Value: data,
	}

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
	Key   string `json:"key"`
}
func (this *databaseController) DeleteData(w http.ResponseWriter, r *http.Request) {
	payload := deleteRequest{}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil || payload.Key == "" {
		this.WriteErrorMessageWithStatus(w, "Invalid payload", 400)
		return
	}

	data, ok := this.fakeDatabase[payload.Key]
	if !ok {
		resp := map[string]interface{}{
			"message": "key_not_found",
		}
		this.ProcessSuccessWithResponseBody(w, resp, 200)
		return
	}
	delete(this.fakeDatabase, payload.Key)

	resp := dbData{
		Key: payload.Key,
		Value: data,
	}
	this.ProcessSuccessWithResponseBody(w, resp, 200)
}
