package minimax

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mcp/minimax/server/define"
	"mcp/minimax/server/storage"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ThinkInAIXYZ/go-mcp/protocol"
)

// MinimaxMCPServer MCP server instance
type MinimaxMCPServer struct {
	Client       *MinimaxAPIClient
	BasePath     string
	ResourceMode string
}

// API method implementations

// HandleTextToAudio processes text-to-speech requests, save To Local
func (s *MinimaxMCPServer) HandleTextToAudio(req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	var params TextToAudioRequest
	if err := protocol.VerifyAndUnmarshal(req.RawArguments, &params); err != nil {
		return createTextErrorResult(fmt.Sprintf("Parameter parsing failed: %v", err)), nil
	}

	// Check required parameters
	if params.Text == "" {
		return createTextErrorResult("The text parameter must be provided"), nil
	}

	// Fill optional parameters with default values
	if params.VoiceID == "" {
		params.VoiceID = define.DefaultVoiceID
	}
	if params.Model == "" {
		params.Model = define.DefaultSpeechModel
	}
	if params.Speed == 0 {
		params.Speed = define.DefaultSpeed
	}
	if params.Vol == 0 {
		params.Vol = define.DefaultVolume
	}
	// Pitch defaults to 0, no need to check
	if params.Emotion == "" {
		params.Emotion = define.DefaultEmotion
	}
	if params.SampleRate == 0 {
		params.SampleRate = define.DefaultSampleRate
	}
	if params.Bitrate == 0 {
		params.Bitrate = define.DefaultBitrate
	}
	if params.Channel == 0 {
		params.Channel = define.DefaultChannel
	}
	if params.Format == "" {
		params.Format = define.DefaultFormat
	}
	if params.LanguageBoost == "" {
		params.LanguageBoost = define.DefaultLanguageBoost
	}

	// Build request payload
	payload := map[string]interface{}{
		"model": params.Model,
		"text":  params.Text,
		"voice_setting": map[string]interface{}{
			"voice_id": params.VoiceID,
			"speed":    params.Speed,
			"vol":      params.Vol,
			"pitch":    params.Pitch,
			"emotion":  params.Emotion,
		},
		"audio_setting": map[string]interface{}{
			"sample_rate": params.SampleRate,
			"bitrate":     params.Bitrate,
			"format":      params.Format,
			"channel":     params.Channel,
		},
		"language_boost": params.LanguageBoost,
	}

	// If resource mode is URL, add output format
	if s.ResourceMode == define.ResourceModeURL {
		payload["output_format"] = "url"
	}

	// Call API
	response, err := s.Client.Post("/v1/t2a_v2", payload)
	if err != nil {
		return createTextErrorResult(fmt.Sprintf("API call failed: %v", err)), nil
	}

	// Process response
	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return createTextErrorResult("Invalid API response format: missing data field"), nil
	}

	audioData, ok := data["audio"].(string)
	if !ok || audioData == "" {
		return createTextErrorResult("Invalid API response format: unable to get audio data"), nil
	}

	// Return different results based on resource mode
	if s.ResourceMode == define.ResourceModeURL {
		return createTextResult(fmt.Sprintf("Success. Audio URL: %s", audioData)), nil
	}

	// Convert hex string to binary data
	audioBytes, err := hex.DecodeString(audioData)
	if err != nil {
		return createTextErrorResult(fmt.Sprintf("Failed to decode audio data: %v", err)), nil
	}

	// Save audio file
	outputPath := storage.BuildOutputPath(params.OutputDirectory, s.BasePath)
	outputFileName := storage.BuildOutputFile("t2a", params.Text, outputPath, params.Format)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(outputFileName), 0755); err != nil {
		return createTextErrorResult(fmt.Sprintf("Failed to create output directory: %v", err)), nil
	}

	// Write file
	if err := ioutil.WriteFile(outputFileName, audioBytes, 0644); err != nil {
		return createTextErrorResult(fmt.Sprintf("Failed to save audio file: %v", err)), nil
	}

	return createTextResult(fmt.Sprintf("Success. File saved as: %s. Voice used: %s", outputFileName, params.VoiceID)), nil
}

// HandleListVoices processes list voices requests
func (s *MinimaxMCPServer) HandleListVoices(req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	var params ListVoicesRequest
	if err := protocol.VerifyAndUnmarshal(req.RawArguments, &params); err != nil {
		return createTextErrorResult(fmt.Sprintf("Parameter parsing failed: %v", err)), nil
	}

	// Use default values
	if params.VoiceType == "" {
		params.VoiceType = "all"
	}

	// Call API
	payload := map[string]interface{}{
		"voice_type": params.VoiceType,
	}

	response, err := s.Client.Post("/v1/get_voice", payload)
	if err != nil {
		return createTextErrorResult(fmt.Sprintf("API call failed: %v", err)), nil
	}

	// Process response
	systemVoices, okSystem := response["system_voice"].([]interface{})
	voiceCloningVoices, okCloning := response["voice_cloning"].([]interface{})

	if !okSystem && !okCloning {
		return createTextErrorResult("Invalid API response format: missing voice information"), nil
	}

	// Build result
	var resultText strings.Builder
	resultText.WriteString("Available voices list:\n\n")

	// System voices
	if okSystem && len(systemVoices) > 0 {
		resultText.WriteString("System voices:\n")
		for i, voice := range systemVoices {
			voiceMap, ok := voice.(map[string]interface{})
			if ok {
				voiceName := voiceMap["voice_name"]
				voiceID := voiceMap["voice_id"]
				resultText.WriteString(fmt.Sprintf("%d. Name: %v, ID: %v\n", i+1, voiceName, voiceID))
			}
		}
		resultText.WriteString("\n")
	} else {
		resultText.WriteString("No system voices\n\n")
	}

	// Voice cloning
	if okCloning && len(voiceCloningVoices) > 0 {
		resultText.WriteString("Voice cloning:\n")
		for i, voice := range voiceCloningVoices {
			voiceMap, ok := voice.(map[string]interface{})
			if ok {
				voiceName := voiceMap["voice_name"]
				voiceID := voiceMap["voice_id"]
				resultText.WriteString(fmt.Sprintf("%d. Name: %v, ID: %v\n", i+1, voiceName, voiceID))
			}
		}
	} else {
		resultText.WriteString("No voice cloning\n")
	}

	return createTextResult(resultText.String()), nil
}

// HandleVoiceClone processes voice cloning requests
func (s *MinimaxMCPServer) HandleVoiceClone(req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	var params VoiceCloneRequest
	if err := protocol.VerifyAndUnmarshal(req.RawArguments, &params); err != nil {
		return createTextErrorResult(fmt.Sprintf("Parameter parsing failed: %v", err)), nil
	}

	// Validate required parameters
	if params.VoiceID == "" {
		return createTextErrorResult("The voice_id parameter must be provided"), nil
	}
	if params.File == "" {
		return createTextErrorResult("The file parameter must be provided"), nil
	}
	if params.Text == "" {
		return createTextErrorResult("The text parameter must be provided"), nil
	}

	var fileID string
	var err error

	// Step 1: Upload file
	if params.IsURL {
		// Download file from URL
		resp, err := http.Get(params.File)
		if err != nil {
			return createTextErrorResult(fmt.Sprintf("Failed to download file: %v", err)), nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return createTextErrorResult(fmt.Sprintf("Failed to download file, status code: %d", resp.StatusCode)), nil
		}

		// Create temporary file
		tempFile, err := ioutil.TempFile("", "voice_clone_*.mp3")
		if err != nil {
			return createTextErrorResult(fmt.Sprintf("Failed to create temporary file: %v", err)), nil
		}
		defer os.Remove(tempFile.Name())
		defer tempFile.Close()

		// Write content to temporary file
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return createTextErrorResult(fmt.Sprintf("Failed to read response: %v", err)), nil
		}

		if _, err := tempFile.Write(bodyBytes); err != nil {
			return createTextErrorResult(fmt.Sprintf("Failed to write to temporary file: %v", err)), nil
		}
		tempFile.Close()

		// Upload temporary file
		fileID, err = s.uploadFile(tempFile.Name())
		if err != nil {
			return createTextErrorResult(fmt.Sprintf("Failed to upload file: %v", err)), nil
		}
	} else {
		// Upload local file
		if _, err := os.Stat(params.File); os.IsNotExist(err) {
			return createTextErrorResult(fmt.Sprintf("Local file does not exist: %s", params.File)), nil
		}

		fileID, err = s.uploadFile(params.File)
		if err != nil {
			return createTextErrorResult(fmt.Sprintf("Failed to upload file: %v", err)), nil
		}
	}

	// Check fileID
	if fileID == "" {
		return createTextErrorResult("File upload failed, fileID not obtained"), nil
	}

	// Step 2: Clone voice
	payload := map[string]interface{}{
		"file_id":  fileID,
		"voice_id": params.VoiceID,
	}

	if params.Text != "" {
		payload["text"] = params.Text
		payload["model"] = define.DefaultSpeechModel
	}

	response, err := s.Client.Post("/v1/voice_clone", payload)
	if err != nil {
		return createTextErrorResult(fmt.Sprintf("Voice cloning API call failed: %v", err)), nil
	}

	demoAudio, ok := response["demo_audio"].(string)
	if !ok || demoAudio == "" {
		// There may be no demo audio, just return success message
		return createTextResult(fmt.Sprintf("Voice cloning successful. Voice ID: %s", params.VoiceID)), nil
	}

	// If in URL mode, return URL directly
	if s.ResourceMode == define.ResourceModeURL {
		return createTextResult(fmt.Sprintf("Success. Demo audio URL: %s", demoAudio)), nil
	}

	// Download demo audio
	resp, err := http.Get(demoAudio)
	if err != nil {
		return createTextErrorResult(fmt.Sprintf("Failed to download demo audio: %v", err)), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return createTextErrorResult(fmt.Sprintf("Failed to download demo audio, status code: %d", resp.StatusCode)), nil
	}

	// Read audio content
	audioBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return createTextErrorResult(fmt.Sprintf("Failed to read demo audio: %v", err)), nil
	}

	// Save demo audio file
	outputPath := storage.BuildOutputPath(params.OutputDirectory, s.BasePath)
	outputFileName := storage.BuildOutputFile("voice_clone", params.Text, outputPath, "wav")

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(outputFileName), 0755); err != nil {
		return createTextErrorResult(fmt.Sprintf("Failed to create output directory: %v", err)), nil
	}

	// Write file
	if err := os.WriteFile(outputFileName, audioBytes, 0644); err != nil {
		return createTextErrorResult(fmt.Sprintf("Failed to save audio file: %v", err)), nil
	}

	return createTextResult(fmt.Sprintf("Voice cloning successful: Voice ID: %s, demo audio saved as: %s", params.VoiceID, outputFileName)), nil
}

// uploadFile uploads a file to the MiniMax API
func (s *MinimaxMCPServer) uploadFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Create multipart request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %v", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return "", fmt.Errorf("failed to copy file content: %v", err)
	}

	// Add purpose field
	if err := writer.WriteField("purpose", "voice_clone"); err != nil {
		return "", fmt.Errorf("failed to add purpose field: %v", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %v", err)
	}

	// Create request
	url := fmt.Sprintf("%s/v1/files/upload", s.Client.APIHost)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.Client.APIKey))

	// Send request
	client := &http.Client{
		Timeout: time.Second * 60, // Upload may take longer
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("API request error (status code: %d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("response parsing failed: %v", err)
	}

	// Get file_id
	fileObj, ok := result["file"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unable to get file field from response")
	}

	fileID, ok := fileObj["file_id"].(string)
	if !ok || fileID == "" {
		return "", fmt.Errorf("unable to get file_id from response")
	}

	return fileID, nil
}

// HandleGenerateVideo processes video generation requests
func (s *MinimaxMCPServer) HandleGenerateVideo(req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	var params GenerateVideoRequest
	if err := protocol.VerifyAndUnmarshal(req.RawArguments, &params); err != nil {
		return createTextErrorResult(fmt.Sprintf("Parameter parsing failed: %v", err)), nil
	}

	// Validate required parameters
	if params.Prompt == "" {
		return createTextErrorResult("The prompt parameter must be provided"), nil
	}

	// Fill optional parameters with default values
	if params.Model == "" {
		params.Model = define.DefaultT2VModel
	}

	// Build request payload
	payload := map[string]interface{}{
		"model":  params.Model,
		"prompt": params.Prompt,
	}

	// If a first frame image is provided
	if params.FirstFrameImage != "" {
		// Validate image
		if !strings.HasPrefix(params.FirstFrameImage, "http://") &&
			!strings.HasPrefix(params.FirstFrameImage, "https://") &&
			!strings.HasPrefix(params.FirstFrameImage, "data:") {
			// Local image, convert to dataURL
			if _, err := os.Stat(params.FirstFrameImage); os.IsNotExist(err) {
				return createTextErrorResult(fmt.Sprintf("First frame image file does not exist: %s", params.FirstFrameImage)), nil
			}

			// Read image file
			imgData, err := ioutil.ReadFile(params.FirstFrameImage)
			if err != nil {
				return createTextErrorResult(fmt.Sprintf("Failed to read image file: %v", err)), nil
			}

			// Convert to base64
			encoded := base64.StdEncoding.EncodeToString(imgData)
			payload["first_frame_image"] = fmt.Sprintf("data:image/jpeg;base64,%s", encoded)
		} else {
			payload["first_frame_image"] = params.FirstFrameImage
		}
	}

	// Call API to submit video generation task
	response, err := s.Client.Post("/v1/video_generation", payload)
	if err != nil {
		return createTextErrorResult(fmt.Sprintf("Video generation API call failed: %v", err)), nil
	}

	// Get task ID
	taskID, ok := response["task_id"].(string)
	if !ok || taskID == "" {
		return createTextErrorResult("Unable to get task_id from response"), nil
	}

	// Poll task completion status
	var fileID string
	maxRetries := 30    // Up to 10 minutes (30 * 20 seconds)
	retryInterval := 20 // seconds

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Check task status
		statusResponse, err := s.Client.Get(fmt.Sprintf("/v1/query/video_generation?task_id=%s", taskID))
		if err != nil {
			return createTextErrorResult(fmt.Sprintf("Failed to query video generation status: %v", err)), nil
		}

		status, ok := statusResponse["status"].(string)
		if !ok {
			return createTextErrorResult("Unable to get status from response"), nil
		}

		if status == "Fail" {
			return createTextErrorResult(fmt.Sprintf("Video generation failed, task ID: %s", taskID)), nil
		} else if status == "Success" {
			// Get file ID
			fileID, ok = statusResponse["file_id"].(string)
			if !ok || fileID == "" {
				return createTextErrorResult(fmt.Sprintf("Unable to get file_id from success response, task ID: %s", taskID)), nil
			}
			break
		}

		// Still processing, wait and retry
		time.Sleep(time.Duration(retryInterval) * time.Second)
	}

	if fileID == "" {
		return createTextErrorResult(fmt.Sprintf("Timeout getting file_id, task ID: %s", taskID)), nil
	}

	// Get video download URL
	fileResponse, err := s.Client.Get(fmt.Sprintf("/v1/files/retrieve?file_id=%s", fileID))
	if err != nil {
		return createTextErrorResult(fmt.Sprintf("Failed to get video file information: %v", err)), nil
	}

	fileObj, ok := fileResponse["file"].(map[string]interface{})
	if !ok {
		return createTextErrorResult("Unable to get file object from response"), nil
	}

	downloadURL, ok := fileObj["download_url"].(string)
	if !ok || downloadURL == "" {
		return createTextErrorResult(fmt.Sprintf("Unable to get download URL, file ID: %s", fileID)), nil
	}

	// If in URL mode, return URL directly
	if s.ResourceMode == define.ResourceModeURL {
		return createTextResult(fmt.Sprintf("Success. Video URL: %s", downloadURL)), nil
	}

	// Download and save video
	resp, err := http.Get(downloadURL)
	if err != nil {
		return createTextErrorResult(fmt.Sprintf("Failed to download video: %v", err)), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return createTextErrorResult(fmt.Sprintf("Failed to download video, status code: %d", resp.StatusCode)), nil
	}

	videoBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return createTextErrorResult(fmt.Sprintf("Failed to read video data: %v", err)), nil
	}

	// Save video file
	outputPath := storage.BuildOutputPath(params.OutputDirectory, s.BasePath)
	outputFileName := storage.BuildOutputFile("video", taskID, outputPath, "mp4")

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(outputFileName), 0755); err != nil {
		return createTextErrorResult(fmt.Sprintf("Failed to create output directory: %v", err)), nil
	}

	// Write file
	if err := os.WriteFile(outputFileName, videoBytes, 0644); err != nil {
		return createTextErrorResult(fmt.Sprintf("Failed to save video file: %v", err)), nil
	}

	return createTextResult(fmt.Sprintf("Success. Video saved as: %s", outputFileName)), nil
}

// HandleTextToImage processes text-to-image requests
func (s *MinimaxMCPServer) HandleTextToImage(req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	var params TextToImageRequest
	if err := protocol.VerifyAndUnmarshal(req.RawArguments, &params); err != nil {
		return createTextErrorResult(fmt.Sprintf("Parameter parsing failed: %v", err)), nil
	}

	// Validate required parameters
	if params.Prompt == "" {
		return createTextErrorResult("The prompt parameter must be provided"), nil
	}

	// Fill optional parameters with default values
	if params.Model == "" {
		params.Model = define.DefaultT2IModel
	}
	if params.AspectRatio == "" {
		params.AspectRatio = "1:1"
	}
	if params.N == 0 {
		params.N = 1
	}
	if params.ResponseFormat == "" {
		params.ResponseFormat = "base64"
	}

	// Default to true if not specified
	promptOptimizer := true
	if !params.PromptOptimizer {
		promptOptimizer = params.PromptOptimizer
	}

	// Build request payload
	payload := map[string]interface{}{
		"model":            params.Model,
		"prompt":           params.Prompt,
		"aspect_ratio":     params.AspectRatio,
		"n":                params.N,
		"prompt_optimizer": promptOptimizer,
		"response_format":  params.ResponseFormat,
	}

	// Call API
	response, err := s.Client.Post("/v1/image_generation", payload)
	if err != nil {
		return createTextErrorResult(fmt.Sprintf("Image generation API call failed: %v", err)), nil
	}

	// Process response
	data, ok := response["data"].(map[string]interface{})
	if !ok {
		log.Printf("resp : %v", response)
		return createTextErrorResult("Invalid API response format: missing data field"), nil
	}

	// Return different results based on resource mode
	//if s.ResourceMode == define.ResourceModeURL {
	//	return createTextResult(fmt.Sprintf("Success. Image URLs: %v", imageURLs)), nil
	//}

	switch params.ResponseFormat {
	case "base64":
		return textToImageResultWithImageContent(data)
	default:
		imageURLs, ok := data["image_urls"].([]interface{})
		if !ok || len(imageURLs) == 0 {
			return createTextErrorResult("No images generated"), nil
		}
		return createTextResult(fmt.Sprintf("Success. Image URLs: %v", imageURLs)), nil
	}
}

func textToImageResultWithTextContent(outputPath, prompt string, imageURLs []interface{}) (*protocol.CallToolResult, error) {

	// Download and save images
	var outputFileNames []string

	for i, urlInterface := range imageURLs {
		imageURL, ok := urlInterface.(string)
		if !ok {
			continue
		}

		// Create output filename
		truncatedPrompt := prompt
		if len(truncatedPrompt) > 50 {
			truncatedPrompt = truncatedPrompt[:50]
		}
		outputFileName := storage.BuildOutputFile("image", fmt.Sprintf("%d_%s", i, truncatedPrompt), outputPath, "jpeg")

		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(outputFileName), 0755); err != nil {
			return createTextErrorResult(fmt.Sprintf("Failed to create output directory: %v", err)), nil
		}

		// Download image
		resp, err := http.Get(imageURL)
		if err != nil {
			return createTextErrorResult(fmt.Sprintf("Failed to download image: %v", err)), nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return createTextErrorResult(fmt.Sprintf("Failed to download image, status code: %d", resp.StatusCode)), nil
		}

		// Read image data
		imageBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return createTextErrorResult(fmt.Sprintf("Failed to read image data: %v", err)), nil
		}

		// Write file
		if err := os.WriteFile(outputFileName, imageBytes, 0644); err != nil {
			return createTextErrorResult(fmt.Sprintf("Failed to save image file: %v", err)), nil
		}

		outputFileNames = append(outputFileNames, outputFileName)
	}

	return createTextResult(fmt.Sprintf("Success. Images saved as: %v", outputFileNames)), nil
}

func textToImageResultWithImageContent(data map[string]interface{}) (*protocol.CallToolResult, error) {
	imageBase64, ok := data["image_base64"].([]interface{})
	if !ok || len(imageBase64) == 0 {
		return createTextErrorResult("No images generated"), nil
	}

	// 获取第一张图片的base64字符串
	base64Str, ok := imageBase64[0].(string)
	if !ok {
		return createTextErrorResult("Invalid image data format"), nil
	}

	// 解码base64字符串为二进制数据
	imageBytes, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return createTextErrorResult(fmt.Sprintf("Failed to decode base64 image: %v", err)), nil
	}

	return createImageResult(imageBytes), nil
}

// Create text result
func createTextResult(text string) *protocol.CallToolResult {
	return &protocol.CallToolResult{
		Content: []protocol.Content{
			protocol.TextContent{
				Type: "text",
				Text: text,
			},
		},
	}
}

// Create error text result
func createTextErrorResult(text string) *protocol.CallToolResult {
	return &protocol.CallToolResult{
		IsError: true,
		Content: []protocol.Content{
			protocol.TextContent{
				Type: "text",
				Text: text,
			},
		},
	}
}

// Create image result
func createImageResult(data []byte) *protocol.CallToolResult {
	return &protocol.CallToolResult{
		Content: []protocol.Content{
			protocol.ImageContent{
				Type:     "image",
				Data:     data,
				MimeType: "image/jpeg",
			},
		},
	}
}
