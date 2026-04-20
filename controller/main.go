package main

import (
	"context"
	"controller/db"
	"controller/node"
	"controller/sessions"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5"
)

type addPeerRequest struct {
	PubKey string `json:"public_key"`
	UserID string `json:"user_id"`
}

type addPeerResponse struct {
	TunnelIP       string `json:"tunnel_ip"`
	ServerPubkey   string `json:"server_pubkey"`
	ServerEndpoint string `json:"server_endpoint"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

func (conn *pgx.Conn, store *sessionsorsomething) handleConnect(userID, pubKey string) (bool, error) { // dop i unwrap the userID prior to this or is it still in json etc
	// needs to post to node the pubkey + userID after validating the connection with validateuser
	if !db.validateUser(conn, userID) { // the conn object is the connection? do we use db. or conn.?
		return false, nil, fmt.Errorf("wtf have u done %w", err)
	}
	// .add checks this as well, but i guess good to do before commiting maybe also good to do both
	if store.count() >= maxConcurrent {
		return false, nil, fmt.ErrorF("too many concurrent connections ")
	}

	// now can forward to node if passed those checks
	n.addPeer(pubKey, userID)

	store.add()
}

func main() {
	conn, err := db.Connect() // a pgx pointer? a context? whats a context?
	if err != nil {
		log.Fatalf("connecting to db: %v", err) // what is %w again?
	}
	defer conn.Close(context.Background())

	store := sessions.NewSessionStore()

	n := node.NewNode()

	maxConcurrent := 2 // default
	if val := os.Getenv("MAX_CONCURRENT"); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			maxConcurrent = n
		}
	}

	http.HandleFunc("/connect", func(w http.ResponseWriter, r *http.Request) {
		handleConnect(w, r, conn, store, n, maxConcurrent)
	})
	http.HandleFunc("/disconnect", func(w http.ResponseWriter, r *http.Request) {
		handleDisconnect(w, r, store, n)
	})
	http.HandleFunc("/peer/disconnected", func(w http.ResponseWriter, r *http.Request) {
		handlePeerDisconnected(w, r, store)
	})

	port := os.Getenv("CONTROLLER_PORT")
	if port == "" {
		port = "9090"
	}
	log.Printf("controller listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
