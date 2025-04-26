package minimax

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// APIError API error structure
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("MinimaxAPI error (status code: %d): %s", e.StatusCode, e.Message)
}

// APIClient encapsulates Minimax API calls
type APIClient struct {
	APIKey  string
	APIHost string
}

// Post sends a POST request to the MiniMax API
func (c *APIClient) Post(endpoint string, jsonData interface{}) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s%s", c.APIHost, endpoint)

	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		return nil, fmt.Errorf("JSON encoding failed: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	client := &http.Client{
		Timeout: time.Second * 30,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(bodyBytes),
		}
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("response parsing failed: %v", err)
	}

	// 检查base_resp.status_code
	if baseResp, ok := result["base_resp"].(map[string]interface{}); ok {
		if statusCode, ok := baseResp["status_code"].(float64); ok && statusCode != 0 {
			statusMsg := "unknown error"
			if msg, ok := baseResp["status_msg"].(string); ok {
				statusMsg = msg
			}
			return nil, fmt.Errorf("API error: status_code=%v, message=%s", statusCode, statusMsg)
		}
	}

	return result, nil
}

// Get sends a GET request to the MiniMax API
func (c *APIClient) Get(endpoint string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s%s", c.APIHost, endpoint)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	client := &http.Client{
		Timeout: time.Second * 30,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(bodyBytes),
		}
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("response parsing failed: %v", err)
	}

	// 检查base_resp.status_code
	if baseResp, ok := result["base_resp"].(map[string]interface{}); ok {
		if statusCode, ok := baseResp["status_code"].(float64); ok && statusCode != 0 {
			statusMsg := "unknown error"
			if msg, ok := baseResp["status_msg"].(string); ok {
				statusMsg = msg
			}
			return nil, fmt.Errorf("API error: status_code=%v, message=%s", statusCode, statusMsg)
		}
	}

	return result, nil
}
