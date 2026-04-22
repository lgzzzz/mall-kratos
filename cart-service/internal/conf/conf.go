package conf

import "time"

type Config struct {
	Server      Server      `json:"server" yaml:"server"`
	Data        Data        `json:"data" yaml:"data"`
	Registry    Registry    `json:"registry" yaml:"registry"`
	Auth        Auth        `json:"auth" yaml:"auth"`
	GrpcClients GrpcClients `json:"grpc_clients" yaml:"grpc_clients"`
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

type Auth struct {
	JwtSecret string `json:"jwt_secret" yaml:"jwt_secret"`
}

type GrpcClients struct {
	ProductService string `json:"product_service" yaml:"product_service"`
}
