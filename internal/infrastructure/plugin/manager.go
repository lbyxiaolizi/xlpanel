package plugin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

const (
	ProtocolVersion = 1
	MagicCookieKey  = "OPENHOST_PLUGIN"
	MagicCookieVal  = "openhost"
	checksumSuffix  = ".sha256"
)

// PluginStatus represents the status of a plugin
type PluginStatus string

const (
	PluginStatusActive   PluginStatus = "active"
	PluginStatusInactive PluginStatus = "inactive"
	PluginStatusError    PluginStatus = "error"
	PluginStatusLoading  PluginStatus = "loading"
)

// PluginInfo contains metadata about a plugin
type PluginInfo struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description"`
	Author       string            `json:"author"`
	AuthorURL    string            `json:"author_url"`
	Type         string            `json:"type"`
	Capabilities []string          `json:"capabilities"`
	ConfigSchema map[string]any    `json:"config_schema,omitempty"`
	Config       map[string]string `json:"config,omitempty"`
}

// PluginState tracks the runtime state of a plugin
type PluginState struct {
	Info       *PluginInfo  `json:"info"`
	Status     PluginStatus `json:"status"`
	Path       string       `json:"path"`
	LoadedAt   *time.Time   `json:"loaded_at,omitempty"`
	LastError  string       `json:"last_error,omitempty"`
	LastUsed   *time.Time   `json:"last_used,omitempty"`
	UseCount   int64        `json:"use_count"`
}

// PluginEvent represents events that occur during plugin lifecycle
type PluginEvent struct {
	Type      string         `json:"type"`
	Plugin    string         `json:"plugin"`
	Timestamp time.Time      `json:"timestamp"`
	Data      map[string]any `json:"data,omitempty"`
	Error     string         `json:"error,omitempty"`
}

// EventType constants
const (
	EventPluginLoaded      = "plugin_loaded"
	EventPluginUnloaded    = "plugin_unloaded"
	EventPluginError       = "plugin_error"
	EventPluginConfigured  = "plugin_configured"
	EventProvisionStart    = "provision_start"
	EventProvisionComplete = "provision_complete"
	EventProvisionError    = "provision_error"
)

// EventHandler is called when plugin events occur
type EventHandler func(event PluginEvent)

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  ProtocolVersion,
	MagicCookieKey:   MagicCookieKey,
	MagicCookieValue: MagicCookieVal,
}

type PluginManager struct {
	pluginsDir    string
	handshake     plugin.HandshakeConfig
	logger        hclog.Logger
	mu            sync.RWMutex
	loaded        map[string]*LoadedPlugin
	states        map[string]*PluginState
	configs       map[string]map[string]string
	eventHandlers []EventHandler
}

func NewPluginManager(pluginsDir string, logger hclog.Logger) *PluginManager {
	if pluginsDir == "" {
		pluginsDir = "./plugins"
	}
	if logger == nil {
		logger = hclog.New(&hclog.LoggerOptions{
			Name:  "plugin-manager",
			Level: hclog.Info,
		})
	}
	return &PluginManager{
		pluginsDir:    pluginsDir,
		handshake:     Handshake,
		logger:        logger,
		loaded:        make(map[string]*LoadedPlugin),
		states:        make(map[string]*PluginState),
		configs:       make(map[string]map[string]string),
		eventHandlers: make([]EventHandler, 0),
	}
}

// OnEvent registers an event handler
func (m *PluginManager) OnEvent(handler EventHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.eventHandlers = append(m.eventHandlers, handler)
}

// emitEvent sends an event to all registered handlers
func (m *PluginManager) emitEvent(event PluginEvent) {
	m.mu.RLock()
	handlers := make([]EventHandler, len(m.eventHandlers))
	copy(handlers, m.eventHandlers)
	m.mu.RUnlock()

	for _, handler := range handlers {
		go handler(event)
	}
}

func (m *PluginManager) Scan() ([]string, error) {
	entries, err := os.ReadDir(m.pluginsDir)
	if err != nil {
		return nil, fmt.Errorf("read plugins directory: %w", err)
	}

	var plugins []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(m.pluginsDir, entry.Name())
		if !isExecutable(path) {
			continue
		}
		plugins = append(plugins, path)
	}
	return plugins, nil
}

// ScanWithInfo returns detailed information about available plugins
func (m *PluginManager) ScanWithInfo() ([]*PluginState, error) {
	plugins, err := m.Scan()
	if err != nil {
		return nil, err
	}

	var states []*PluginState
	for _, path := range plugins {
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		
		m.mu.RLock()
		state, exists := m.states[name]
		m.mu.RUnlock()

		if !exists {
			state = &PluginState{
				Info: &PluginInfo{
					Name: name,
				},
				Status: PluginStatusInactive,
				Path:   path,
			}
		}
		states = append(states, state)
	}
	return states, nil
}

// GetPluginState returns the state of a specific plugin
func (m *PluginManager) GetPluginState(name string) *PluginState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.states[name]
}

// GetAllStates returns all plugin states
func (m *PluginManager) GetAllStates() map[string]*PluginState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	states := make(map[string]*PluginState, len(m.states))
	for k, v := range m.states {
		states[k] = v
	}
	return states
}

func (m *PluginManager) LoadPlugin(path string) (*LoadedPlugin, error) {
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	
	// Update state to loading
	m.mu.Lock()
	m.states[name] = &PluginState{
		Info:   &PluginInfo{Name: name},
		Status: PluginStatusLoading,
		Path:   path,
	}
	m.mu.Unlock()

	loaded, err := LoadPlugin(path, m.handshake, m.logger)
	if err != nil {
		m.mu.Lock()
		m.states[name].Status = PluginStatusError
		m.states[name].LastError = err.Error()
		m.mu.Unlock()
		
		m.emitEvent(PluginEvent{
			Type:      EventPluginError,
			Plugin:    name,
			Timestamp: time.Now(),
			Error:     err.Error(),
		})
		return nil, err
	}

	now := time.Now()
	m.mu.Lock()
	m.states[name].Status = PluginStatusActive
	m.states[name].LoadedAt = &now
	m.states[name].LastError = ""
	m.mu.Unlock()

	m.emitEvent(PluginEvent{
		Type:      EventPluginLoaded,
		Plugin:    name,
		Timestamp: now,
	})

	return loaded, nil
}

// UnloadPlugin unloads a plugin and releases resources
func (m *PluginManager) UnloadPlugin(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	loaded, exists := m.loaded[name]
	if !exists {
		return fmt.Errorf("plugin not loaded: %s", name)
	}

	// Kill the plugin process
	if loaded.Client != nil {
		loaded.Client.Kill()
	}

	delete(m.loaded, name)
	
	if state, ok := m.states[name]; ok {
		state.Status = PluginStatusInactive
		state.LoadedAt = nil
	}

	m.emitEvent(PluginEvent{
		Type:      EventPluginUnloaded,
		Plugin:    name,
		Timestamp: time.Now(),
	})

	return nil
}

// EnablePlugin enables and loads a plugin
func (m *PluginManager) EnablePlugin(name string) error {
	path, err := m.resolvePluginPath(name)
	if err != nil {
		return err
	}

	loaded, err := m.LoadPlugin(path)
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.loaded[name] = loaded
	m.mu.Unlock()

	return nil
}

// DisablePlugin disables a plugin
func (m *PluginManager) DisablePlugin(name string) error {
	return m.UnloadPlugin(name)
}

// SetPluginConfig sets configuration for a plugin
func (m *PluginManager) SetPluginConfig(name string, config map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.configs[name] = config

	if state, ok := m.states[name]; ok && state.Info != nil {
		state.Info.Config = config
	}

	m.emitEvent(PluginEvent{
		Type:      EventPluginConfigured,
		Plugin:    name,
		Timestamp: time.Now(),
		Data:      map[string]any{"config": config},
	})

	return nil
}

// GetPluginConfig returns the configuration for a plugin
func (m *PluginManager) GetPluginConfig(name string) map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if config, ok := m.configs[name]; ok {
		return config
	}
	return nil
}

// SaveConfigs saves plugin configurations to a file
func (m *PluginManager) SaveConfigs(path string) error {
	m.mu.RLock()
	data, err := json.MarshalIndent(m.configs, "", "  ")
	m.mu.RUnlock()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadConfigs loads plugin configurations from a file
func (m *PluginManager) LoadConfigs(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var configs map[string]map[string]string
	if err := json.Unmarshal(data, &configs); err != nil {
		return err
	}

	m.mu.Lock()
	m.configs = configs
	m.mu.Unlock()
	return nil
}

func (m *PluginManager) GetClient(moduleName string) (*grpc.ClientConn, error) {
	if moduleName == "" {
		return nil, errors.New("module name is required")
	}
	m.mu.RLock()
	loaded := m.loaded[moduleName]
	m.mu.RUnlock()
	if loaded != nil {
		// Update usage stats
		now := time.Now()
		m.mu.Lock()
		if state, ok := m.states[moduleName]; ok {
			state.LastUsed = &now
			state.UseCount++
		}
		m.mu.Unlock()
		return loaded.Conn, nil
	}

	path, err := m.resolvePluginPath(moduleName)
	if err != nil {
		return nil, err
	}

	loaded, err = m.LoadPlugin(path)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.loaded[moduleName] = loaded
	m.mu.Unlock()
	return loaded.Conn, nil
}

type LoadedPlugin struct {
	Client *plugin.Client
	Conn   *grpc.ClientConn
}

func LoadPlugin(path string, handshake plugin.HandshakeConfig, logger hclog.Logger) (*LoadedPlugin, error) {
	if logger == nil {
		logger = hclog.New(&hclog.LoggerOptions{
			Name:  "plugin-loader",
			Level: hclog.Info,
		})
	}
	if err := verifyChecksum(path); err != nil {
		return nil, err
	}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: handshake,
		Plugins: map[string]plugin.Plugin{
			"grpc": GRPCConnector{},
		},
		Cmd:              exec.Command(path),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           logger,
	})

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("create rpc client: %w", err)
	}

	dispensed, err := rpcClient.Dispense("grpc")
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("dispense grpc client: %w", err)
	}

	conn, ok := dispensed.(*grpc.ClientConn)
	if !ok {
		client.Kill()
		return nil, fmt.Errorf("unexpected grpc client type: %T", dispensed)
	}

	return &LoadedPlugin{Client: client, Conn: conn}, nil
}

type GRPCConnector struct {
	plugin.Plugin
}

func (GRPCConnector) GRPCServer(_ *plugin.GRPCBroker, _ *grpc.Server) error {
	return errors.New("grpc server not supported on host")
}

func (GRPCConnector) GRPCClient(_ context.Context, _ *plugin.GRPCBroker, conn *grpc.ClientConn) (interface{}, error) {
	return conn, nil
}

func verifyChecksum(path string) error {
	sumPath := path + checksumSuffix
	data, err := os.ReadFile(sumPath)
	if err != nil {
		return fmt.Errorf("read checksum file: %w", err)
	}

	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return fmt.Errorf("checksum file is empty: %s", sumPath)
	}
	expected := strings.ToLower(fields[0])

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open plugin binary: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return fmt.Errorf("hash plugin binary: %w", err)
	}
	actual := hex.EncodeToString(hasher.Sum(nil))
	if actual != expected {
		return fmt.Errorf("checksum mismatch for %s", path)
	}
	return nil
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	mode := info.Mode()
	return mode.IsRegular() && mode&0o111 != 0
}

func (m *PluginManager) resolvePluginPath(moduleName string) (string, error) {
	direct := filepath.Join(m.pluginsDir, moduleName)
	if isExecutable(direct) {
		return direct, nil
	}
	plugins, err := m.Scan()
	if err != nil {
		return "", err
	}
	for _, candidate := range plugins {
		if strings.TrimSuffix(filepath.Base(candidate), filepath.Ext(candidate)) == moduleName {
			return candidate, nil
		}
		if filepath.Base(candidate) == moduleName {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("plugin not found: %s", moduleName)
}
