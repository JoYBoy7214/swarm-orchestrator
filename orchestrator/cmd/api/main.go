package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	pub "github.com/JoYBoy7214/swarm-orchestrator/internal/pub"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type Payload struct {
	Url string `json:"url"`
}
type UserRequest struct {
	TaskType string  `json:"task_type"`
	Payload  Payload `json:"payload"`
}
type Event struct {
	EventID   uuid.UUID   `json:"event_id"`
	UserReq   UserRequest `json:"user_req"`
	Status    string      `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
}
type Response struct {
	EventID string `json:"task_id"`
}

type NServer struct {
	Jstream *pub.Publisher
}

func (ns *NServer) TaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var req UserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	//log.Println(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	e := UserReqToEvent(req)
	b, err := json.Marshal(e)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err = ns.Jstream.Publish(b, ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res := Response{
		EventID: e.EventID.String(),
	}
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
func UserReqToEvent(req UserRequest) Event {
	id := uuid.New()
	t := time.Now()
	return Event{
		EventID:   id,
		UserReq:   req,
		Status:    "Pending",
		CreatedAt: t,
	}
}

func main() {
	var err error
	var nc *nats.Conn
	var server NServer
	server.Jstream, nc, err = pub.CreateStream()
	defer nc.Close()
	if err != nil {
		log.Fatal("error in creating JetStream", err)
		return
	}

	http.HandleFunc("POST /api/v1/tasks", server.TaskHandler)

	log.Println("Server is Started at 8080")
	if http.ListenAndServe(":8080", nil) != nil {
		log.Fatal("failed to start the server at 8080")
		return
	}

}
