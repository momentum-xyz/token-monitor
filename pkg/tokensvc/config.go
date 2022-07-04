package tokensvc

import (
	"fmt"
	"github.com/momentum-xyz/token-monitor/pkg/cache"
	"io"
	"os"

	"github.com/momentum-xyz/token-monitor/pkg/log"
	"github.com/momentum-xyz/token-monitor/pkg/mqtt"
	"github.com/momentum-xyz/token-monitor/pkg/networks"
	"gopkg.in/yaml.v2"
)

const configFileName = "config.dev.yaml"

type Config struct {
	Cache          *cache.Config    `yaml:"cache"`
	MQTT           *mqtt.Config     `yaml:"mqtt"`
	NetworkManager *networks.Config `yaml:"network_manager"`
	Log            *log.Config      `yaml:"log"`
}

func (x *Config) Init() {
	*x = Config{
		Cache:          &cache.Config{},
		MQTT:           &mqtt.Config{},
		NetworkManager: &networks.Config{},
	}
	x.MQTT.Init()
}

func (cfg *Config) processError(err error) {
	fmt.Println(err)
	os.Exit(2)
}

func (cfg *Config) fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (cfg *Config) readFile(path string) {
	if !cfg.fileExists(path) {
		return
	}
	f, err := os.Open(path)
	if err != nil {
		cfg.processError(err)
	}
	defer f.Close()
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		if err != io.EOF {
			cfg.processError(err)
		}
	}
}

func (cfg *Config) PrettyPrint() {
	d, _ := yaml.Marshal(cfg)
	log.Logf(1, "--- Config ---\n%s\n\n", string(d))
}

func LoadConfig(path string) *Config {
	cfg := new(Config)

	cfg.Init()
	cfg.readFile(path)
	log.SetLogLevel(1)

	return cfg
}

func GetDefaultConfig() *Config {
	return LoadConfig(configFileName)
}
