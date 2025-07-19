package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Payload struct {
	Data    string `json:"data"`
	Message string `json:"message"`
}

type JobBody struct {
	Type     string  `json:"type"`
	Payload  Payload `json:"payload"`
	Priority string  `json:"priority"`
	Delay    int     `json:"delay"`
}

func submitJob(priority string, count int) {
	url := "http://localhost:8000/apis/v1/submit-job"

	for i := 1; i <= count; i++ {
		job := JobBody{
			Type: "Email",
			Payload: Payload{
				Data:    fmt.Sprintf("Payload #%d [%s]", i, priority),
				Message: "Queued by script",
			},
			Priority: priority,
		}

		jobJSON, err := json.Marshal(job)
		if err != nil {
			fmt.Println("Failed to marshal JSON:", err)
			continue
		}

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jobJSON))
		if err != nil {
			fmt.Printf("Failed to send job #%d [%s]: %v\n", i, priority, err)
			continue
		}
		defer resp.Body.Close()

		fmt.Printf("Sent job #%d [%s], status: %s\n", i, priority, resp.Status)
	}
}

func main() {
	submitJob("HIGH", 10000)
	submitJob("MEDIUM", 100)
	submitJob("LOW", 100)
}
