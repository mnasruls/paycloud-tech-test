package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

// ItemDetails represents the detailed information for an item
type ItemDetails struct {
	ID          string
	Name        string
	Description string
	Price       float64
}

// simulateFetchItemDetails simulates an external API call to fetch item details
// It introduces random delays and occasional errors
// It respects the context for cancellation
func simulateFetchItemDetails(ctx context.Context, itemID string) (*ItemDetails, error) {
	// Simulate network latency
	delay := time.Duration(500+rand.Intn(1500)) * time.Millisecond // 0.5s to 1.5s
	select {
	case <-time.After(delay):
		// Continue
	case <-ctx.Done():
		log.Printf("Context cancelled for item %s, aborting fetch.", itemID)
		return nil, ctx.Err() // Context cancelled
	}

	// Simulate occasional API errors
	if rand.Intn(100) < 15 { // 15% chance of error
		log.Printf("Simulated API error for item %s", itemID)
		return nil, fmt.Errorf("simulated API error for item %s: service unavailable", itemID)
	}

	details := &ItemDetails{
		ID:          itemID,
		Name:        fmt.Sprintf("Product %s", itemID),
		Description: fmt.Sprintf("Detailed description for product %s.", itemID),
		Price:       rand.Float64() * 100,
	}

	log.Printf("Successfully fetched details for item %s", itemID)
	return details, nil
}

// FetchResult represents the result of fetching a single item
type FetchResult struct {
	ItemID  string
	Details *ItemDetails
	Error   error
}

// FetchAndAggregate fetches item details concurrently with controlled concurrency and timeout
func FetchAndAggregate(
	ctx context.Context,
	itemIDs []string,
	maxConcurrent int,
	perItemTimeout time.Duration,
) (map[string]ItemDetails, []error) {
	// Initialize result containers
	results := make(map[string]ItemDetails)
	var errors []error
	var mu sync.Mutex // Mutex to protect shared data
	var wg sync.WaitGroup
	bufferChannel := make(chan struct{}, maxConcurrent)
	resultChan := make(chan FetchResult, len(itemIDs))

	for _, itemID := range itemIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			select {
			case bufferChannel <- struct{}{}:
			case <-ctx.Done():
				// Global context cancelled
				resultChan <- FetchResult{
					ItemID: id,
					Error:  ctx.Err(),
				}
				return
			}

			// Release
			defer func() { <-bufferChannel }()
			itemCtx, cancel := context.WithTimeout(ctx, perItemTimeout)
			defer cancel()

			details, err := simulateFetchItemDetails(itemCtx, id)
			resultChan <- FetchResult{
				ItemID:  id,
				Details: details,
				Error:   err,
			}
		}(itemID)
	}

	// Close result channel when all goroutines are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for result := range resultChan {
		mu.Lock()
		if result.Error != nil {
			// Add error with item ID context
			errorWithContext := fmt.Errorf("item %s: %w", result.ItemID, result.Error)
			errors = append(errors, errorWithContext)
			log.Printf("Failed to fetch item %s: %v", result.ItemID, result.Error)
		} else {
			// Add successful result
			results[result.ItemID] = *result.Details
			log.Printf("Successfully processed item %s", result.ItemID)
		}
		mu.Unlock()
	}

	return results, errors
}

func main() {
	// rand.Seed(time.Now().UnixNano())
	// Test data
	itemIDs := []string{"001", "002", "003", "004", "005", "006", "007", "008", "009", "010"}
	maxConcurrent := 4
	perItemTimeout := 1500 * time.Millisecond
	globalTimeout := 10 * time.Second // change this for testing. example: 100 * time.Millisecond

	log.Println("=== Testing ===")

	ctx, cancel := context.WithTimeout(context.Background(), globalTimeout)
	defer cancel()

	start := time.Now()
	results, errors := FetchAndAggregate(ctx, itemIDs, maxConcurrent, perItemTimeout)

	duration := time.Since(start)

	// Print results
	log.Printf("\n=== TEST RESULTS ===")
	log.Printf("Total execution time: %v", duration)
	log.Printf("Successful items: %d", len(results))
	log.Printf("Failed items: %d", len(errors))

	log.Println("Successful items:")
	for id, details := range results {
		log.Printf("- %s: %s (Price: $%.2f)", id, details.Name, details.Price)
	}

	log.Println("Errors:")
	for _, err := range errors {
		log.Printf("- %v", err)
	}
}
