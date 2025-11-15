#!/bin/bash

# Example script to test homerun2-omni-pitcher API
# Make sure the service is running before executing this script

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "=== Testing homerun2-omni-pitcher API ==="
echo ""

# Test 1: Health check
echo "1. Testing health endpoint..."
curl -s "${BASE_URL}/health" | python3 -m json.tool
echo ""
echo ""

# Test 2: Invalid request - empty payload
echo "2. Testing validation - empty payload..."
curl -s -X POST "${BASE_URL}/pitch" \
  -H "Content-Type: application/json" \
  -d '{}' | python3 -m json.tool
echo ""
echo ""

# Test 3: Invalid request - missing message
echo "3. Testing validation - missing message..."
curl -s -X POST "${BASE_URL}/pitch" \
  -H "Content-Type: application/json" \
  -d '{"title":"test"}' | python3 -m json.tool
echo ""
echo ""

# Test 4: Valid request - minimal
echo "4. Testing valid request - minimal fields..."
curl -s -X POST "${BASE_URL}/pitch" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Message",
    "message": "This is a test message"
  }' | python3 -m json.tool || echo "Expected to fail if Redis is not available"
echo ""
echo ""

# Test 5: Valid request - full
echo "5. Testing valid request - all fields..."
curl -s -X POST "${BASE_URL}/pitch" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Deployment Notification",
    "message": "Service xyz deployed successfully to production",
    "severity": "success",
    "author": "ci-pipeline",
    "system": "demo-system",
    "tags": "deployment,production,success",
    "assigneeaddress": "ops-team@example.com",
    "assigneename": "Ops Team",
    "artifacts": "docker://registry.example.com/xyz:1.0.0",
    "url": "http://example.com/deployment/xyz"
  }' | python3 -m json.tool || echo "Expected to fail if Redis is not available"
echo ""
echo ""

echo "=== Testing completed ==="
