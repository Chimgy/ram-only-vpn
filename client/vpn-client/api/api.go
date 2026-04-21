package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Change to controller ip now instead of node
const baseURL = "https://api.ramonlyvpn.net"

type PeerResponse struct {
	TunnelIP       string `json:"tunnel_ip"`
	ServerPubkey   string `json:"server_pubkey"`
	ServerEndpoint string `json:"server_endpoint"`
}

func Connect(publicKey, userID string) (PeerResponse, error) {
	body, _ := json.Marshal(map[string]string{
		"public_key": publicKey,
		"user_id":    userID,
	})
	resp, err := http.Post(baseURL+"/connect", "application/json", bytes.NewReader(body))
	if err != nil {
		return PeerResponse{}, fmt.Errorf("POST /peer: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return PeerResponse{}, fmt.Errorf("server returned %d", resp.StatusCode)
	}
	var pr PeerResponse
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return PeerResponse{}, fmt.Errorf("decoding response: %w", err)
	}
	return pr, nil
}

func Disconnect(publicKey string) error {
	body, _ := json.Marshal(map[string]string{"public_key": publicKey})
	req, _ := http.NewRequest(http.MethodPost, baseURL+"/disconnect", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("POST /disconnect: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %d", resp.StatusCode)
	}
	return nil
}
