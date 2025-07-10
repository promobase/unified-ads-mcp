# Facebook Video Upload Examples

The Facebook MCP server now includes video upload capabilities that support:
- Large file uploads with automatic chunking
- Progress tracking through sessions
- Encoding status monitoring
- Batch video uploads

## Prerequisites

1. Set your Facebook access token:
   ```bash
   export FACEBOOK_ACCESS_TOKEN="your_access_token"
   ```

2. Ensure you have the necessary permissions for video uploads to the ad account.

## Available Tools

### 1. facebook_video_upload

Upload a single video to an ad account.

**Example:**
```json
{
  "tool": "facebook_video_upload",
  "arguments": {
    "account_id": "act_123456789",
    "file_path": "/path/to/video.mp4",
    "wait_for_encoding": true,
    "encoding_timeout_seconds": 300,
    "polling_interval_seconds": 5
  }
}
```

**Response:**
```json
{
  "success": true,
  "video_id": "987654321",
  "title": "video.mp4",
  "upload_duration": "45.2s",
  "account_id": "act_123456789",
  "encoding_complete": true
}
```

### 2. facebook_video_status

Check the encoding status of an uploaded video.

**Example:**
```json
{
  "tool": "facebook_video_status",
  "arguments": {
    "video_id": "987654321"
  }
}
```

**Response:**
```json
{
  "video_id": "987654321",
  "status": "ready",
  "ready": true
}
```

### 3. facebook_video_upload_batch

Upload multiple videos in a batch.

**Example:**
```json
{
  "tool": "facebook_video_upload_batch",
  "arguments": {
    "account_id": "act_123456789",
    "file_paths": [
      "/path/to/video1.mp4",
      "/path/to/video2.mp4",
      "/path/to/video3.mp4"
    ],
    "wait_for_encoding": false
  }
}
```

**Response:**
```json
{
  "total_videos": 3,
  "successful": 3,
  "failed": 0,
  "account_id": "act_123456789",
  "wait_for_encoding": false,
  "results": [
    {
      "index": 0,
      "file_path": "/path/to/video1.mp4",
      "success": true,
      "video_id": "111111111",
      "title": "video1.mp4"
    },
    {
      "index": 1,
      "file_path": "/path/to/video2.mp4",
      "success": true,
      "video_id": "222222222",
      "title": "video2.mp4"
    },
    {
      "index": 2,
      "file_path": "/path/to/video3.mp4",
      "success": true,
      "video_id": "333333333",
      "title": "video3.mp4"
    }
  ]
}
```

## Loading Video Tools

You can load video tools in two ways:

1. **Load the video scope:**
   ```json
   {
     "tool": "tool_manager",
     "arguments": {
       "action": "set",
       "scopes": ["video"]
     }
   }
   ```

2. **Load the creative scope (includes video tools):**
   ```json
   {
     "tool": "tool_manager",
     "arguments": {
       "action": "set",
       "scopes": ["creative"]
     }
   }
   ```

## Video Upload Process

The video upload implementation follows Facebook's resumable upload protocol:

1. **Start Phase**: Initiates upload session and gets session ID
2. **Transfer Phase**: Uploads file in chunks (handles retries for network issues)
3. **Finish Phase**: Completes the upload and gets video ID
4. **Encoding Phase** (optional): Waits for video processing to complete

## Error Handling

The uploader handles several error scenarios:
- Network interruptions (automatic retry)
- Transient API errors (automatic retry with backoff)
- File access errors
- Encoding failures
- Timeout conditions

## Best Practices

1. **File Size**: The uploader automatically handles large files by chunking
2. **Encoding Wait**: Only wait for encoding if you need to use the video immediately
3. **Batch Uploads**: Use batch upload for multiple videos to save time
4. **Error Recovery**: The chunked upload supports resume on failure

## Limitations

- Maximum 10 videos per batch upload
- Encoding timeout default is 180 seconds (can be increased to 600)
- Requires appropriate Facebook permissions for the ad account