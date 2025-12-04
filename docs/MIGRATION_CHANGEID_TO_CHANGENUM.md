# Migration: ChangeID to ChangeNum

## Overview
This document describes the migration from using `changeId` (Gerrit Change-Id) to `changeNum` (Gerrit change number) as the primary identifier for patches.

## Problem
Gerrit allows multiple changes to have the same Change-ID (e.g., when cherry-picking across branches). Using Change-ID as a unique identifier can lead to ambiguity and incorrect patch lookups.

## Solution
Use the Gerrit change number (`_number` field in API responses), which is guaranteed to be unique across the entire Gerrit instance.

## Changes Made

### 1. API Layer (`internal/web/handler_reviews.go`)
- Changed `RunReviewReq.ChangeId` to `RunReviewReq.ChangeNum`
- Updated all validation and error messages
- Modified API responses to return `changeNum` instead of `changeId`

### 2. Core Data Structures
- `internal/app/eino/core/context.go`: `FlowContext.ChangeId` → `FlowContext.ChangeNum`
- `internal/app/eino/core/store.go`: `ReviewStored.ChangeId` → `ReviewStored.ChangeNum`

### 3. Scheduler (`internal/app/scheduler/`)
- `Task.ChangeId` → `Task.ChangeNum`
- Updated `watcher.go` to extract `_number` field from Gerrit API response
- Updated `handler_changes.go` to use change number when submitting tasks

### 4. Eino Flow (`internal/app/eino/`)
- Updated `eino_flow.go` to use `changeNum` in graph inputs and node processing
- Updated `flows/review_flow.go` to pass `changeNum` to graph invocation
- Updated `mock_model.go` to use `changeNum` in mock tool calls

### 5. Tools (`internal/app/tools/`)
- Updated all Gerrit API function signatures:
  - `GerritTool.GetDiffs(changeNum, patchset)`
  - `GerritTool.GetFileContent(changeNum, revision, file)`
  - `GerritTool.GetFileContentFromParent(changeNum, revision, file)`
  - `GerritTool.PostReview(changeNum, revision, payload)`
- Updated `CodeContextTool.Fetch(enable, changeNum, patchset, diffs)`
- Updated `eino_code_review_tool.go` to use `changeNum` in request structure

### 6. Scripts
- Updated `trigger_review.ps1` to use `ChangeNum` parameter instead of `ChangeId`
- Changed default value from Change-ID hash to numeric example "12345"

## API Changes

### Before
```json
POST /reviews/run
{
  "changeId": "I311b5b222ad17002926076a4c9bfff1aad480760",
  "patchset": "3",
  "enableContext": true
}
```

### After
```json
POST /reviews/run
{
  "changeNum": "12345",
  "patchset": "3",
  "enableContext": true
}
```

## Gerrit API Compatibility
The Gerrit REST API accepts change numbers in all endpoints that previously accepted Change-IDs:
- `/a/changes/{change-number}/revisions/{revision-id}/files/`
- `/a/changes/{change-number}/revisions/{revision-id}/review`

Change numbers can be used directly in place of the full change identifier.

## Migration Notes

### For Watcher/Scheduler
The watcher now extracts the `_number` field from Gerrit API responses:
```go
if n, ok := c["_number"].(float64); ok {
    num := fmt.Sprintf("%.0f", n)
    pool.Submit(Task{ChangeNum: num, Patchset: "1", EnableContext: enableContext})
}
```

### For Manual API Calls
Users should now provide the numeric change number instead of the Change-ID hash:
```bash
# Before
curl -X POST http://localhost:8000/reviews/run \
  -H "Content-Type: application/json" \
  -d '{"changeId":"I311b5b222ad17002926076a4c9bfff1aad480760","patchset":"3"}'

# After
curl -X POST http://localhost:8000/reviews/run \
  -H "Content-Type: application/json" \
  -d '{"changeNum":"12345","patchset":"3"}'
```

## Testing
All test files have been updated to use `changeNum` in their invocations:
- `internal/app/eino/eino_flow_test.go`

## Benefits
1. **Uniqueness**: Change numbers are guaranteed unique across the Gerrit instance
2. **Simplicity**: Numeric identifiers are easier to work with than hash strings
3. **Correctness**: Eliminates ambiguity when the same Change-ID exists in multiple branches
4. **API Compatibility**: Gerrit API fully supports change numbers in all endpoints
