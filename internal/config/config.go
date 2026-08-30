package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	Watcher  WatcherConfig  `mapstructure:"watcher"`
	Database DatabaseConfig `mapstructure:"database"`
	Parser   ParserConfig   `mapstructure:"parser"`
}
type ServerConfig struct {
	Port    int    `mapstructure:"port"`
	Address string `mapstructure:"address"`
	Debug   bool   `mapstructure:"debug"`
}
type LoggerConfig struct {
	Level          string `mapstructure:"level"`
	Target         string `mapstructure:"target"`
	Filename       string `mapstructure:"filename"`
	MaxFileSizeMB  int    `mapstructure:"maxfilesizemb"`
	MaxFileBackups int    `mapstructure:"maxfilebackups"`
	MaxFileAgeDays int    `mapstructure:"maxfileagedays"`
	FileCompress   bool   `mapstructure:"filecompress"`
}
type WatcherConfig struct {
	Wait        time.Duration `mapstructure:"wait"`
	JournalWait time.Duration `mapstructure:"journalwait"`
}
type DatabaseConfig struct {
	Host     string        `mapstructure:"host"`
	Port     int           `mapstructure:"port"`
	Username string        `mapstructure:"username"`
	Password string        `mapstructure:"password"`
	Database string        `mapstructure:"database"`
	Timeout  time.Duration `mapstructure:"timeout"`
}
type ParserConfig struct {
	Timeout     time.Duration     `mapstructure:"timeout"`
	BaseURI     string            `mapstructure:"baseuri"`
	URIsuffix   string            `mapstructure:"urisuffix"`
	URIprefixes map[string]string `mapstructure:"uriprefixes"`
	Bells       BellsConfig       `mapstructure:"bells"`
}
type BellsConfig struct {
	Monday  map[int]string `mapstructure:"monday"`
	General map[int]string `mapstructure:"general"`
}

func Load() (*Config, error) {
	viper.SetConfigName("default")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Config file on path './configs/default.yaml' not found, writing default config")
			viper.Set("server", DEFAULT_CONFIG.Server)
			viper.Set("database", DEFAULT_CONFIG.Database)
			viper.Set("logger", DEFAULT_CONFIG.Logger)
			viper.Set("parser", DEFAULT_CONFIG.Parser)
			viper.Set("watcher", DEFAULT_CONFIG.Watcher)
			viper.SafeWriteConfig()
		} else {
			return nil, fmt.Errorf("Error reading config file: %v", err)
		}
	}

	var conf Config
	if err := viper.Unmarshal(&conf); err != nil {
		return nil, fmt.Errorf("Error unmarshaling config: %v", err)
	}
	return &conf, nil
}
