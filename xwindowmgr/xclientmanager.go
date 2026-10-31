package xwindowmgr

import (
	"errors"
	"sync"
)

var (
	ErrClientNotFound = errors.New("client not found")
)

type ClientManager struct {
	// for thread safety
	mu sync.RWMutex

	// two maps for fast access
	byClientWindow map[WindowID]*Client
	byFrameWindow  map[WindowID]*Client

	mouseMovedWindow *Client
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		byClientWindow: make(map[WindowID]*Client),
		byFrameWindow:  make(map[WindowID]*Client),
	}
}

// Registry window and it's frame
func (cm *ClientManager) Add(client *Client) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.byClientWindow[client.Window] = client
	cm.byFrameWindow[client.Frame] = client
}

// FindByClient find client object by window handle
func (cm *ClientManager) FindByClient(win WindowID) (*Client, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	client, exists := cm.byClientWindow[win]
	if !exists {
		return nil, ErrClientNotFound
	}
	return client, nil
}

// FindByFrame return client object by frame
func (cm *ClientManager) FindByFrame(frame WindowID) (*Client, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	client, exists := cm.byFrameWindow[frame]
	if !exists {
		return nil, ErrClientNotFound
	}
	return client, nil
}

// Remove client object by window handle
func (cm *ClientManager) Remove(clientWin WindowID) (*Client, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	client, exists := cm.byClientWindow[clientWin]
	if !exists {
		return nil, ErrClientNotFound
	}

	// remove from both maps
	delete(cm.byClientWindow, client.Window)
	delete(cm.byFrameWindow, client.Frame)

	return client, nil
}

// GetAllClients return all windos as slice (for future use)
func (cm *ClientManager) GetAllClients() []*Client {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	clients := make([]*Client, 0, len(cm.byClientWindow))
	for _, client := range cm.byClientWindow {
		clients = append(clients, client)
	}
	return clients
}

func (cm *ClientManager) SetMouseMovedClient(client *Client) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.mouseMovedWindow = client
	return
}
func (cm *ClientManager) GetMouseMovedClient() *Client {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.mouseMovedWindow
}
