package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/YehiaGewily/Agent-Mesh/internal/models"
)

// TaskRequest matches the JSON structure expected by the Producer
type TaskRequest struct {
	AgentType string                 `json:"agent_type"`
	Priority  int                    `json:"priority"`
	Payload   map[string]interface{} `json:"payload"`
}

const (
	URL = "http://localhost:8081/v1/tasks"
)

func main() {
	count := flag.Int("count", 500, "Total number of tasks to create")
	concurrency := flag.Int("concurrency", 10, "Number of concurrent workers")
	flag.Parse()

	log.Printf("Starting stress test: %d tasks with %d concurrency...", *count, *concurrency)
	start := time.Now()

	tasks := make(chan int, *count)
	var wg sync.WaitGroup

	// Start Workers
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go worker(&wg, tasks)
	}

	// Enqueue Jobs
	for i := 0; i < *count; i++ {
		tasks <- i
	}
	close(tasks)

	// Wait
	wg.Wait()
	duration := time.Since(start)
	log.Printf("Done! Sent %d tasks in %v (%.2f req/s)", *count, duration, float64(*count)/duration.Seconds())
}

func worker(wg *sync.WaitGroup, tasks <-chan int) {
	defer wg.Done()
	client := &http.Client{Timeout: 5 * time.Second}

	// Original agentTypes declaration (kept as per instruction context)
	_ = []string{"MAGNUS_STRATEGIST", "CEDRIC_WRITER", "LYRA_AUDITOR"}

	for range tasks {
		// Updated agentTypes declaration inside the loop
		agentTypes := []string{
			models.AgentTypeArchitect,
			models.AgentTypeDeveloper,
			models.AgentTypeQA,
		}
		agentType := agentTypes[rand.Intn(len(agentTypes))]
		priority := rand.Intn(5) + 1

		reqBody := TaskRequest{
			AgentType: agentType,
			Priority:  priority,
			Payload: map[string]interface{}{
				"source": "stress_test",
				"ts":     time.Now().Unix(),
				"note":   "Performance Check",
			},
		}

		data, _ := json.Marshal(reqBody)
		resp, err := client.Post(URL, "application/json", bytes.NewBuffer(data))
		if err != nil {
			log.Printf("Request failed: %v", err)
			continue
		}

		// Consume body to reuse connection
		resp.Body.Close()

		if resp.StatusCode != http.StatusAccepted {
			log.Printf("Unexpected status: %s", resp.Status)
		}
	}
}
