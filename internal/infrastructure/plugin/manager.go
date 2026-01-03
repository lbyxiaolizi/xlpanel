package plugin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

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

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  ProtocolVersion,
	MagicCookieKey:   MagicCookieKey,
	MagicCookieValue: MagicCookieVal,
}

type PluginManager struct {
	pluginsDir string
	handshake  plugin.HandshakeConfig
	logger     hclog.Logger
	mu         sync.RWMutex
	loaded     map[string]*LoadedPlugin
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
		pluginsDir: pluginsDir,
		handshake:  Handshake,
		logger:     logger,
		loaded:     make(map[string]*LoadedPlugin),
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

func (m *PluginManager) LoadPlugin(path string) (*LoadedPlugin, error) {
	return LoadPlugin(path, m.handshake, m.logger)
}

func (m *PluginManager) GetClient(moduleName string) (*grpc.ClientConn, error) {
	if moduleName == "" {
		return nil, errors.New("module name is required")
	}
	m.mu.RLock()
	loaded := m.loaded[moduleName]
	m.mu.RUnlock()
	if loaded != nil {
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
