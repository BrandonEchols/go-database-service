package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type controller struct {
	//Placeholder struct to re-use common functionality
}

func (this *controller) WriteErrorMessageWithStatus(w http.ResponseWriter, msg string, code int) {
	response, _ := json.Marshal(map[string]interface{}{
		"error": msg,
	})

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err := w.Write(response)
	if err != nil {
		fmt.Println("Error writing response body: ", err.Error())
	}
}

func (this *controller) ProcessSuccessWithResponseBody(w http.ResponseWriter, resp interface{}, code int) {
	response, err := json.Marshal(resp)
	if err != nil {
		fmt.Println("Error marshalling response body: ", err.Error())
		w.WriteHeader(500)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(response)
	if err != nil {
		fmt.Println("Error writing response body: ", err.Error())
	}
}
