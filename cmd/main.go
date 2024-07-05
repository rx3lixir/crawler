package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/rx3lixir/crawler/appconfig"
	"github.com/rx3lixir/crawler/telegram"
)

func main() {
	configFile := flag.String("config", "", "Path to config file")
	flag.Parse()

	// Loading data
	if err := appconfig.LoadConfig(*configFile); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Println("Configuration loaded successfully")

	// Starting the Puppejteer server
	puppeteerCmd := startPuppeteerServer()
	defer stopPuppeteerServer(puppeteerCmd)

	// Wait for the Puppeteer server to start
	time.Sleep(5 * time.Second)

	// Running bot instance
	telegram.StartBot(*appconfig.CrawlerApp)
}

func startPuppeteerServer() *exec.Cmd {
	cmd := exec.Command("node", "web/scrape.js")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Fatalf("Failed to Start Puppeteer server: %v", err)
	}

	log.Println("Puppeteer server started")

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		stopPuppeteerServer(cmd)
		os.Exit(0)
	}()

	return cmd
}

func stopPuppeteerServer(cmd *exec.Cmd) {
	if cmd.Process != nil {
		log.Println("Stopping Puppeteer server")
		err := cmd.Process.Signal(syscall.SIGTERM)
		if err != nil {
			log.Printf("Error stopping Puppetter server: %v", err)
			return
		}
		_, err = cmd.Process.Wait()
		if err != nil {
			log.Printf("Error wainting for Puppeteer server to stop: %v", err)
		}
		log.Println("Puppeteer server stopped")
	}
}
