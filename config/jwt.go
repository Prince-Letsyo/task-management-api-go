package config

type JwtConfig struct {
	Secret string `yaml:"secret"`
	Expire int64  `yaml:"expire"`
}

type Refresh struct {
	Token  string
	Expire string
}

type Access struct {
	Token  string
	Expire string
}
type JwtSecrets struct {
	Refresh Refresh
	Access  Access
	App     JwtConfig `yaml:"app"`
	Api     JwtConfig `yaml:"api"`
}
