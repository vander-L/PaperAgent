package config

// Config 存储从 config.yaml 和 .env 加载的应用配置。
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	LLM    LLMConfig    `mapstructure:"llm"`
	Milvus MilvusConfig `mapstructure:"milvus"`
	MCP    MCPConfig    `mapstructure:"mcp"`
}

type ServerConfig struct {
	Address string `mapstructure:"address"`
}

type LLMConfig struct {
	Chat     LLMProviderConfig `mapstructure:"chat"`
	Embedder EmbedderConfig    `mapstructure:"embedder"`
}

type LLMProviderConfig struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
	Model   string `mapstructure:"model"`
}

type EmbedderConfig struct {
	APIKey    string `mapstructure:"api_key"`
	BaseURL   string `mapstructure:"base_url"`
	Model     string `mapstructure:"model"`
	Dimension int    `mapstructure:"dimension"`
}

type MilvusConfig struct {
	Address    string `mapstructure:"address"`
	Database   string `mapstructure:"database"`
	Collection string `mapstructure:"collection"`
	Dimension  int    `mapstructure:"dimension"`
}

type MCPConfig struct {
	Url string `mapstructure:"url"`
}
