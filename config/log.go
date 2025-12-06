package config

type Graylog struct{
	Host string `mapstructure:"GRAYLOG_HOST" yaml:"host" env:"GRAYLOG_HOST" env-default:"localhost"`
	Port string `mapstructure:"GRAYLOG_PORT" yaml:"port" env:"GRAYLOG_PORT" env-default:"12201"`
}

type Filelog struct {
	Path string `mapstructure:"LOG_PATH" yaml:"path" env-default:"storage/logs"`
	TimeFormat string `mapstructure:"LOG_TIME_PATH" yaml:"timeformat" env-default:"2006-01-19"`
}

type ConsoleLog struct{
	Level string `mapstructure:"CONSOLE_LOG_LEVEL" yaml:"level" env-default:""`
}
