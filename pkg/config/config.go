package config

import (
	"github.com/handewo/gojump/pkg/log"
	"github.com/spf13/viper"
)

type Config struct {
	BindHost    string `mapstructure:"BIND_HOST" json:"BIND_HOST"`
	SSHPort     string `mapstructure:"SSHD_PORT" json:"SSHD_PORT"`
	SSHTimeout  int    `mapstructure:"SSH_TIMEOUT" json:"SSH_TIMEOUT"`
	LogFile     string `mapstructure:"LOG_FILE" json:"LOG_FILE"`
	OtpDuration int64  `mapstructure:"OTP_DURATION" json:"OTP_DURATION"`

	LogLevel         string `mapstructure:"LOG_LEVEL" json:"LOG_LEVEL"`
	DbPath           string `mapstructure:"DB_PATH" json:"DB_PATH"`
	ReplayFolderPath string ` mapstructure:"REPLAY_PATH" json:"REPLAY_PATH"`

	ClientAliveInterval int  `mapstructure:"CLIENT_ALIVE_INTERVAL" json:"CLIENT_ALIVE_INTERVAL"`
	RetryAliveCountMax  int  `mapstructure:"RETRY_ALIVE_COUNT_MAX" json:"RETRY_ALIVE_COUNT_MAX"`
	ReuseConnection     bool `mapstructure:"REUSE_CONNECTION" json:"REUSE_CONNECTION"`
	DisableRecorder     bool `mapstructure:"DISABLE_RECORDER" json:"DISABLE_RECORDER"`
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
		DbPath:           "gojumpdb",
		ReplayFolderPath: "gojumpreplay",
		OtpDuration:      120,
	}
}
