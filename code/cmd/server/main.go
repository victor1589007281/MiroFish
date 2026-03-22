package main

import (
	"log"
	"os"

	"swarm-predict/internal/api"
	"swarm-predict/internal/graph"
	"swarm-predict/internal/openclaw"
	"swarm-predict/internal/store"
)

func main() {
	gatewayURL := envOrDefault("OPENCLAW_GATEWAY_URL", "http://localhost:18789")
	apiKey := envOrDefault("OPENCLAW_API_KEY", "")
	llmModel := envOrDefault("LLM_MODEL", "dashscope:qwen-plus")
	flashModel := envOrDefault("LLM_FLASH_MODEL", "dashscope:qwen-flash")
	listenAddr := envOrDefault("LISTEN_ADDR", ":8080")

	var chatClient openclaw.ChatCompleter
	var spawner openclaw.Spawner

	if apiKey != "" {
		chatClient = openclaw.NewHTTPChatClient(gatewayURL, apiKey)
		spawner = openclaw.NewHTTPSpawnClient(gatewayURL, apiKey)
	} else {
		log.Println("WARNING: No OPENCLAW_API_KEY set, using mock clients")
		chatClient = openclaw.NewMockChatClient()
		spawner = openclaw.NewMockSpawnClient()
	}

	graphClient := graph.NewMockClient()
	projectStore := store.NewInMemoryStore()

	deps := api.Deps{
		ChatClient: chatClient,
		Spawner:    spawner,
		Graph:      graphClient,
		Store:      projectStore,
		Model:      llmModel,
		FlashModel: flashModel,
	}

	router := api.NewRouter(deps)
	log.Printf("Starting server on %s", listenAddr)
	if err := router.Run(listenAddr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
