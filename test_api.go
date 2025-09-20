package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	// Test API with a real plate
	plate := "GEH5A72" // Real plate from your memory
	apiURL := fmt.Sprintf("https://wdapi2.com.br/consulta/%s/4f624c5b7ddb8b746d947fb22983eaa3", plate)

	fmt.Printf("Testing API with plate: %s\n", plate)
	fmt.Printf("URL: %s\n", apiURL)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Make HTTP request
	resp, err := client.Get(apiURL)
	if err != nil {
		fmt.Printf("❌ HTTP request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("📡 Response Status: %d\n", resp.StatusCode)

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Failed to read response: %v\n", err)
		return
	}

	fmt.Printf("📄 Response Body:\n%s\n", string(bodyBytes))

	// Try to parse as JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &jsonData); err != nil {
		fmt.Printf("❌ Failed to parse JSON: %v\n", err)
		return
	}

	fmt.Printf("✅ Successfully parsed JSON response\n")

	// Check for error message
	if mensagem, exists := jsonData["mensagemRetorno"]; exists {
		fmt.Printf("📝 Message: %v\n", mensagem)
	}
}

