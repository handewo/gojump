package config

import (
	"github.com/handewo/gojump/pkg/log"
	"github.com/spf13/viper"
)

type Config struct {
	BindHost    string `mapstructure:"BIND_HOST"`
	SSHPort     string `mapstructure:"SSHD_PORT"`
	SSHTimeout  int    `mapstructure:"SSH_TIMEOUT"`
	LogFile     string `mapstructure:"LOG_FILE"`
	OtpDuration int64  `mapstructure:"OTP_DURATION"`

	LogLevel         string `mapstructure:"LOG_LEVEL"`
	DbFile           string `mapstructure:"DB_FILE"`
	ReplayFolderPath string ` mapstructure:"REPLAY_PATH"`

	ClientAliveInterval int  `mapstructure:"CLIENT_ALIVE_INTERVAL"`
	RetryAliveCountMax  int  `mapstructure:"RETRY_ALIVE_COUNT_MAX"`
	ReuseConnection     bool `mapstructure:"REUSE_CONNECTION"`
	DisableRecorder     bool `mapstructure:"DISABLE_RECORDER"`
}

var GlobalConfig *Config

func GetConf() Config {
	if GlobalConfig == nil {
		return *newDefaultConfig()
	}
	return *GlobalConfig
}

func Initial(cfgFile string) {
	GlobalConfig = newDefaultConfig()

	viper.SetConfigFile(cfgFile)
	if err := viper.ReadInConfig(); err != nil {
		log.Warning.Printf("Read config failed: %s", err)
		return
	}
	if err := viper.Unmarshal(GlobalConfig); err != nil {
		if err != nil {
			log.Warning.Printf("Load config failed: %s", err)
		}
	}
}

func newDefaultConfig() *Config {
	return &Config{
		BindHost:         "127.0.0.1",
		SSHPort:          "22222",
		SSHTimeout:       30,
		LogLevel:         "INFO",
		LogFile:          "gojump.log",
		DbFile:           "gojumpdb",
		ReplayFolderPath: "gojumpreplay",
		OtpDuration:      120,
	}
}
