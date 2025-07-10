package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"unified-ads-mcp/internal/facebook/tools"
	"unified-ads-mcp/internal/utils"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ProductionLogger wraps the standard logger with structured logging
type ProductionLogger struct {
	logger *log.Logger
	mu     sync.Mutex
}

func NewProductionLogger(debugMode bool) *ProductionLogger {
	flags := log.Ldate | log.Ltime | log.Lmicroseconds
	if debugMode {
		flags |= log.Lshortfile
	}
	return &ProductionLogger{
		logger: log.New(os.Stdout, "[FB-MCP] ", flags),
	}
}

func (l *ProductionLogger) Info(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Printf("[INFO] "+format, args...)
}

func (l *ProductionLogger) Debug(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Printf("[DEBUG] "+format, args...)
}

func (l *ProductionLogger) Error(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Printf("[ERROR] "+format, args...)
}

func (l *ProductionLogger) Warn(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Printf("[WARN] "+format, args...)
}

// RequestTracker tracks request timing and provides logging
type RequestTracker struct {
	logger     *ProductionLogger
	startTimes sync.Map
	debugMode  bool
}

func NewRequestTracker(logger *ProductionLogger, debugMode bool) *RequestTracker {
	return &RequestTracker{
		logger:    logger,
		debugMode: debugMode,
	}
}

func (rt *RequestTracker) trackStart(id any) {
	rt.startTimes.Store(id, time.Now())
}

func (rt *RequestTracker) trackEnd(id any, method string) {
	if startTime, ok := rt.startTimes.LoadAndDelete(id); ok {
		duration := time.Since(startTime.(time.Time))
		rt.logger.Info("%s completed in %v", method, duration)
	}
}

func (rt *RequestTracker) getDuration(id any) time.Duration {
	if startTime, ok := rt.startTimes.Load(id); ok {
		return time.Since(startTime.(time.Time))
	}
	return 0
}

// CreateProductionHooks creates hooks with proper logging and monitoring
func CreateProductionHooks(logger *ProductionLogger, tracker *RequestTracker, debugMode bool) *server.Hooks {
	hooks := &server.Hooks{}

	// Log before method execution
	hooks.AddBeforeAny(func(ctx context.Context, id any, method mcp.MCPMethod, message any) {
		tracker.trackStart(id)

		switch method {
		case "initialize":
			logger.Info("Client initializing connection (request_id=%v)", id)
		case "tools/list":
			logger.Info("Client requesting tool list (request_id=%v)", id)
		case "tools/call":
			msgJSON, _ := json.Marshal(message)
			var callReq struct {
				Params struct {
					Name      string                 `json:"name"`
					Arguments map[string]interface{} `json:"arguments"`
				} `json:"params"`
			}
			if err := json.Unmarshal(msgJSON, &callReq); err == nil && callReq.Params.Name != "" {
				// Special handling for tool_manager for cleaner logs
				if callReq.Params.Name == "tool_manager" {
					if action, ok := callReq.Params.Arguments["action"].(string); ok {
						switch action {
						case "get":
							logger.Info("Tool call: tool_manager - listing available scopes (request_id=%v)", id)
						case "set":
							if scopes, ok := callReq.Params.Arguments["scopes"].([]interface{}); ok {
								logger.Info("Tool call: tool_manager - setting scopes %v (request_id=%v)", scopes, id)
							}
						default:
							logger.Info("Tool call: tool_manager - action '%s' (request_id=%v)", action, id)
						}
					}
				} else {
					logger.Info("Tool call: %s (request_id=%v, args=%v)",
						callReq.Params.Name, id, callReq.Params.Arguments)
				}
			} else if debugMode {
				logger.Debug("Tool call raw (request_id=%v): %s", id, string(msgJSON))
			}
		default:
			if debugMode {
				msgJSON, _ := json.Marshal(message)
				logger.Debug("Method %s called (request_id=%v): %s", method, id, string(msgJSON))
			}
		}
	})

	// Log successful tool calls
	hooks.AddAfterCallTool(func(ctx context.Context, id any, request *mcp.CallToolRequest, result *mcp.CallToolResult) {
		duration := tracker.getDuration(id)
		toolName := ""
		if request != nil {
			toolName = request.Params.Name
		}

		if result != nil && result.IsError {
			logger.Warn("Tool '%s' returned error after %v (request_id=%v)", toolName, duration, id)
		} else {
			logger.Info("Tool '%s' completed successfully in %v (request_id=%v)", toolName, duration, id)
		}
		tracker.startTimes.Delete(id)
	})

	// Log successful completions for other methods
	hooks.AddOnSuccess(func(ctx context.Context, id any, method mcp.MCPMethod, message any, result any) {
		// Skip if already logged by specific handler
		if method == "tools/call" {
			return
		}
		tracker.trackEnd(id, string(method))
	})

	// Log errors with details
	hooks.AddOnError(func(ctx context.Context, id any, method mcp.MCPMethod, message any, err error) {
		duration := tracker.getDuration(id)
		msgJSON, _ := json.Marshal(message)
		logger.Error("Method %s failed after %v (request_id=%v): %v | Request: %s",
			method, duration, id, err, string(msgJSON))
		tracker.startTimes.Delete(id)
	})

	return hooks
}

// Note: RecoveryMiddleware is not used directly as mcp-go has built-in recovery
// This is kept for reference on how to implement middleware if needed

func NewFacebookMCPServer(logger *ProductionLogger, debugMode bool) *server.MCPServer {
	// Create request tracker
	tracker := NewRequestTracker(logger, debugMode)

	// Create production hooks
	hooks := CreateProductionHooks(logger, tracker, debugMode)

	// Create the MCP server with production configuration
	mcpServer := server.NewMCPServer(
		"facebook-business-mcp",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithLogging(),
		server.WithHooks(hooks),
		server.WithRecovery(), // Enable panic recovery
	)

	// Register tools with logging
	logger.Info("Registering tools...")

	if err := tools.RegisterToolManagerTool(mcpServer); err != nil {
		logger.Error("Failed to register tool manager: %v", err)
		// In production, we might want to continue but log the error
	} else {
		logger.Info("✓ Tool manager registered")
	}

	if err := tools.RegisterAccountTools(mcpServer); err != nil {
		logger.Error("Failed to register account tools: %v", err)
	} else {
		logger.Info("✓ Account tools registered")
	}

	if err := tools.RegisterBatchTools(mcpServer); err != nil {
		logger.Error("Failed to register batch tools: %v", err)
	} else {
		logger.Info("✓ Batch tools registered")
	}

	// Register video tools
	if err := tools.RegisterVideoUploadTool(mcpServer); err != nil {
		logger.Error("Failed to register video upload tool: %v", err)
	} else {
		logger.Info("✓ Video upload tool registered")
	}

	if err := tools.RegisterVideoStatusTool(mcpServer); err != nil {
		logger.Error("Failed to register video status tool: %v", err)
	} else {
		logger.Info("✓ Video status tool registered")
	}

	if err := tools.RegisterVideoUploadBatchTool(mcpServer); err != nil {
		logger.Error("Failed to register video batch upload tool: %v", err)
	} else {
		logger.Info("✓ Video batch upload tool registered")
	}

	// Add health check tool
	healthCheckTool := mcp.NewTool(
		"health_check",
		mcp.WithDescription("Check server health status"),
	)
	mcpServer.AddTool(healthCheckTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		status := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   "1.0.0",
			"pid":       os.Getpid(),
			// Note: Tool count would require tracking during registration
			"uptime": time.Since(serverStartTime).String(),
		}
		statusJSON, _ := json.Marshal(status)
		return mcp.NewToolResultText(string(statusJSON)), nil
	})
	logger.Info("✓ Health check registered")

	logger.Info("All tools registered successfully")

	return mcpServer
}

var serverStartTime = time.Now()

func startWithGracefulShutdown(mcpServer *server.MCPServer, transport string, logger *ProductionLogger) {
	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Channel to signal server completion
	serverDone := make(chan error, 1)

	// Start server in goroutine
	go func() {
		var err error
		if transport == "http" {
			httpServer := server.NewStreamableHTTPServer(mcpServer)
			logger.Info("Starting HTTP server on :8080/mcp")
			err = httpServer.Start(":8080")
		} else {
			logger.Info("Starting STDIO server")
			err = server.ServeStdio(mcpServer)
		}
		serverDone <- err
	}()

	logger.Info("Server started successfully. Press Ctrl+C to stop.")

	// Wait for shutdown signal or server error
	select {
	case sig := <-sigChan:
		logger.Info("Received signal: %v", sig)
		logger.Info("Initiating graceful shutdown...")
	case err := <-serverDone:
		if err != nil {
			logger.Error("Server error: %v", err)
		}
		return
	}

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Info("Waiting for ongoing requests to complete...")

	// Give ongoing requests time to complete
	select {
	case <-shutdownCtx.Done():
		logger.Warn("Shutdown timeout exceeded, forcing stop")
	case <-time.After(2 * time.Second):
		logger.Info("Graceful shutdown completed")
	}

	logger.Info("Server stopped. Uptime: %v", time.Since(serverStartTime))
}

func main() {
	var (
		transport string
		debugMode bool
		version   bool
	)

	flag.StringVar(&transport, "t", "stdio", "Transport type (stdio or http)")
	flag.StringVar(&transport, "transport", "stdio", "Transport type (stdio or http)")
	flag.BoolVar(&debugMode, "debug", false, "Enable debug logging")
	flag.BoolVar(&version, "version", false, "Show version information")
	flag.Parse()

	if version {
		fmt.Println("Facebook Business MCP Server v1.0.0")
		os.Exit(0)
	}

	// Initialize production logger
	logger := NewProductionLogger(debugMode)

	// Print startup banner
	logger.Info("========================================")
	logger.Info("Facebook Business MCP Server")
	logger.Info("Version: 0.0.1")
	logger.Info("Transport: %s", transport)
	logger.Info("Debug mode: %v", debugMode)
	logger.Info("PID: %d", os.Getpid())
	logger.Info("========================================")

	// Load configuration
	logger.Info("Loading Facebook configuration...")
	utils.LoadFacebookConfig()

	// Create and start the server
	mcpServer := NewFacebookMCPServer(logger, debugMode)

	// Start with graceful shutdown
	startWithGracefulShutdown(mcpServer, transport, logger)
}
