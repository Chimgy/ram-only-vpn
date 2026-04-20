package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Node struct {
	url    string
	apiKey string
}

type PeerResponse struct {
	TunnelIP       string `json:"tunnel_ip"`
	ServerPubkey   string `json:"server_pubkey"`
	ServerEndpoint string `json:"server_endpoint"`
}

func (n *Node) AddPeer(pubkey, userID string) (PeerResponse, error) {
	body, _ := json.Marshal(map[string]string{
		"public_key": pubkey,
		"user_id":    userID,
	})

	req, _ := http.NewRequest(http.MethodPost, n.url+"/peer", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", n.apiKey)
	resp, err := http.DefaultClient.Do(req)

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

func (n *Node) RemovePeer(pubkey string) error {

	body, _ := json.Marshal(map[string]string{"public_key": pubkey})
	req, _ := http.NewRequest(http.MethodDelete, n.url+"/peer", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", n.apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("DELETE /peer: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("server returned %d", resp.StatusCode)
	}
	return nil

}

// constructor
func NewNode() *Node {
	return &Node{

		url:    os.Getenv("NODE_URL"),
		apiKey: os.Getenv("NODE_API_KEY"),
	}
}
