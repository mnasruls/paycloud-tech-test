package main

import (
	answer7 "knowladge-test/answer_7"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	// uncomment the following lines to run the answers
	// answer_3()
	// answer_5()
	// answer_10()

	// answer 7
	var wg sync.WaitGroup
	wg.Add(1)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	answer7.Answer_7()
	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %v, initiating graceful shutdown...", sig)
		wg.Done()
	}()
	wg.Wait()
}
