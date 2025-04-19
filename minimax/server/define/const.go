package define

// Constant definitions
const (
	DefaultVoiceID       = "male-qn-qingse"
	DefaultSpeechModel   = "speech-01"
	DefaultSpeed         = 1.0
	DefaultVolume        = 1.0
	DefaultPitch         = 0
	DefaultEmotion       = "happy"
	DefaultSampleRate    = 32000
	DefaultBitrate       = 128000
	DefaultChannel       = 1
	DefaultFormat        = "mp3"
	DefaultLanguageBoost = "auto"
	DefaultT2IModel      = "image-01"
	DefaultT2VModel      = "T2V-01-Director"
	ResourceModeURL      = "url"
	ResourceModeData     = "data"
	DefaultModel         = "MiniMaxAbility"
	DefaultChatModel     = "abab5.5-chat"
)

// Environment variable keys
const (
	EnvMinimaxAPIKey      = "MINIMAX_API_KEY"
	EnvMinimaxAPIHost     = "MINIMAX_API_HOST"
	EnvMinimaxMCPBasePath = "MINIMAX_MCP_BASE_PATH"
	EnvResourceMode       = "RESOURCE_MODE"
)
