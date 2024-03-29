package config

import (
	"github.com/handewo/gojump/pkg/log"
	"github.com/spf13/viper"
)

type Config struct {
	BindHost string `mapstructure:"BIND_HOST" json:"BIND_HOST"`
	SSHPort  string `mapstructure:"SSHD_PORT" json:"SSHD_PORT"`

	//Second
	SSHTimeout int `mapstructure:"SSH_TIMEOUT" json:"SSH_TIMEOUT"`
	//Second
	ClientAliveInterval int `mapstructure:"CLIENT_ALIVE_INTERVAL" json:"CLIENT_ALIVE_INTERVAL"`
	//Minute
	LoginBlockTime int64 `mapstructure:"LOGIN_BLOCK_TIME" json:"LOGIN_BLOCK_TIME"`
	//Second
	OtpDuration int64 `mapstructure:"OTP_DURATION" json:"OTP_DURATION"`

	LogFile          string `mapstructure:"LOG_FILE" json:"LOG_FILE"`
	LogLevel         string `mapstructure:"LOG_LEVEL" json:"LOG_LEVEL"`
	ReplayFolderPath string `mapstructure:"REPLAY_PATH" json:"REPLAY_PATH"`
	Database         string `mapstructure:"DATABASE" json:"DATABASE"`
	GenjiDbPath      string `mapstructure:"GENJI_DB_PATH" json:"GENJI_DB_PATH"`

	MaxTryLogin        uint64 `mapstructure:"MAX_TRY_LOGIN" json:"MAX_TRY_LOGIN"`
	RetryAliveCountMax int    `mapstructure:"RETRY_ALIVE_COUNT_MAX" json:"RETRY_ALIVE_COUNT_MAX"`
	ReuseConnection    bool   `mapstructure:"REUSE_CONNECTION" json:"REUSE_CONNECTION"`
	DisableRecorder    bool   `mapstructure:"DISABLE_RECORDER" json:"DISABLE_RECORDER"`

	EnableLocalPortForward bool `mapstructure:"ENABLE_LOCAL_PORT_FORWARD" json:"ENABLE_LOCAL_PORT_FORWARD"`
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
		ReplayFolderPath: "gojumpreplay",
		Database:         "genji",
		GenjiDbPath:      "gojumpdb",
		OtpDuration:      120,
		MaxTryLogin:      15,
		LoginBlockTime:   5,
	}
}
