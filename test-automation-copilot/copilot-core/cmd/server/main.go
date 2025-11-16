package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourusername/copilot-core/pkg/cloud"
	pkgcontext "github.com/yourusername/copilot-core/pkg/context"
	"github.com/yourusername/copilot-core/pkg/database"
	"github.com/yourusername/copilot-core/pkg/embeddings"
	"github.com/yourusername/copilot-core/pkg/server"
)

func main() {
	// Parse command line flags
	port := flag.Int("port", 8080, "Server port")
	dbPath := flag.String("db", "./data/copilot.db", "Database path")
	vectorPath := flag.String("vector", "./data/vectordb", "Vector database path")
	flag.Parse()

	log.Println("🚀 Starting Test Automation Copilot Server...")

	// Initialize database
	log.Println("📊 Initializing database...")
	db, err := database.NewDB(database.DBConfig{
		Path: *dbPath,
	})
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Initialize schema
	if err := db.Initialize(); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	// Initialize embedding service
	log.Println("🔗 Initializing OpenAI embedding service...")
	embeddingService, err := embeddings.NewEmbeddingService(embeddings.EmbeddingConfig{
		// API key from environment
	})
	if err != nil {
		log.Fatalf("Failed to create embedding service: %v", err)
	}

	// Initialize vector store
	log.Println("💾 Initializing vector database...")
	vectorStore, err := embeddings.NewVectorStore(embeddings.VectorStoreConfig{
		PersistPath:      *vectorPath,
		CollectionName:   "test-copilot",
		EmbeddingService: embeddingService,
	})
	if err != nil {
		log.Fatalf("Failed to create vector store: %v", err)
	}
	defer vectorStore.Close()

	// Initialize semantic search
	log.Println("🔍 Initializing semantic search...")
	semanticSearch, err := embeddings.NewSemanticSearch(embeddings.SemanticSearchConfig{
		VectorStore: vectorStore,
		DB:          db,
	})
	if err != nil {
		log.Fatalf("Failed to create semantic search: %v", err)
	}

	// Initialize cloud client
	log.Println("☁️  Initializing cloud API client...")
	cloudClient, err := cloud.NewCloudClient(cloud.CloudConfig{
		// BaseURL and APIKey will be read from environment variables
		// TESTCOPILOT_CLOUD_URL and TESTCOPILOT_API_KEY
	})
	if err != nil {
		log.Fatalf("Failed to create cloud client: %v", err)
	}
	log.Println("✅ Cloud client configured successfully")

	// Initialize context extractor
	log.Println("🧠 Initializing context extractor...")
	contextExtractor, err := pkgcontext.NewContextExtractor(pkgcontext.ContextExtractorConfig{
		DB:             db,
		SemanticSearch: semanticSearch,
	})
	if err != nil {
		log.Fatalf("Failed to create context extractor: %v", err)
	}

	// Create server
	log.Printf("🌐 Creating HTTP/WebSocket server on port %d...\n", *port)
	srv, err := server.NewServer(server.ServerConfig{
		Port:             *port,
		DB:               db,
		VectorStore:      vectorStore,
		SemanticSearch:   semanticSearch,
		CloudClient:      cloudClient,
		ContextExtractor: contextExtractor,
	})
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\n🛑 Shutting down server...")
		cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
		os.Exit(0)
	}()

	// Start server
	log.Println("✨ Server started successfully!")
	log.Printf("📡 HTTP API: http://localhost:%d", *port)
	log.Printf("🔌 WebSocket: ws://localhost:%d/ws/stream", *port)
	log.Println("\nEndpoints available:")
	log.Println("  GET  /health")
	log.Println("  POST /api/v1/workspace/index")
	log.Println("  GET  /api/v1/workspace/stats")
	log.Println("  POST /api/v1/search/code")
	log.Println("  POST /api/v1/search/pageobjects")
	log.Println("  POST /api/v1/search/tests")
	log.Println("  POST /api/v1/generate/pageobject")
	log.Println("  POST /api/v1/generate/test")
	log.Println("  POST /api/v1/generate/fix")
	log.Println("  POST /api/v1/chat/message")
	log.Println("  GET  /api/v1/db/classes")
	log.Println("  GET  /ws/stream (WebSocket)")
	log.Println("\nPress Ctrl+C to stop")

	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
