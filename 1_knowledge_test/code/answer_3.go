package main

import (
	"fmt"
	"sync"
	"time"
)

func worker(id int, wg *sync.WaitGroup, jobsChan <-chan int32, resultsChan chan<- string) {
	defer wg.Done()

	for job := range jobsChan {
		fmt.Printf("Worker %d memulai job %d\n", id, job)

		// Simulasi pekerjaan yang memakan waktu
		time.Sleep(time.Second)

		result := fmt.Sprintf("Worker %d menyelesaikan job %d", id, job)
		resultsChan <- result
	}
}

func answer_3() {
	const numJobs = 5
	const numWorkers = 3

	jobsChan := make(chan int32, numJobs)
	resultsChan := make(chan string, numJobs)

	var wg sync.WaitGroup

	fmt.Println("Memulai worker...")
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go worker(w, &wg, jobsChan, resultsChan)
	}

	fmt.Println("Mengirim pekerjaan...")
	for j := 1; j <= numJobs; j++ {
		jobsChan <- int32(j)
	}

	close(jobsChan)

	//  Menunggu semua worker selesai.
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	fmt.Println("Menunggu hasil...")

	//  Mengumpulkan semua hasil
	for result := range resultsChan {
		fmt.Println(result)
	}

	fmt.Println("Semua pekerjaan selesai.")
}
