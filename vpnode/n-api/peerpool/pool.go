package peerpool

import (
	"fmt"
	"net"
	"sync"
	"math/rand"
)

// Peer represents an active WireGuard peer
type Peer struct {
	PublicKey string
	TunnelIP  net.IP
}

// Pool manages a pool of tunnel IPs and tracks active peers
// All state is in-memory — intentionally lost on reboot
type Pool struct {
	mu       sync.Mutex
	available []net.IP          // IPs not yet assigned
	assigned  map[string]net.IP // pubkey -> tunnel IP
}

// New creates a Pool covering 10.8.0.start through 10.8.0.end (inclusive)
// The server itself holds 10.8.0.1 -> start from 2
func New(start, end int) (*Pool, error) {
	if start < 2 || end > 254 || start > end {
		return nil, fmt.Errorf("invalid range: %d-%d (must be 2-254)", start, end)
	}

	pool := &Pool{
		available: make([]net.IP, 0, end-start+1),
		assigned:  make(map[string]net.IP),
	}

	for i := start; i <= end; i++ {
		pool.available = append(pool.available, net.ParseIP(fmt.Sprintf("10.8.0.%d", i)))
	}

	// after building the slice, shuffle it
	rand.Shuffle(len(pool.available), func(i, j int) {
		pool.available[i], pool.available[j] = pool.available[j], pool.available[i]
	})

	return pool, nil
}

// Assign returns an existing IP if pubkey already registered,
// otherwise pops the next available IP from the pool
func (p *Pool) Assign(pubkey string) (net.IP, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Already assigned — idempotent
	if ip, ok := p.assigned[pubkey]; ok {
		return ip, nil
	}

	if len(p.available) == 0 {
		return nil, fmt.Errorf("tunnel IP pool exhausted")
	}

	// Pop from front
	ip := p.available[0]
	p.available = p.available[1:]
	p.assigned[pubkey] = ip

	return ip, nil
}

// Release removes a peer and returns its IP to the pool
func (p *Pool) Release(pubkey string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	ip, ok := p.assigned[pubkey]
	if !ok {
		return false
	}

	delete(p.assigned, pubkey)
	p.available = append(p.available, ip)
	return true
}

// List returns a snapshot of all active peers
func (p *Pool) List() []Peer {
	p.mu.Lock()
	defer p.mu.Unlock()

	peers := make([]Peer, 0, len(p.assigned))
	for pubkey, ip := range p.assigned {
		peers = append(peers, Peer{PublicKey: pubkey, TunnelIP: ip})
	}
	return peers
}

// Available returns how many IPs remain in the pool
func (p *Pool) Available() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.available)
}