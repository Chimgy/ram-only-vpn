package main

import (
	"context"
	"controller/db"
	"controller/node"
	"controller/sessions"
	"encoding/json"
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

func handleConnect(w http.ResponseWriter, r *http.Request,
	conn *pgx.Conn, store *sessions.SessionStore,
	n *node.Node, maxConcurrent int) {

	var req addPeerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate user first
	valid, err := db.ValidateUser(conn, req.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	if !valid {
		writeError(w, http.StatusUnauthorized, "invalid or expired user. Please resubscribe to gain access or submit a ticket.")
	}

	// Check concurrent limit
	if store.Count(req.UserID) >= maxConcurrent {
		writeError(w, http.StatusTooManyRequests, "connection limit reached")
		return
	}

	peer, err := n.AddPeer(req.PubKey, req.UserID)
	if err != nil {
		writeError(w, http.StatusBadGateway, "node error")
		return
	}

	store.Add(req.UserID, req.PubKey, maxConcurrent)

	// return connection config info for user to connect to
	writeJSON(w, http.StatusOK, addPeerResponse{
		TunnelIP:       peer.TunnelIP,
		ServerPubkey:   peer.ServerPubkey,
		ServerEndpoint: peer.ServerEndpoint,
	})
}

//  ACTUALLY DON't NEED THIS RIGHT NOW -- will keep it for now, cos it was a bitch to write -- just in case
// func handleDisconnect(w http.ResponseWriter, r *http.Request,
// 	store *sessions.SessionStore, n *node.Node) {
// 		var req addPeerRequest

// 		// They tried disconnecting but sent an unreadable json packet
// 		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 			writeError(w, http.StatusBadRequest, "Invalid JSON")
// 			return
// 		}

// 		// Check if the person disconnecting is actually connected first
// 		// valid, err := db.ValidateUser(conn, req.UserID) they will not be connected if user doesnt exist redundant check

// 		if store.Count(req.UserID) == 0 {
// 			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
// 			return
// 		}

// 		err := n.RemovePeer(req.PubKey)
// 		if err != nil {
// 			log.Printf("RemovePeer Failed: %v", err)
// 			// Don't return need to remove the session store still
// 		}

// 		// after removing peer THEN we remove from concurrent connections (race conditions idk, but this order gives ME the benefit of the doubt)
// 		store.Remove(req.UserID, req.PubKey)
// 		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
// 	}

// don't need access to node anymore
func handlePeerDisconnected(w http.ResponseWriter, r *http.Request, store *sessions.SessionStore) {
	// jsons from the n-api on disconnection will be handled in this function
	var req addPeerRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid Json sent from node")
		return
	}
	// only sends pubKey NO USER_ID CANNOT COUNT
	store.Remove(req.PubKey)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
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
		if parsed, err := strconv.Atoi(val); err == nil {
			maxConcurrent = parsed
		}
	}

	http.HandleFunc("/connect", func(w http.ResponseWriter, r *http.Request) {
		handleConnect(w, r, conn, store, n, maxConcurrent)
	})
	// http.HandleFunc("/disconnect", func(w http.ResponseWriter, r *http.Request) {
	// 	handleDisconnect(w, r, store, n)
	// })
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
