package v1

// 该文件由 ListModelProvider 接口结果导出（aigw / 2023-08-01 / ListModelProvider）。
// 仅记录 ModelProvider 注册表中的内置 Base Model 元数据，便于 SDK 用户在 CreateModel 时引用。

// BaseModelType 表示 aigw 注册的模型能力类型。
const (
	BaseModelTypeTextGeneration = "text-generation"
	BaseModelTypeEmbeddings     = "embeddings"
	BaseModelTypeVision         = "vision"
	BaseModelTypeAudio          = "audio"
	BaseModelTypeReranking      = "reranking"
)

// BaseModelProvider 表示 aigw 注册的模型供应方。
const (
	BaseModelProviderVolcengine     = "volcengine"
	BaseModelProviderByteplus       = "byteplus"
	BaseModelProviderVolcengineAicc = "volcengine_aicc"
	BaseModelProviderZhipu          = "zhipu"
	BaseModelProviderKimi           = "kimi"
	BaseModelProviderMinimax        = "minimax"
	BaseModelProviderDeepseek       = "deepseek"
	BaseModelProviderAws            = "aws"
	BaseModelProviderAzureOpenai    = "azure_openai"
	BaseModelProviderOpenai         = "openai"
	BaseModelProviderTongyi         = "tongyi"
	BaseModelProviderWenxin         = "wenxin"
	BaseModelProviderGoogle         = "google"
	BaseModelProviderAnthropic      = "anthropic"
	BaseModelProviderLocalai        = "localai"
)

// BaseModel 描述一条 aigw 内置模型条目。
type BaseModel struct {
	Provider  string
	Type      string
	ModelName string
}

// BaseModels 是 aigw 内置模型清单（按 Provider, Type, ModelName 字典序）。
// 数据来源：ListModelProvider 真实集群返回，共 100 条。
var BaseModels = []BaseModel{
	{Provider: "azure_openai", Type: "embeddings", ModelName: "text-embedding-3-large"},
	{Provider: "azure_openai", Type: "embeddings", ModelName: "text-embedding-3-small"},
	{Provider: "azure_openai", Type: "embeddings", ModelName: "text-embedding-ada-002"},
	{Provider: "azure_openai", Type: "text-generation", ModelName: "gpt-4"},
	{Provider: "azure_openai", Type: "text-generation", ModelName: "gpt-4o-mini"},
	{Provider: "azure_openai", Type: "text-generation", ModelName: "o1"},
	{Provider: "azure_openai", Type: "text-generation", ModelName: "o1-preview"},
	{Provider: "byteplus", Type: "audio", ModelName: "seed-tts-1.0"},
	{Provider: "byteplus", Type: "audio", ModelName: "seed-tts-2.0"},
	{Provider: "byteplus", Type: "audio", ModelName: "volc.bigasr.sauc.duration"},
	{Provider: "byteplus", Type: "audio", ModelName: "volc.seedasr.sauc.duration"},
	{Provider: "byteplus", Type: "text-generation", ModelName: "deepseek-v3"},
	{Provider: "byteplus", Type: "text-generation", ModelName: "deepseek-v3-2-251201"},
	{Provider: "byteplus", Type: "text-generation", ModelName: "glm-4-7-251222"},
	{Provider: "byteplus", Type: "text-generation", ModelName: "kimi-k2-thinking-251104"},
	{Provider: "byteplus", Type: "text-generation", ModelName: "seed-1-6-250615"},
	{Provider: "byteplus", Type: "text-generation", ModelName: "seed-1-6-flash-250615"},
	{Provider: "byteplus", Type: "text-generation", ModelName: "seed-1-8-251228"},
	{Provider: "byteplus", Type: "text-generation", ModelName: "seed-2-0-lite-260228"},
	{Provider: "byteplus", Type: "text-generation", ModelName: "seed-2-0-mini-260215"},
	{Provider: "byteplus", Type: "text-generation", ModelName: "seed-2-0-pro-260328"},
	{Provider: "byteplus", Type: "vision", ModelName: "dreamina-seedance-2-0-260128"},
	{Provider: "byteplus", Type: "vision", ModelName: "dreamina-seedance-2-0-fast-260128"},
	{Provider: "byteplus", Type: "vision", ModelName: "seedream-5.0-lite-260128"},
	{Provider: "kimi", Type: "text-generation", ModelName: "kimi-k2.5"},
	{Provider: "minimax", Type: "text-generation", ModelName: "minimax-m2.5"},
	{Provider: "minimax", Type: "text-generation", ModelName: "minimax-m2.7"},
	{Provider: "openai", Type: "embeddings", ModelName: "text-embedding-3-large"},
	{Provider: "openai", Type: "embeddings", ModelName: "text-embedding-3-small"},
	{Provider: "openai", Type: "embeddings", ModelName: "text-embedding-ada-002"},
	{Provider: "openai", Type: "text-generation", ModelName: "gpt-3.5-turbo"},
	{Provider: "openai", Type: "text-generation", ModelName: "gpt-4"},
	{Provider: "openai", Type: "text-generation", ModelName: "gpt-4o"},
	{Provider: "openai", Type: "text-generation", ModelName: "gpt-4o-mini"},
	{Provider: "openai", Type: "text-generation", ModelName: "gpt-5"},
	{Provider: "openai", Type: "text-generation", ModelName: "gpt-5-chat-latest"},
	{Provider: "openai", Type: "text-generation", ModelName: "gpt-5-mini"},
	{Provider: "openai", Type: "text-generation", ModelName: "gpt-5-nano"},
	{Provider: "openai", Type: "text-generation", ModelName: "o1"},
	{Provider: "openai", Type: "text-generation", ModelName: "o1-mini"},
	{Provider: "openai", Type: "text-generation", ModelName: "o1-preview"},
	{Provider: "tongyi", Type: "embeddings", ModelName: "qwen3-vl-embedding"},
	{Provider: "tongyi", Type: "embeddings", ModelName: "text-embedding-v1"},
	{Provider: "tongyi", Type: "embeddings", ModelName: "text-embedding-v2"},
	{Provider: "tongyi", Type: "reranking", ModelName: "qwen3-rerank"},
	{Provider: "tongyi", Type: "text-generation", ModelName: "qwen-plus-latest"},
	{Provider: "tongyi", Type: "text-generation", ModelName: "qwen-turbo-latest"},
	{Provider: "tongyi", Type: "text-generation", ModelName: "qwen3-0.6b"},
	{Provider: "tongyi", Type: "text-generation", ModelName: "qwen3-1.7b"},
	{Provider: "tongyi", Type: "text-generation", ModelName: "qwen3-14b"},
	{Provider: "tongyi", Type: "text-generation", ModelName: "qwen3-235b-a22b"},
	{Provider: "tongyi", Type: "text-generation", ModelName: "qwen3-30b-a3b"},
	{Provider: "tongyi", Type: "text-generation", ModelName: "qwen3-32b"},
	{Provider: "tongyi", Type: "text-generation", ModelName: "qwen3-4b"},
	{Provider: "tongyi", Type: "text-generation", ModelName: "qwen3-8b"},
	{Provider: "volcengine", Type: "audio", ModelName: "seed-tts-2.0"},
	{Provider: "volcengine", Type: "audio", ModelName: "volc.bigasr.auc_turbo"},
	{Provider: "volcengine", Type: "audio", ModelName: "volc.seedasr.sauc.duration"},
	{Provider: "volcengine", Type: "text-generation", ModelName: "deepseek-v3-2-251201"},
	{Provider: "volcengine", Type: "text-generation", ModelName: "doubao-1-5-lite"},
	{Provider: "volcengine", Type: "text-generation", ModelName: "doubao-seed-2-0-code-preview-260215"},
	{Provider: "volcengine", Type: "text-generation", ModelName: "doubao-seed-2-0-lite-260215"},
	{Provider: "volcengine", Type: "text-generation", ModelName: "doubao-seed-2-0-mini-260215"},
	{Provider: "volcengine", Type: "text-generation", ModelName: "doubao-seed-2-0-pro-260215"},
	{Provider: "volcengine", Type: "text-generation", ModelName: "glm-4-7-251222"},
	{Provider: "volcengine", Type: "vision", ModelName: "doubao-seedance-2-0-260128"},
	{Provider: "volcengine", Type: "vision", ModelName: "doubao-seedance-2-0-fast-260128"},
	{Provider: "volcengine_aicc", Type: "text-generation", ModelName: "deepseek-v3-2-251201"},
	{Provider: "volcengine_aicc", Type: "text-generation", ModelName: "doubao-seed-1-6-250615"},
	{Provider: "volcengine_aicc", Type: "text-generation", ModelName: "doubao-seed-2-0-lite-260215"},
	{Provider: "volcengine_aicc", Type: "text-generation", ModelName: "doubao-seed-2-0-pro-260215"},
	{Provider: "volcengine_aicc", Type: "text-generation", ModelName: "glm-4-7-251222"},
	{Provider: "wenxin", Type: "embeddings", ModelName: "bge-large-zh"},
	{Provider: "wenxin", Type: "text-generation", ModelName: "AquilaChat-7B"},
	{Provider: "wenxin", Type: "text-generation", ModelName: "BLOOMZ-7B"},
	{Provider: "wenxin", Type: "text-generation", ModelName: "ChatGLM2-6B-32K"},
	{Provider: "wenxin", Type: "text-generation", ModelName: "ERNIE 3.5"},
	{Provider: "wenxin", Type: "text-generation", ModelName: "ERNIE Speed"},
	{Provider: "wenxin", Type: "text-generation", ModelName: "ERNIE-3.5-8K-0205"},
	{Provider: "wenxin", Type: "text-generation", ModelName: "ERNIE-3.5-8K-1222"},
	{Provider: "wenxin", Type: "text-generation", ModelName: "ERNIE-4.0-8K"},
	{Provider: "wenxin", Type: "text-generation", ModelName: "ERNIE-Bot"},
	{Provider: "wenxin", Type: "text-generation", ModelName: "ERNIE-Bot-4"},
	{Provider: "wenxin", Type: "text-generation", ModelName: "ERNIE-Bot-8k"},
	{Provider: "wenxin", Type: "text-generation", ModelName: "ERNIE-Lite-8K-0308"},
	{Provider: "wenxin", Type: "text-generation", ModelName: "ERNIE-Speed"},
	{Provider: "wenxin", Type: "text-generation", ModelName: "ERNIE-Speed-128k"},
	{Provider: "wenxin", Type: "text-generation", ModelName: "Llama-2-13b-chat"},
	{Provider: "wenxin", Type: "text-generation", ModelName: "Llama-2-7b-chat"},
	{Provider: "wenxin", Type: "text-generation", ModelName: "Mixtral-8x7B-Instruct"},
	{Provider: "wenxin", Type: "text-generation", ModelName: "Qianfan-BLOOMZ-7B-compressed"},
	{Provider: "wenxin", Type: "text-generation", ModelName: "Qianfan-Chinese-Llama-2-13B"},
	{Provider: "wenxin", Type: "text-generation", ModelName: "Qianfan-Chinese-Llama-2-7B"},
	{Provider: "wenxin", Type: "text-generation", ModelName: "XuanYuan-70B-Chat-4bit"},
	{Provider: "wenxin", Type: "text-generation", ModelName: "Yi-34B-Chat"},
	{Provider: "zhipu", Type: "text-generation", ModelName: "glm-3-turbo"},
	{Provider: "zhipu", Type: "text-generation", ModelName: "glm-4"},
	{Provider: "zhipu", Type: "text-generation", ModelName: "glm-4v"},
	{Provider: "zhipu", Type: "text-generation", ModelName: "glm-5"},
	{Provider: "zhipu", Type: "text-generation", ModelName: "glm-5.1"},
}
