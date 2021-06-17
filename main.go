package main

import (
	"fmt"
	"github.com/BrandonEchols/go-database-service/controllers"
	"github.com/gorilla/mux"
	"net/http"
)

func main() {
	dbController := controllers.GetDatabaseController()

	r := mux.NewRouter()
	r.HandleFunc("/set", dbController.SetData).Methods("POST")
	r.HandleFunc("/get", dbController.GetData).Methods("GET")
	r.HandleFunc("/delete", dbController.DeleteData).Methods("POST") //Noted in requirement this should be a POST not a DELETE

	addr := "127.0.0.1:4000"
	fmt.Println("Starting service on", addr)
	err := http.ListenAndServe(addr, r)
	if err != nil {
		panic(err)
	}
}
