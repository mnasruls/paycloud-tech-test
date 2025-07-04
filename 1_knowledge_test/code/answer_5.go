package main

import (
	"fmt"
	"sync"
	"time"
)

func answer_5() {
	fmt.Println("=== Queue Using Mutex Example ===")
	queueUsingMutex()
	fmt.Println("=== Queue Using Channel Example ===")
	queueUsingChannel()
}

func queueUsingMutex() {
	type BankAccount struct {
		mu      sync.Mutex
		balance int
	}

	deposit := func(account *BankAccount, amount int, id int) {
		account.mu.Lock()
		defer account.mu.Unlock()

		oldBalance := account.balance
		time.Sleep(time.Millisecond * 10) // Simulasi operasi yang memakan waktu
		account.balance = oldBalance + amount
		fmt.Printf("Goroutine %d: Deposit %d, Balance: %d -> %d\n",
			id, amount, oldBalance, account.balance)
	}

	account := &BankAccount{balance: 1000}
	var wg sync.WaitGroup

	// Multiple goroutines trying to deposit simultaneously
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			deposit(account, 100, id)
		}(i)
	}

	wg.Wait()
	fmt.Printf("Final balance: %d\n", account.balance)
}

func queueUsingChannel() {
	queue := make(chan string, 5) // Buffer size 5

	go func() {
		for i := 1; i <= 10; i++ {
			item := fmt.Sprintf("Item-%d", i)
			queue <- item
			fmt.Printf("Menambahkan %s ke antrian\n", item)
			time.Sleep(time.Millisecond * 100)
		}
		close(queue)
	}()

	for item := range queue {
		fmt.Printf("Memproses %s\n", item)
		time.Sleep(time.Millisecond * 200)
	}
}
