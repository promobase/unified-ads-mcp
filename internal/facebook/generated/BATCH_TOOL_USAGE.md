# Facebook Batch Tool Usage Guide

The `facebook_batch` tool allows you to execute multiple Facebook Graph API operations in a single request, providing enhanced error handling and result processing.

## Overview

- **Tool Name**: `facebook_batch`
- **Max Operations**: 50 per batch
- **Supported Methods**: GET, POST, PUT, DELETE
- **Error Handling**: Individual operation failures don't stop the entire batch
- **Response Format**: Structured JSON with success/failure tracking

## Basic Usage

### Simple GET Operations

```json
{
  "tool": "facebook_batch",
  "arguments": {
    "operations": [
      {
        "method": "GET",
        "relative_url": "1234567890?fields=id,name,status",
        "name": "get_campaign"
      },
      {
        "method": "GET", 
        "relative_url": "9876543210?fields=id,name,daily_budget",
        "name": "get_adset"
      }
    ]
  }
}
```

### Mixed Operations (GET + POST)

```json
{
  "tool": "facebook_batch",
  "arguments": {
    "operations": [
      {
        "method": "GET",
        "relative_url": "1234567890?fields=id,name,status",
        "name": "get_campaign_status"
      },
      {
        "method": "POST",
        "relative_url": "1234567890",
        "body": {
          "status": "ACTIVE",
          "name": "Updated Campaign Name"
        },
        "name": "update_campaign"
      },
      {
        "method": "GET",
        "relative_url": "1234567890/insights?date_preset=yesterday&fields=impressions,clicks,spend",
        "name": "get_campaign_insights"
      }
    ]
  }
}
```

### Complex Operations with Nested Objects

```json
{
  "tool": "facebook_batch",
  "arguments": {
    "operations": [
      {
        "method": "POST",
        "relative_url": "act_123456789/campaigns",
        "body": {
          "name": "New Campaign",
          "objective": "CONVERSIONS",
          "status": "PAUSED",
          "daily_budget": 5000,
          "promoted_object": {
            "pixel_id": "123456789012345",
            "custom_event_type": "PURCHASE"
          }
        },
        "name": "create_campaign"
      },
      {
        "method": "POST",
        "relative_url": "act_123456789/adsets",
        "body": {
          "name": "New AdSet",
          "campaign_id": "{result=create_campaign:$.id}",
          "daily_budget": 2500,
          "billing_event": "IMPRESSIONS",
          "optimization_goal": "CONVERSIONS",
          "targeting": {
            "age_min": 25,
            "age_max": 45,
            "geo_locations": {
              "countries": ["US"],
              "location_types": ["home"]
            }
          }
        },
        "name": "create_adset"
      }
    ]
  }
}
```

## Response Format

The tool returns a structured response with the following format:

```json
{
  "total_operations": 3,
  "successful_operations": 2,
  "failed_operations": 1,
  "results": [
    {
      "code": 200,
      "success": true,
      "name": "get_campaign",
      "parsed_body": {
        "id": "1234567890",
        "name": "My Campaign",
        "status": "ACTIVE"
      }
    },
    {
      "code": 400,
      "success": false,
      "name": "invalid_operation",
      "error": "Invalid object ID",
      "parsed_body": {
        "error": {
          "message": "Invalid object ID",
          "type": "OAuthException"
        }
      }
    }
  ],
  "summary": {
    "success_rate": 0.67,
    "total_operations": 3,
    "successful_operations": 2,
    "failed_operations": 1
  }
}
```

## Common Use Cases

### 1. Campaign Performance Analysis

Get campaign data and insights in a single request:

```json
{
  "tool": "facebook_batch",
  "arguments": {
    "operations": [
      {
        "method": "GET",
        "relative_url": "1234567890?fields=id,name,status,daily_budget,lifetime_budget",
        "name": "campaign_details"
      },
      {
        "method": "GET",
        "relative_url": "1234567890/insights?date_preset=last_7d&fields=impressions,clicks,spend,cpm,cpc,ctr",
        "name": "campaign_insights"
      },
      {
        "method": "GET",
        "relative_url": "1234567890/adsets?fields=id,name,status,daily_budget",
        "name": "campaign_adsets"
      }
    ]
  }
}
```

### 2. Bulk Status Updates

Update multiple campaigns/adsets at once:

```json
{
  "tool": "facebook_batch",
  "arguments": {
    "operations": [
      {
        "method": "POST",
        "relative_url": "1234567890",
        "body": {"status": "PAUSED"},
        "name": "pause_campaign_1"
      },
      {
        "method": "POST",
        "relative_url": "1234567891",
        "body": {"status": "PAUSED"},
        "name": "pause_campaign_2"
      },
      {
        "method": "POST",
        "relative_url": "1234567892",
        "body": {"status": "ACTIVE"},
        "name": "activate_campaign_3"
      }
    ]
  }
}
```

### 3. Account-Level Operations

Get comprehensive account information:

```json
{
  "tool": "facebook_batch",
  "arguments": {
    "operations": [
      {
        "method": "GET",
        "relative_url": "act_123456789?fields=id,name,account_status,balance,currency",
        "name": "account_info"
      },
      {
        "method": "GET",
        "relative_url": "act_123456789/campaigns?fields=id,name,status,effective_status&limit=25",
        "name": "active_campaigns"
      },
      {
        "method": "GET",
        "relative_url": "act_123456789/insights?date_preset=yesterday&fields=account_id,spend,impressions,clicks",
        "name": "account_insights"
      }
    ]
  }
}
```

## Error Handling

### Individual Operation Failures

When an operation fails, it doesn't stop the entire batch:

```json
{
  "total_operations": 3,
  "successful_operations": 2,
  "failed_operations": 1,
  "results": [
    {
      "code": 200,
      "success": true,
      "name": "valid_operation"
    },
    {
      "code": 400,
      "success": false,
      "name": "invalid_operation",
      "error": "Invalid parameter",
      "parsed_body": {
        "error": {
          "message": "Invalid parameter",
          "type": "OAuthException",
          "code": 100
        }
      }
    },
    {
      "code": 200,
      "success": true,
      "name": "another_valid_operation"
    }
  ]
}
```

### Validation Errors

The tool validates input before making requests:

- Empty operations array: "At least one operation is required"
- Too many operations: "Maximum 50 operations allowed per batch"
- Missing required fields: Field-specific validation errors

## Best Practices

### 1. Use Descriptive Names

Always provide meaningful names for operations to make results easier to process:

```json
{
  "method": "GET",
  "relative_url": "1234567890",
  "name": "main_campaign_details"  // Good
}
```

Instead of:
```json
{
  "method": "GET", 
  "relative_url": "1234567890"  // No name - harder to identify in results
}
```

### 2. Group Related Operations

Batch related operations together for logical processing:

```json
{
  "operations": [
    // Campaign operations
    {"method": "GET", "relative_url": "campaign_id", "name": "campaign_info"},
    {"method": "GET", "relative_url": "campaign_id/insights", "name": "campaign_insights"},
    
    // AdSet operations  
    {"method": "GET", "relative_url": "adset_id", "name": "adset_info"},
    {"method": "GET", "relative_url": "adset_id/insights", "name": "adset_insights"}
  ]
}
```

### 3. Handle Mixed Success/Failure

Always check the `success` field for each result:

```javascript
const response = await callTool("facebook_batch", {operations: [...]});
const batchResult = JSON.parse(response.content);

batchResult.results.forEach(result => {
  if (result.success) {
    console.log(`✓ ${result.name}: Success`);
    // Process result.parsed_body
  } else {
    console.log(`✗ ${result.name}: ${result.error}`);
    // Handle error
  }
});
```

### 4. Optimize for Rate Limits

Use batch requests to reduce the number of API calls and stay within rate limits:

- Single batch with 10 operations = 1 API call
- 10 individual requests = 10 API calls

### 5. Monitor Success Rates

Use the summary information to monitor batch performance:

```javascript
const {summary} = batchResult;
if (summary.success_rate < 0.8) {
  console.warn(`Low success rate: ${summary.success_rate}`);
  // Investigate failed operations
}
```

## Limitations

1. **Maximum 50 operations** per batch request
2. **Individual rate limits** still apply to each operation
3. **No transaction support** - operations are independent
4. **Dependency handling** requires manual result referencing
5. **File uploads** not supported in batch operations

## Migration from Individual Tools

### Before (Multiple Individual Calls)
```json
// Call 1
{"tool": "campaign_get", "arguments": {"id": "123", "fields": ["name", "status"]}}

// Call 2  
{"tool": "campaign_get_insights", "arguments": {"id": "123", "date_preset": "yesterday"}}

// Call 3
{"tool": "campaign_update", "arguments": {"id": "123", "status": "ACTIVE"}}
```

### After (Single Batch Call)
```json
{
  "tool": "facebook_batch",
  "arguments": {
    "operations": [
      {
        "method": "GET",
        "relative_url": "123?fields=name,status",
        "name": "get_campaign"
      },
      {
        "method": "GET", 
        "relative_url": "123/insights?date_preset=yesterday",
        "name": "get_insights"
      },
      {
        "method": "POST",
        "relative_url": "123",
        "body": {"status": "ACTIVE"},
        "name": "update_status"
      }
    ]
  }
}
```

This reduces 3 API calls to 1 while providing better error handling and result aggregation.