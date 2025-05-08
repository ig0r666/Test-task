package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"testtask/serverpool/config"
)

// Это пул серверов - он нужен нашему балансировщику, чтобы распределять нагрузку между этими серверами
// с помощью этого пула было проведено тестирование работоспособности балансировщика
func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()

	cfg := config.MustLoad(configPath)
	addrs := strings.Split(cfg.URLs, ",")

	for _, addr := range addrs {
		mux := http.NewServeMux()

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Server running on %s\n", addr)
		})

		server := &http.Server{
			Addr:    addr,
			Handler: mux,
		}

		go func() {
			log.Printf("Starting server on %s", addr)
			if err := server.ListenAndServe(); err != nil {
				log.Fatalf("Failed to start server on %s: %v", addr, err)
			}
		}()
	}

	select {}
}
