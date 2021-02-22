package pkg

import (
	"encoding/json"
	"fmt"
	"os"
)

type BackendConfig struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	MaxConnections int    `json:"maxconnections"`
}

type BackendRouterConfig struct {
	SelectionMethod string            `json:"SelectionMethod"`
	AcceptedPaths   []string          `json:"AcceptedPaths,omitempty"`
	AcceptedHeaders map[string]string `json:"AcceptedHeaders,omitempty"`
	BackendConfigs  []BackendConfig   `json:"BackendConfigs,omitempty"`
}

type Config struct {
	CertCrtPath          string                `json:"certcrtpath"`
	CertKeyPath          string                `json:"certkeypath"`
	Host                 string                `json:"host"`
	Port                 int                   `json:"port"`
	TlsListener          bool                  `json:"tlslistener"`
	BackendRouterConfigs []BackendRouterConfig `json:"BackendRouterConfigs"`
}

// LoadConfig, loads configuation for LBLight. Primarily backend host, port, paths etc.
func LoadConfig(filePath string) Config {
	var config Config
	configFile, err := os.Open(filePath)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}
