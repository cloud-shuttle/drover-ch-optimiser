package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	ClickHouse struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Database string `mapstructure:"database"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		Secure   bool   `mapstructure:"secure"`
	} `mapstructure:"clickhouse"`
}

var C Config

// Overrides apply non-empty CLI flag values after Load.
type Overrides struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
	Secure   *bool
}

func ApplyOverrides(o Overrides) {
	if o.Host != "" {
		C.ClickHouse.Host = o.Host
	}
	if o.Port > 0 {
		C.ClickHouse.Port = o.Port
	}
	if o.Database != "" {
		C.ClickHouse.Database = o.Database
	}
	if o.Username != "" {
		C.ClickHouse.Username = o.Username
	}
	if o.Password != "" {
		C.ClickHouse.Password = o.Password
	}
	if o.Secure != nil {
		C.ClickHouse.Secure = *o.Secure
	}
}

func Load(cfgFile string) error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		viper.AddConfigPath(filepath.Join(home, ".drover"))
		viper.SetConfigName("ch-optimiser")
	}

	viper.SetDefault("clickhouse.host", "localhost")
	viper.SetDefault("clickhouse.port", 9000)
	viper.SetDefault("clickhouse.database", "default")
	viper.SetDefault("clickhouse.username", "default")
	viper.SetDefault("clickhouse.password", "")
	viper.SetDefault("clickhouse.secure", false)

	viper.SetEnvPrefix("CLICKHOUSE")
	viper.AutomaticEnv()
	_ = viper.BindEnv("clickhouse.host", "CLICKHOUSE_HOST")
	_ = viper.BindEnv("clickhouse.port", "CLICKHOUSE_PORT")
	_ = viper.BindEnv("clickhouse.database", "CLICKHOUSE_DATABASE")
	_ = viper.BindEnv("clickhouse.username", "CLICKHOUSE_USER")
	_ = viper.BindEnv("clickhouse.password", "CLICKHOUSE_PASSWORD")

	if cfgFile != "" {
		if _, err := os.Stat(cfgFile); err == nil {
			if err := viper.ReadInConfig(); err != nil {
				return err
			}
		}
	} else if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}
	return viper.Unmarshal(&C)
}
