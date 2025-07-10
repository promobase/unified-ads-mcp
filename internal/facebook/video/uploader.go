package video

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

// VideoUploader handles video uploads to Facebook ad accounts
type VideoUploader struct {
	session *VideoUploadSession
}

// NewVideoUploader creates a new video uploader instance
func NewVideoUploader() *VideoUploader {
	return &VideoUploader{}
}

// Upload uploads a video file to the specified ad account
func (vu *VideoUploader) Upload(ctx context.Context, accountID string, filePath string, waitForEncoding bool, options ...UploadOption) (*VideoUploadResult, error) {
	// Check there is no existing session
	if vu.session != nil {
		return nil, fmt.Errorf("there is already an upload session in progress")
	}

	// Apply options
	opts := &uploadOptions{
		interval: 3 * time.Second,
		timeout:  180 * time.Second,
	}
	for _, opt := range options {
		opt(opts)
	}

	// Create and start session
	vu.session = NewVideoUploadSession(accountID, filePath, waitForEncoding, opts)
	result, err := vu.session.Start(ctx)
	vu.session = nil
	return result, err
}

// UploadOption configures video upload behavior
type UploadOption func(*uploadOptions)

type uploadOptions struct {
	interval time.Duration
	timeout  time.Duration
}

// WithInterval sets the polling interval for encoding status
func WithInterval(interval time.Duration) UploadOption {
	return func(o *uploadOptions) {
		o.interval = interval
	}
}

// WithTimeout sets the timeout for encoding completion
func WithTimeout(timeout time.Duration) UploadOption {
	return func(o *uploadOptions) {
		o.timeout = timeout
	}
}

// VideoUploadSession manages a single video upload session
type VideoUploadSession struct {
	accountID       string
	filePath        string
	waitForEncoding bool
	interval        time.Duration
	timeout         time.Duration
	sessionID       string
	startOffset     int64
	endOffset       int64
	videoID         string
}

// NewVideoUploadSession creates a new upload session
func NewVideoUploadSession(accountID, filePath string, waitForEncoding bool, opts *uploadOptions) *VideoUploadSession {
	return &VideoUploadSession{
		accountID:       accountID,
		filePath:        filePath,
		waitForEncoding: waitForEncoding,
		interval:        opts.interval,
		timeout:         opts.timeout,
	}
}

// VideoUploadResult contains the result of a video upload
type VideoUploadResult struct {
	VideoID     string                 `json:"id"`
	Title       string                 `json:"title,omitempty"`
	Description string                 `json:"description,omitempty"`
	Status      string                 `json:"status,omitempty"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

// Start begins the upload process
func (s *VideoUploadSession) Start(ctx context.Context) (*VideoUploadResult, error) {
	// Get file info
	fileInfo, err := os.Stat(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Start upload session
	startResp, err := s.startUpload(ctx, fileInfo.Size())
	if err != nil {
		return nil, fmt.Errorf("failed to start upload: %w", err)
	}

	s.startOffset = startResp.StartOffset
	s.endOffset = startResp.EndOffset
	s.sessionID = startResp.SessionID
	s.videoID = startResp.VideoID

	// Transfer file chunks
	if err := s.transferChunks(ctx); err != nil {
		return nil, fmt.Errorf("failed to transfer chunks: %w", err)
	}

	// Finish upload
	finishResp, err := s.finishUpload(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to finish upload: %w", err)
	}

	// Wait for encoding if requested
	if s.waitForEncoding {
		if err := s.waitForVideoEncoding(ctx); err != nil {
			return nil, fmt.Errorf("video encoding failed: %w", err)
		}
	}

	// Build result
	result := &VideoUploadResult{
		VideoID:     s.videoID,
		Title:       filepath.Base(s.filePath),
		Description: finishResp.Description,
		Extra:       finishResp.Extra,
	}

	return result, nil
}

// startUploadResponse represents the response from start upload API
type startUploadResponse struct {
	SessionID   string `json:"upload_session_id"`
	VideoID     string `json:"video_id"`
	StartOffset int64  `json:"start_offset"`
	EndOffset   int64  `json:"end_offset"`
}

// startUpload initiates the upload session
func (s *VideoUploadSession) startUpload(ctx context.Context, fileSize int64) (*startUploadResponse, error) {
	params := map[string]string{
		"file_size":    fmt.Sprintf("%d", fileSize),
		"upload_phase": "start",
	}

	resp, err := makeGraphAPIRequest(ctx, "POST", fmt.Sprintf("%s/advideos", s.accountID), params, nil)
	if err != nil {
		return nil, err
	}

	var result startUploadResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse start response: %w", err)
	}

	return &result, nil
}

// transferChunks uploads the video file in chunks
func (s *VideoUploadSession) transferChunks(ctx context.Context) error {
	file, err := os.Open(s.filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Calculate retry count: at least 2, or 1 retry per 10MB
	retryCount := max(2, int(fileInfo.Size()/(10*1024*1024)))

	for s.startOffset != s.endOffset {
		// Seek to start offset
		if _, err := file.Seek(s.startOffset, 0); err != nil {
			return fmt.Errorf("failed to seek file: %w", err)
		}

		// Read chunk
		chunkSize := s.endOffset - s.startOffset
		chunk := make([]byte, chunkSize)
		if _, err := io.ReadFull(file, chunk); err != nil {
			return fmt.Errorf("failed to read chunk: %w", err)
		}

		// Upload chunk with retry
		err := s.uploadChunkWithRetry(ctx, chunk, retryCount)
		if err != nil {
			return err
		}
	}

	return nil
}

// uploadChunkWithRetry uploads a chunk with retry logic
func (s *VideoUploadSession) uploadChunkWithRetry(ctx context.Context, chunk []byte, maxRetries int) error {
	retries := maxRetries

	for {
		resp, err := s.uploadChunk(ctx, chunk)
		if err == nil {
			s.startOffset = resp.StartOffset
			s.endOffset = resp.EndOffset
			return nil
		}

		// Check if error is retryable
		if fbErr, ok := err.(*FacebookError); ok {
			// Handle specific error subcodes
			if fbErr.ErrorSubcode == 1363037 && retries > 0 {
				// Resume from error data offsets
				if fbErr.ErrorData != nil {
					if startOffset, ok := fbErr.ErrorData["start_offset"].(float64); ok {
						s.startOffset = int64(startOffset)
					}
					if endOffset, ok := fbErr.ErrorData["end_offset"].(float64); ok {
						s.endOffset = int64(endOffset)
					}
				}
				retries--
				continue
			}

			// Handle transient errors
			if fbErr.IsTransient && retries > 0 {
				time.Sleep(1 * time.Second)
				retries--
				continue
			}
		}

		return fmt.Errorf("failed to upload chunk: %w", err)
	}
}

// chunkUploadResponse represents the response from chunk upload
type chunkUploadResponse struct {
	StartOffset int64 `json:"start_offset"`
	EndOffset   int64 `json:"end_offset"`
}

// uploadChunk uploads a single chunk
func (s *VideoUploadSession) uploadChunk(ctx context.Context, chunk []byte) (*chunkUploadResponse, error) {
	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add form fields
	fields := map[string]string{
		"upload_phase":      "transfer",
		"start_offset":      fmt.Sprintf("%d", s.startOffset),
		"upload_session_id": s.sessionID,
	}

	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, fmt.Errorf("failed to write field %s: %w", key, err)
		}
	}

	// Add file chunk
	part, err := writer.CreateFormFile("video_file_chunk", filepath.Base(s.filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := part.Write(chunk); err != nil {
		return nil, fmt.Errorf("failed to write chunk: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Make request
	resp, err := makeGraphVideoAPIRequest(ctx, "POST", fmt.Sprintf("%s/advideos", s.accountID), &buf, writer.FormDataContentType())
	if err != nil {
		return nil, err
	}

	var result chunkUploadResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse chunk response: %w", err)
	}

	return &result, nil
}

// finishUploadResponse represents the response from finish upload API
type finishUploadResponse struct {
	Success     bool                   `json:"success"`
	Description string                 `json:"description,omitempty"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

// finishUpload completes the upload session
func (s *VideoUploadSession) finishUpload(ctx context.Context) (*finishUploadResponse, error) {
	params := map[string]string{
		"upload_phase":      "finish",
		"upload_session_id": s.sessionID,
		"title":             filepath.Base(s.filePath),
	}

	resp, err := makeGraphAPIRequest(ctx, "POST", fmt.Sprintf("%s/advideos", s.accountID), params, nil)
	if err != nil {
		return nil, err
	}

	var result finishUploadResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse finish response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("upload finish failed")
	}

	return &result, nil
}

// waitForVideoEncoding waits for video encoding to complete
func (s *VideoUploadSession) waitForVideoEncoding(ctx context.Context) error {
	checker := NewVideoEncodingStatusChecker()
	return checker.WaitUntilReady(ctx, s.videoID, s.interval, s.timeout)
}

// VideoEncodingStatusChecker checks video encoding status
type VideoEncodingStatusChecker struct{}

// NewVideoEncodingStatusChecker creates a new status checker
func NewVideoEncodingStatusChecker() *VideoEncodingStatusChecker {
	return &VideoEncodingStatusChecker{}
}

// WaitUntilReady waits until video encoding is complete
func (c *VideoEncodingStatusChecker) WaitUntilReady(ctx context.Context, videoID string, interval, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		status, err := c.GetStatus(ctx, videoID)
		if err != nil {
			return fmt.Errorf("failed to get video status: %w", err)
		}

		if status.VideoStatus != "processing" {
			if status.VideoStatus == "ready" {
				return nil
			}
			return fmt.Errorf("video encoding failed with status: %s", status.VideoStatus)
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("video encoding timeout after %v", timeout)
		}

		time.Sleep(interval)
	}
}

// VideoStatus represents video encoding status
type VideoStatus struct {
	VideoStatus string `json:"video_status"`
}

// StatusResponse wraps the status response
type StatusResponse struct {
	Status VideoStatus `json:"status"`
}

// GetStatus retrieves current video encoding status
func (c *VideoEncodingStatusChecker) GetStatus(ctx context.Context, videoID string) (*VideoStatus, error) {
	params := map[string]string{
		"fields": "status",
	}

	resp, err := makeGraphAPIRequest(ctx, "GET", videoID, params, nil)
	if err != nil {
		return nil, err
	}

	var result StatusResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse status response: %w", err)
	}

	return &result.Status, nil
}

// Helper function to get max of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}