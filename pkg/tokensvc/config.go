package tokensvc

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/OdysseyMomentumExperience/token-service/pkg/cache"
	"github.com/OdysseyMomentumExperience/token-service/pkg/sentry"
	"go.uber.org/zap/zapcore"

	"github.com/OdysseyMomentumExperience/token-service/pkg/mqtt"
	"github.com/OdysseyMomentumExperience/token-service/pkg/networks"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Cache          *cache.Config    `yaml:"cache"`
	MQTT           *mqtt.Config     `yaml:"mqtt"`
	NetworkManager *networks.Config `yaml:"network_manager"`
	LogLevel       *LogLevel        `yaml:"log_level" envconfig:"LOG_LEVEL"`
	Sentry         sentry.Config    `yaml:"sentry"`
}

type LogLevel struct {
	zapcore.Level
}

func (x *Config) Init() {
	*x = Config{
		Cache:          &cache.Config{},
		MQTT:           &mqtt.Config{},
		NetworkManager: &networks.Config{},
	}
	x.MQTT.Init()
}

func (cfg *Config) fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (cfg *Config) readFile(path string) error {
	if !cfg.fileExists(path) {
		return errors.New("config path does not exits")
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		if err != io.EOF {
			return err
		}
	}
	return nil
}

func (cfg *Config) PrettyPrint() {
	d, _ := yaml.Marshal(cfg)
	fmt.Printf("--- Config ---\n%s\n\n", string(d))
}

func LoadConfig(path string) (*Config, error) {
	cfg := new(Config)

	cfg.Init()
	err := cfg.readFile(path)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
