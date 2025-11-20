#!/bin/bash

# Janus Performance Benchmark Script
# Tests throughput and latency of the URL shortener service

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_URL="${API_URL:-http://localhost:8083}"
API_KEY="${API_KEY:-secret-api-key}"
NUM_REQUESTS="${NUM_REQUESTS:-1000}"
CONCURRENCY="${CONCURRENCY:-10}"

echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   Janus Performance Benchmark         ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
echo ""

# Check if service is running
echo -e "${YELLOW}→ Checking service health...${NC}"
if ! curl -s "${API_URL}/s/ping" > /dev/null; then
    echo -e "${RED}✗ Service is not running at ${API_URL}${NC}"
    echo -e "${YELLOW}  Start the service with: make docker-up${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Service is healthy${NC}"
echo ""

# Check if Apache Bench is installed
if ! command -v ab &> /dev/null; then
    echo -e "${RED}✗ Apache Bench (ab) is not installed${NC}"
    echo -e "${YELLOW}  Install with: sudo apt-get install apache2-utils${NC}"
    exit 1
fi

# Create test data
TEST_URL="https://example.com/test/$(date +%s)"
TEST_DATA=$(cat <<EOF
{
  "link": "${TEST_URL}"
}
EOF
)

echo -e "${YELLOW}→ Running benchmark...${NC}"
echo -e "  Requests: ${NUM_REQUESTS}"
echo -e "  Concurrency: ${CONCURRENCY}"
echo -e "  Endpoint: ${API_URL}/api/v1/link"
echo ""

# Create a temporary file for POST data
TMP_FILE=$(mktemp)
echo "${TEST_DATA}" > "${TMP_FILE}"

# Run Apache Bench
ab -n "${NUM_REQUESTS}" \
   -c "${CONCURRENCY}" \
   -p "${TMP_FILE}" \
   -T "application/json" \
   -H "Authorization: Bearer ${API_KEY}" \
   "${API_URL}/api/v1/link" | tee benchmark_results.txt

# Clean up
rm -f "${TMP_FILE}"

echo ""
echo -e "${GREEN}✓ Benchmark completed${NC}"
echo -e "${YELLOW}  Results saved to: benchmark_results.txt${NC}"

# Extract key metrics
echo ""
echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   Key Metrics                          ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"

grep "Requests per second" benchmark_results.txt | sed "s/^/${GREEN}  /"
grep "Time per request" benchmark_results.txt | head -1 | sed "s/^/${GREEN}  /"
grep "Transfer rate" benchmark_results.txt | sed "s/^/${GREEN}  /"

echo -e "${NC}"
