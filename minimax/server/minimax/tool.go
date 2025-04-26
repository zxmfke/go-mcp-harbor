package minimax

import (
	"context"
	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/ThinkInAIXYZ/go-mcp/server"
	"log"
)

// TextToAudioRequest 文本转语音请求
type TextToAudioRequest struct {
	Text          string  `json:"text" description:"The text to convert to speech."`
	VoiceID       string  `json:"voice_id,omitempty" description:"Voice ID, e.g., 'male-qn-qingse'/'audiobook_female_1'/'cute_boy', etc."`
	Model         string  `json:"model,omitempty" description:"The model to use. Values range [\"speech-02-hd\"、\"speech-02-turbo\"、\"speech-01-hd\"、\"speech-01-turbo\"、\"speech-01-240228\"、\"speech-01-turbo-240228\"]"`
	Speed         float64 `json:"speed,omitempty" description:"Speech speed, range 0.5 to 2.0, default 1.0."`
	Vol           float64 `json:"vol,omitempty" description:"Volume, range 0 to 10, default 1.0."`
	Pitch         int     `json:"pitch,omitempty" description:"Pitch, range -12 to 12, default 0."`
	Emotion       string  `json:"emotion,omitempty" description:"Emotion, optional values ['happy', 'sad', 'angry', 'fearful', 'disgusted', 'surprised', 'neutral'], default 'happy'."`
	SampleRate    int     `json:"sample_rate,omitempty" description:"Sample rate, optional values [8000,16000,22050,24000,32000,44100], default 16000."`
	Bitrate       int     `json:"bitrate,omitempty" description:"Bitrate, optional values [32000,64000,128000,256000], default 128000."`
	Channel       int     `json:"channel,omitempty" description:"Channel, optional values [1, 2], default 1."`
	Format        string  `json:"format,omitempty" description:"Format, optional values ['pcm', 'mp3','flac'], default 'mp3'."`
	LanguageBoost string  `json:"language_boost,omitempty" description:"Language boost, default 'auto'."`
}

// ListVoicesRequest 列出声音请求
type ListVoicesRequest struct {
	VoiceType string `json:"voice_type,omitempty" description:"The type of voices to list. Values range [\"all\", \"system\", \"voice_cloning\"], with \"all\" being the default."`
}

// VoiceCloneRequest 声音克隆请求，需开通个人或者企业认证
type VoiceCloneRequest struct {
	VoiceID string `json:"voice_id" description:"The id of the voice to use, length more than 8, smaller than 256."`
	File    string `json:"file" description:"The path to the audio file to clone or a URL to the audio file."`
	Text    string `json:"text" description:"The text to use for the demo audio."`
	IsURL   bool   `json:"is_url,omitempty" description:"Whether the file is a URL. Defaults to False."`
	//Model   string `json:"model" description:"The model to use. Values range [\"speech-02-hd\"、\"speech-02-turbo\"、\"speech-01-hd\"、\"speech-01-turbo\"、\"speech-01-240228\"、\"speech-01-turbo-240228\"]"`
}

// GenerateVideoRequest 生成视频请求
type GenerateVideoRequest struct {
	Model           string `json:"model,omitempty" description:"The model to use. Values range [\"T2V-01\", \"T2V-01-Director\", \"I2V-01\", \"I2V-01-Director\", \"I2V-01-live\"]. \"Director\" supports inserting instructions for camera movement control. \"I2V\" for image to video. \"T2V\" for text to video."`
	Prompt          string `json:"prompt" description:"The prompt to generate the video from. When use Director model, the prompt supports 15 Camera Movement Instructions (Enumerated Values)\n            -Truck: [Truck left], [Truck right]\n            -Pan: [Pan left], [Pan right]\n           -Push: [Push in], [Pull out]\n            -Pedestal: [Pedestal up], [Pedestal down]\n            -Tilt: [Tilt up], [Tilt down]\n            -Zoom: [Zoom in], [Zoom out]\n          -Shake: [Shake]\n            -Follow: [Tracking shot]\n            -Static: [Static shot]"`
	FirstFrameImage string `json:"first_frame_image,omitempty" description:"The first frame image. The model must be \"I2V\" Series."`
}

// TextToImageRequest 文本转图像请求
type TextToImageRequest struct {
	Model           string `json:"model,omitempty" description:"The model to use. Values range [\"image-01\"], with \"image-01\" being the default.'"`
	Prompt          string `json:"prompt,omitempty" description:"The prompt to generate the image from."`
	AspectRatio     string `json:"aspect_ratio,omitempty" description:"The aspect ratio of the image. Values range [\"1:1\", \"16:9\",\"4:3\", \"3:2\", \"2:3\", \"3:4\", \"9:16\", \"21:9\"], with \"1:1\" being the default."`
	N               int    `json:"n,omitempty" description:"he number of images to generate. Values range [1, 9], with 1 being the default."`
	PromptOptimizer bool   `json:"prompt_optimizer,omitempty" description:"Whether to optimize the prompt. Values range [True, False], with True being the default."`
	ResponseFormat  string `json:"response_format,omitempty" description:"Used to specify the image response format with base64 or url, default url."`
	OutputDirectory string `json:"output_directory,omitempty" description:"The directory to save the image to, option."`
}

// RegisterTools Register all tools
func RegisterTools(s *server.Server, mcp *MCPServer) {
	// Text-to-speech tool
	textToAudioTool, err := protocol.NewTool(
		"text_to_audio",
		"Convert text to audio with a given voice and save the output audio file to a given directory.\n    Directory is optional, if not provided, the output file will be saved to $HOME/Desktop.\n    Voice id is optional, if not provided, the default voice will be used.\n COST WARNING: This tool makes an API call to Minimax which may incur costs. Only use when explicitly requested by the user.",
		TextToAudioRequest{},
	)
	if err != nil {
		log.Fatalf("Failed to create text_to_audio tool: %v", err)
	}
	s.RegisterTool(textToAudioTool, func(_ context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
		return mcp.HandleTextToAudio(req)
	})

	// List available voices tool
	listVoicesTool, err := protocol.NewTool(
		"list_voices",
		"List all voices available. Only supports when api_host is https://api.minimax.chat",
		ListVoicesRequest{},
	)
	if err != nil {
		log.Fatalf("Failed to create list_voices tool: %v", err)
	}
	s.RegisterTool(listVoicesTool, func(_ context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
		return mcp.HandleListVoices(req)
	})

	// Voice cloning tool
	voiceCloneTool, err := protocol.NewTool(
		"voice_clone",
		"Clone a voice using provided audio files. The new voice will be charged upon first use. COST WARNING: This tool makes an API call to Minimax which may incur costs. Only use when explicitly requested by the user.",
		VoiceCloneRequest{},
	)
	if err != nil {
		log.Fatalf("Failed to create voice_clone tool: %v", err)
	}
	s.RegisterTool(voiceCloneTool, func(_ context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
		return mcp.HandleVoiceClone(req)
	})

	// Generate video tool
	generateVideoTool, err := protocol.NewTool(
		"generate_video",
		"Generate a video from a prompt. COST WARNING: This tool makes an API call to Minimax which may incur costs. Only use when explicitly requested by the user.",
		GenerateVideoRequest{},
	)
	if err != nil {
		log.Fatalf("Failed to create generate_video tool: %v", err)
	}
	s.RegisterTool(generateVideoTool, func(_ context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
		return mcp.HandleGenerateVideo(req)
	})

	// Text-to-image tool
	textToImageTool, err := protocol.NewTool(
		"text_to_image",
		"Generate an image from a prompt. COST WARNING: This tool makes an API call to Minimax which may incur costs. Only use when explicitly requested by the user.",
		TextToImageRequest{},
	)
	if err != nil {
		log.Fatalf("Failed to create text_to_image tool: %v", err)
	}
	s.RegisterTool(textToImageTool, func(_ context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
		return mcp.HandleTextToImage(req)
	})
}
