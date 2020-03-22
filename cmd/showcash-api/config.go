package main

import "github.com/17twenty/showcash-api/pkg/env"

type databaseConfig struct {
	User     string
	Password string
	Host     string
	Port     int
	Name     string
}

type config struct {
	UseS3    bool
	Database databaseConfig
}

func loadConfig(strict *bool) *config {
	env.FatalOnMissingEnv = *strict
	return &config{
		UseS3: env.GetAsBool("USES3", false),
		Database: databaseConfig{
			User:     env.GetAsString("DB_USER", "local"),
			Password: env.GetAsString("DB_PASSWORD", "asecurepassword"),
			Host:     env.GetAsString("DB_HOST", "localhost"),
			Port:     env.GetAsInt("DB_PORT", 5003),
			Name:     env.GetAsString("DB_NAME", "showcash"),
		},
	}
}
