package main

import (
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/neticlabs/env"
)

type Config struct {
	Port                    uint16   `env:"PORT"                      envDefault:"8700" validate:"required,number,gte=0,lte=65535"`
	APIKeys                 []string `env:"API_KEYS"                  envSeparator:","  validate:"required"`
	MaxWorkspaces           uint64   `env:"MAX_WORKSPACES"            envDefault:"4"    validate:"required,number,gte=0,lte=4"`
	MaxWorkspaceConnections uint64   `env:"MAX_WORKSPACE_CONNECTIONS" envDefault:"10"   validate:"required,number,gte=0,lte=255"`
}

func newConfig() *Config {
	var c Config

	if err := env.Parse(&c); err != nil {
		log.Fatal(err)
	}

	err := validator.New(validator.WithRequiredStructEnabled()).Struct(&c)
	if err != nil {
		log.Fatal(err)
	}

	return &c
}
