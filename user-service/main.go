package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aerospike/aerostation/pkg/auth"
	"github.com/gorilla/mux"
)

func main() {
	//initialize cognito client
	cognitoClient, err := auth.Init()
	if err != nil {
		panic(err.Error())
	}
	r := Router(cognitoClient)
	fmt.Println("Starting user-service service server on the port 8081 ...")
	log.Fatal(http.ListenAndServe(":8081", r))
}

func Router(client *auth.CognitoClient) *mux.Router {
	router := mux.NewRouter()
	cognitoRouter := auth.NewCognitoRouter(*client)

	router.HandleFunc("/api/v1/user/signup", cognitoRouter.SignUp).Methods("POST")
	router.HandleFunc("/api/v1/user/signin", cognitoRouter.SignIn).Methods("POST")
	router.Use(auth.LoggingMiddleware)
	return router
}
