package config

import (
	"flag"
	"os"
)

type Flags struct {
	ConfigFile string
}

func ParseFlags() *Flags {
	flags := &Flags{}

	flag.StringVar(&flags.ConfigFile, "config.file", "config.yaml", "Path to the configuration file")
	flag.Parse()

	// Se o arquivo de configuração não existir, retorna erro
	if _, err := os.Stat(flags.ConfigFile); os.IsNotExist(err) {
		flag.Usage()
		os.Exit(1)
	}

	return flags
}
