package conf

import "time"

type Config struct {
	Server   Server   `json:"server" yaml:"server"`
	Data     Data     `json:"data" yaml:"data"`
	Registry Registry `json:"registry" yaml:"registry"`
	Auth     Auth     `json:"auth" yaml:"auth"`
}

type Auth struct {
	JwtSecret        string `json:"jwt_secret" yaml:"jwt_secret"`
	TokenExpireHours int    `json:"token_expire_hours" yaml:"token_expire_hours"`
}

type Server struct {
	Grpc Grpc `json:"grpc" yaml:"grpc"`
}

type Grpc struct {
	Network string        `json:"network" yaml:"network"`
	Addr    string        `json:"addr" yaml:"addr"`
	Timeout time.Duration `json:"timeout" yaml:"timeout"`
}

type Data struct {
	Database Database `json:"database" yaml:"database"`
}

type Database struct {
	Driver string `json:"driver" yaml:"driver"`
	Source string `json:"source" yaml:"source"`
}

type Registry struct {
	Endpoints []string `json:"endpoints" yaml:"endpoints"`
}
