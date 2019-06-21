package main

import (
	"github.com/garyburd/redigo/redis"
	"github.com/go-nacelle/nacelle"
)

type RedisInitializer struct {
	Services nacelle.ServiceContainer `service:"services"`
	conn     redis.Conn
}

type Config struct {
	RedisAddr string `env:"redis_addr" required:"true"`
}

func NewRedisInitializer() nacelle.Initializer {
	return &RedisInitializer{}
}

func (ri *RedisInitializer) Init(config nacelle.Config) error {
	redisConfig := &Config{}
	if err := config.Load(redisConfig); err != nil {
		return err
	}

	conn, err := redis.DialURL(redisConfig.RedisAddr)
	if err != nil {
		return err
	}

	ri.conn = conn
	return ri.Services.Set("redis", conn)
}

func (ri *RedisInitializer) Finalize() error {
	ri.conn.Close()
	return nil
}
