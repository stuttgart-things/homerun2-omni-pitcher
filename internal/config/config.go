package config

import homerun "github.com/stuttgart-things/homerun-library"

type RedisConfig struct {
	Addr     string
	Port     string
	Password string
	Stream   string
}

func LoadRedisConfig() RedisConfig {
	return RedisConfig{
		Addr:     homerun.GetEnv("REDIS_ADDR", "localhost"),
		Port:     homerun.GetEnv("REDIS_PORT", "6379"),
		Password: homerun.GetEnv("REDIS_PASSWORD", ""),
		Stream:   homerun.GetEnv("REDIS_STREAM", "messages"),
	}
}

func (c RedisConfig) ToMap() map[string]string {
	return map[string]string{
		"addr":     c.Addr,
		"port":     c.Port,
		"password": c.Password,
		"stream":   c.Stream,
	}
}
