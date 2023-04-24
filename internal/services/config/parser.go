package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/traefik/paerser/env"
	"github.com/traefik/paerser/file"
)

func Parse[T any](cfgFile string, envPrefix string, def ...T) (cfg T, err error) {
	cfg = Opt(def)

	if err = file.Decode(cfgFile, &cfg); err != nil && !os.IsNotExist(err) {
		return
	}

	godotenv.Load()
	if err = env.Decode(os.Environ(), envPrefix, &cfg); err != nil {
		return
	}

	return
}

func Opt[T any](v []T, def ...T) T {
	if len(v) == 0 {
		var altDef T
		return Opt(def, altDef)
	}
	return v[0]
}
