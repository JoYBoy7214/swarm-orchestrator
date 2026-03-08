package main

import (
	"log"
	"net/http"
)
type Payload struct{
	url string 
}
type UserRequest struct{
	TaskType string `json:"task_type"`
	Payload Payload `json:"pay_load"`

}
type Event struct{
	
}
func TaskHandler(w http.ResponseWriter, r *http.Request) {

}
func main() {
	http.HandleFunc("/api/v1/tasks", TaskHandler)

	log.Println("Server is Started at 8080")
	if http.ListenAndServe(":8080", nil) != nil {
		log.Fatal("failed to start the server at 8080")
	}

}
