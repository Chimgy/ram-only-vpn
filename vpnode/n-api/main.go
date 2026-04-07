package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"vpnode-api/peerpool"
	"vpnode-api/wg"
)

// ----------------------------------------------------------------
// Controller seam
//
// Right now: handler logic lives directly in handleAddPeer.
// When the physical controller is ready, replace this file's
// handler with one that proxies to the controller instead.
// peerpool and wg packages are untouched.
// ----------------------------------------------------------------

var pool *peerpool.Pool

// Request/response types

type addPeerRequest struct {
	PublicKey string `json:"public_key"`
	UserID    string `json:"user_id"` // accepted but not validated yet, the controller will own this
}

type addPeerResponse struct {
	TunnelIP    string `json:"tunnel_ip"`
	ServerPubkey string `json:"server_pubkey"`
	ServerEndpoint string `json:"server_endpoint"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// Helpers

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

// POST /peer
// Body: { "public_key": "...", "user_id": "..." }
// Returns: { "tunnel_ip": "10.8.0.x", "server_pubkey": "...", "server_endpoint": "x.x.x.x:51820" }
func handleAddPeer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "POST only")
		return
	}

	var req addPeerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	req.PublicKey = strings.TrimSpace(req.PublicKey)
	if req.PublicKey == "" {
		writeError(w, http.StatusBadRequest, "public_key required")
		return
	}

	// Assign tunnel IP (idempotent — same pubkey gets same IP)
	tunnelIP, err := pool.Assign(req.PublicKey)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}

	// Add peer to live WireGuard interface
	if err := wg.AddPeer(req.PublicKey, tunnelIP.String()); err != nil {
		// Roll back pool assignment so IP isn't leaked
		pool.Release(req.PublicKey)
		log.Printf("wg AddPeer failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to configure WireGuard peer")
		return
	}

	serverPubkey, err := wg.ServerPubkey()
	if err != nil {
		log.Printf("WARNING: could not read server pubkey: %v", err)
		serverPubkey = "unavailable"
	}

	// Best-effort: get LAN IP for endpoint hint
	lanIP := os.Getenv("VPN_LAN_IP") // set this in vpn-boot.sh: export VPN_LAN_IP=$MY_IP
	if lanIP == "" {
		lanIP = "192.168.1.108" // fallback
	}

	log.Printf("Peer added: user=%s pubkey=%s tunnel=%s", req.UserID, req.PublicKey[:8]+"...", tunnelIP)

	writeJSON(w, http.StatusOK, addPeerResponse{
		TunnelIP:       tunnelIP.String(),
		ServerPubkey:   serverPubkey,
		ServerEndpoint: fmt.Sprintf("%s:51820", lanIP),
	})
}

// While the web application is used for parsing webkeys will need this to read what
// browser sends
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "POST, GET, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusOK)
            return
        }
        next(w, r)
    }
}


// DELETE /peer
// Body: { "public_key": "..." }
func handleRemovePeer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "DELETE only")
		return
	}

	var req struct {
		PublicKey string `json:"public_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if err := wg.RemovePeer(req.PublicKey); err != nil {
		log.Printf("wg RemovePeer failed: %v", err)
	}

	pool.Release(req.PublicKey)
	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

// GET /peers — debug endpoint, shows active peers + pool status
func handleListPeers(w http.ResponseWriter, r *http.Request) {
	peers := pool.List()
	writeJSON(w, http.StatusOK, map[string]any{
		"active":    peers,
		"available": pool.Available(),
	})
}

// GET /health
func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func main() {
	var err error
	// Pool: 10.8.0.2 — 10.8.0.50 (48 concurrent peers, expand as needed)
	pool, err = peerpool.New(2, 50)
	if err != nil {
		log.Fatalf("Failed to init peer pool: %v", err)
	}

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

http.HandleFunc("/peer", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {		
	switch r.Method {
		case http.MethodPost:
			handleAddPeer(w, r)
		case http.MethodDelete:
			handleRemovePeer(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "POST or DELETE only")
		}
	}))
	http.HandleFunc("/peers", corsMiddleware(handleListPeers))
	http.HandleFunc("/health", corsMiddleware(handleHealth))

	log.Printf("vpnode-api listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}