package exchanges

import (
	"fmt"
	"sync"
)

// Registry holds all available exchange testers
type Registry struct {
	testers map[string]ExchangeTester
	mutex   sync.RWMutex
}

// NewRegistry creates a new exchange registry
func NewRegistry() *Registry {
	registry := &Registry{
		testers: make(map[string]ExchangeTester),
	}

	// Register default testers
	registry.Register("binance", NewBinanceTester())
	registry.Register("coinbase", NewCoinbaseTester())

	return registry
}

// Register adds a new exchange tester to the registry
func (r *Registry) Register(name string, tester ExchangeTester) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.testers[name] = tester
}

// Get retrieves an exchange tester by name
func (r *Registry) Get(name string) (ExchangeTester, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tester, exists := r.testers[name]
	if !exists {
		return nil, fmt.Errorf("exchange tester '%s' not found", name)
	}

	return tester, nil
}

// List returns all available exchange names
func (r *Registry) List() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	names := make([]string, 0, len(r.testers))
	for name := range r.testers {
		names = append(names, name)
	}

	return names
}

// GetTesterForExchange returns the appropriate tester for the given exchange
func GetTesterForExchange(exchangeName string) (ExchangeTester, error) {
	registry := NewRegistry()
	return registry.Get(exchangeName)
}
