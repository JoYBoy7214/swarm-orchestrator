package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/JoYBoy7214/swarm-orchestrator/internal/orchestrator"
	"github.com/nats-io/nats.go"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	orch, err := orchestrator.CreateOrchestrator(ctx, "postgres://postgres:postgres123@localhost:5432/postgres?sslmode=disable", nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
		return
	}
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		orch.StartOrchestrating(ctx)
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/tasks", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		rctx, rcancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer rcancel()
		workflow_id, err := orch.DbDriver.CreateWorkflow(rctx)
		if err != nil {
			log.Println("Error in creating workflow %w", err)
			http.Error(w, "Error in creating workflow", http.StatusInternalServerError)
			return
		}

		err = orch.RootNodeUpdater(rctx, workflow_id)
		if err != nil {
			log.Println("Error in updating workflow %w", err)
			http.Error(w, "Error in creating workflow", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		log.Println("Server started at :8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutdowning grace fully")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}
	wg.Wait()
	stop() //this will not wait it will just trigger the signal
	log.Println("Application stopped cleanly.")
}
