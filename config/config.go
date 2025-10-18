package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	Server   ServerConfig   `yaml:"server"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

type RedisConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"PORT"`
	Db   int    `yaml:"db"`
}

type JWTConfig struct {
	Secret     string `yaml:"secret"`
	ExpiryHour int    `yaml:"expiry_hour"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

var appConfig *Config

func Load(path string) *Config {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}
	appConfig = &config
	return &config
}

func GetConfig() *Config {
	if appConfig == nil {
		return Load("config.yaml")
	}
	return appConfig
}
