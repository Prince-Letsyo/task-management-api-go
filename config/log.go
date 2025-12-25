package config

import "github.com/sirupsen/logrus"

type Graylog struct {
	Host string `mapstructure:"GRAYLOG_HOST" yaml:"host" env:"GRAYLOG_HOST" env-default:"localhost"`
	Port string `mapstructure:"GRAYLOG_PORT" yaml:"port" env:"GRAYLOG_PORT" env-default:"12201"`
}

type FileLog struct {
	Path       string `mapstructure:"LOG_PATH" yaml:"path" env-default:"storage/logs"`
	TimeFormat string `mapstructure:"LOG_TIME_PATH" yaml:"timeformat" env-default:"2006-01-19"`
}

type ConsoleLog struct {
	Level string `mapstructure:"CONSOLE_LOG_LEVEL" yaml:"level" env-default:"info"`
	Show  string `mapstructure:"CONSOLE_LOG_SHOW" yaml:"show" env-default:"false"`
}

type LogConfig struct {
	TimeField  string     `yaml:"timefield"`
	TimeFormat string     `yaml:"timeformat"`
	ConsoleLog ConsoleLog `yaml:"console"`
	Monitor    Graylog    `yaml:"monitor"`
	InfoLevel  FileLog    `yaml:"info"`
	WarnLevel  FileLog    `yaml:"warn"`
	ErrorLevel FileLog    `yaml:"error"`
}

type Logger struct {
	Level string `json:"level" yaml:"level" env:"LEVEL" env-default:"info"`
}

func (logCfg *LogConfig) NewLogger() *logrus.Logger {
	logger := logrus.New()
	lvl, err := logrus.ParseLevel(logCfg.ConsoleLog.Level)
	if err != nil {
		logger.WithError(err).Fatalf("cannot parse logrus level [%s]", logCfg.ConsoleLog.Level)
	}
	logger.SetLevel(lvl)
	logger.Formatter = &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "",
		// DisableSorting:         true,
		DisableLevelTruncation: true,
		PadLevelText:           true,
		QuoteEmptyFields:       true,
	}
	return logger
}
