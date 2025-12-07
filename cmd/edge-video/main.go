package main

import (
	"flag"
	"log"

	"github.com/T3-Labs/edge-video/internal/app"
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "config.toml", "Caminho para o arquivo de configuração")
	flag.Parse()

	// Initialize and run the application
	application := app.New(*configFile)
	if err := application.Run(); err != nil {
		log.Fatalf("Erro fatal na aplicação: %v", err)
	}
}
