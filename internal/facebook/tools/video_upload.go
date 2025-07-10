package tools

import (
	"context"
	"encoding/json"
	"time"

	"unified-ads-mcp/internal/facebook/video"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterVideoUploadTool registers the video upload tool
func RegisterVideoUploadTool(s *server.MCPServer) error {
	tool := mcp.NewTool(
		"facebook_video_upload",
		mcp.WithDescription("Upload a video to a Facebook ad account. Supports large files with chunked upload and optional encoding status monitoring."),
		mcp.WithString("account_id",
			mcp.Required(),
			mcp.Description("The ad account ID (format: act_XXXXXXXXX)"),
			mcp.Pattern("^act_[0-9]+$"),
		),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("Absolute path to the video file to upload"),
		),
		mcp.WithBoolean("wait_for_encoding",
			mcp.Description("Whether to wait for video encoding to complete before returning (default: false)"),
		),
		mcp.WithNumber("encoding_timeout_seconds",
			mcp.Description("Timeout in seconds to wait for encoding (default: 180, min: 30, max: 600)"),
		),
		mcp.WithNumber("polling_interval_seconds",
			mcp.Description("Interval in seconds between encoding status checks (default: 3, min: 1, max: 30)"),
		),
	)

	s.AddTool(tool, handleVideoUpload)
	return nil
}

// VideoUploadArgs defines the arguments for video upload
type VideoUploadArgs struct {
	AccountID               string  `json:"account_id"`
	FilePath                string  `json:"file_path"`
	WaitForEncoding         bool    `json:"wait_for_encoding,omitempty"`
	EncodingTimeoutSeconds  float64 `json:"encoding_timeout_seconds,omitempty"`
	PollingIntervalSeconds  float64 `json:"polling_interval_seconds,omitempty"`
}

func handleVideoUpload(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args VideoUploadArgs
	if err := request.BindArguments(&args); err != nil {
		return mcp.NewToolResultErrorf("Failed to parse arguments: %v", err), nil
	}

	// Set defaults
	if args.EncodingTimeoutSeconds == 0 {
		args.EncodingTimeoutSeconds = 180
	}
	if args.PollingIntervalSeconds == 0 {
		args.PollingIntervalSeconds = 3
	}

	// Create uploader
	uploader := video.NewVideoUploader()

	// Configure options
	var options []video.UploadOption
	if args.EncodingTimeoutSeconds > 0 {
		options = append(options, video.WithTimeout(time.Duration(int(args.EncodingTimeoutSeconds))*time.Second))
	}
	if args.PollingIntervalSeconds > 0 {
		options = append(options, video.WithInterval(time.Duration(int(args.PollingIntervalSeconds))*time.Second))
	}

	// Start upload
	startTime := time.Now()
	result, err := uploader.Upload(ctx, args.AccountID, args.FilePath, args.WaitForEncoding, options...)
	if err != nil {
		return mcp.NewToolResultErrorf("Video upload failed: %v", err), nil
	}

	// Calculate upload duration
	uploadDuration := time.Since(startTime)

	// Build response
	response := map[string]interface{}{
		"success":         true,
		"video_id":        result.VideoID,
		"title":           result.Title,
		"upload_duration": uploadDuration.String(),
		"account_id":      args.AccountID,
	}

	// Add optional fields
	if result.Description != "" {
		response["description"] = result.Description
	}
	if result.Status != "" {
		response["status"] = result.Status
	}
	if args.WaitForEncoding {
		response["encoding_complete"] = true
	}

	// Include any extra metadata
	if len(result.Extra) > 0 {
		response["metadata"] = result.Extra
	}

	// Format response
	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to format response: %v", err), nil
	}

	return mcp.NewToolResultText(string(responseJSON)), nil
}

// RegisterVideoStatusTool registers the video status checking tool
func RegisterVideoStatusTool(s *server.MCPServer) error {
	tool := mcp.NewTool(
		"facebook_video_status",
		mcp.WithDescription("Check the encoding status of an uploaded video"),
		mcp.WithString("video_id",
			mcp.Required(),
			mcp.Description("The video ID to check status for"),
			mcp.Pattern("^[0-9]+$"),
		),
	)

	s.AddTool(tool, handleVideoStatus)
	return nil
}

func handleVideoStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	videoID := request.GetString("video_id", "")
	if videoID == "" {
		return mcp.NewToolResultError("video_id is required"), nil
	}

	// Check video status
	checker := video.NewVideoEncodingStatusChecker()
	status, err := checker.GetStatus(ctx, videoID)
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to get video status: %v", err), nil
	}

	// Build response
	response := map[string]interface{}{
		"video_id": videoID,
		"status":   status.VideoStatus,
		"ready":    status.VideoStatus == "ready",
	}

	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to format response: %v", err), nil
	}

	return mcp.NewToolResultText(string(responseJSON)), nil
}

// VideoUploadBatchTool allows uploading multiple videos
func RegisterVideoUploadBatchTool(s *server.MCPServer) error {
	tool := mcp.NewTool(
		"facebook_video_upload_batch",
		mcp.WithDescription("Upload multiple videos to a Facebook ad account"),
		mcp.WithString("account_id",
			mcp.Required(),
			mcp.Description("The ad account ID (format: act_XXXXXXXXX)"),
			mcp.Pattern("^act_[0-9]+$"),
		),
		mcp.WithArray("file_paths",
			mcp.Required(),
			mcp.Description("Array of absolute paths to video files to upload"),
			mcp.Items(map[string]interface{}{
				"type": "string",
			}),
		),
		mcp.WithBoolean("wait_for_encoding",
			mcp.Description("Whether to wait for all videos to finish encoding (default: false)"),
		),
		mcp.WithBoolean("parallel",
			mcp.Description("Whether to upload videos in parallel (default: false)"),
		),
	)

	s.AddTool(tool, handleVideoUploadBatch)
	return nil
}

type VideoUploadBatchArgs struct {
	AccountID       string   `json:"account_id"`
	FilePaths       []string `json:"file_paths"`
	WaitForEncoding bool     `json:"wait_for_encoding,omitempty"`
	Parallel        bool     `json:"parallel,omitempty"`
}

func handleVideoUploadBatch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args VideoUploadBatchArgs
	if err := request.BindArguments(&args); err != nil {
		return mcp.NewToolResultErrorf("Failed to parse arguments: %v", err), nil
	}

	if len(args.FilePaths) == 0 {
		return mcp.NewToolResultError("At least one file path is required"), nil
	}

	if len(args.FilePaths) > 10 {
		return mcp.NewToolResultError("Maximum 10 videos can be uploaded in a batch"), nil
	}

	results := make([]map[string]interface{}, 0, len(args.FilePaths))
	uploader := video.NewVideoUploader()

	// Sequential upload (parallel upload would need goroutines and proper synchronization)
	for i, filePath := range args.FilePaths {
		uploadResult := map[string]interface{}{
			"index":     i,
			"file_path": filePath,
		}

		result, err := uploader.Upload(ctx, args.AccountID, filePath, args.WaitForEncoding)
		if err != nil {
			uploadResult["success"] = false
			uploadResult["error"] = err.Error()
		} else {
			uploadResult["success"] = true
			uploadResult["video_id"] = result.VideoID
			uploadResult["title"] = result.Title
			if result.Status != "" {
				uploadResult["status"] = result.Status
			}
		}

		results = append(results, uploadResult)
	}

	// Build summary
	successCount := 0
	for _, r := range results {
		if r["success"].(bool) {
			successCount++
		}
	}

	response := map[string]interface{}{
		"total_videos":     len(args.FilePaths),
		"successful":       successCount,
		"failed":           len(args.FilePaths) - successCount,
		"account_id":       args.AccountID,
		"wait_for_encoding": args.WaitForEncoding,
		"results":          results,
	}

	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to format response: %v", err), nil
	}

	return mcp.NewToolResultText(string(responseJSON)), nil
}