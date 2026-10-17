package apiclient

import "os"

const defaultBaseURL = "http://localhost:8000"

// API_URL pointe vers URL ngrok
func BaseURLFromEnv() string {
	if url := os.Getenv("API_URL"); url != "" {
		return url
	}
	return defaultBaseURL
}
