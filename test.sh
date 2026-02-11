#!/bin/bash
# Test script for MCP Vikunja
# Runs the same checks as CI for local development

set -e

echo "🧪 Running MCP Vikunja test suite..."
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track overall status
FAILED=0

# Function to run a check
run_check() {
  local name=$1
  local cmd=$2

  echo -n "Running ${name}... "
  if eval "$cmd" > /dev/null 2>&1; then
    echo -e "${GREEN}✅${NC}"
    return 0
  else
    echo -e "${RED}❌${NC}"
    FAILED=1
    return 1
  fi
}

# Check 1: Formatting
echo "📋 Checking code formatting..."
if [ -n "$(gofmt -l .)" ]; then
  echo -e "${RED}❌ Code is not formatted${NC}"
  echo "Run: go fmt ./..."
  gofmt -l .
  FAILED=1
else
  echo -e "${GREEN}✅ Code formatting OK${NC}"
fi

# Check 2: Imports (if goimports is available)
if command -v goimports &> /dev/null; then
  echo "📋 Checking imports..."
  if [ -n "$(goimports -l .)" ]; then
    echo -e "${RED}❌ Imports need formatting${NC}"
    echo "Run: goimports -w ."
    goimports -l .
    FAILED=1
  else
    echo -e "${GREEN}✅ Imports formatting OK${NC}"
  fi
else
  echo -e "${YELLOW}⚠️  goimports not installed, skipping import check${NC}"
fi

# Check 3: Linting (if golangci-lint is available)
if command -v golangci-lint &> /dev/null; then
  echo "📋 Running golangci-lint..."
  if golangci-lint run --timeout=10m; then
    echo -e "${GREEN}✅ Linting passed${NC}"
  else
    echo -e "${RED}❌ Linting failed${NC}"
    FAILED=1
  fi
else
  echo -e "${YELLOW}⚠️  golangci-lint not installed, skipping lint check${NC}"
  echo "   Install: https://golangci-lint.run/usage/install/"
fi

# Check 4: go vet
echo "📋 Running go vet..."
if go vet ./...; then
  echo -e "${GREEN}✅ go vet passed${NC}"
else
  echo -e "${RED}❌ go vet failed${NC}"
  FAILED=1
fi

# Check 5: Tests with coverage
echo ""
echo "📋 Running tests with coverage..."
if go test -race -coverprofile=coverage.out ./...; then
  echo -e "${GREEN}✅ Tests passed${NC}"
  
  # Check coverage threshold
  COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
  echo ""
  echo "📊 Coverage: ${COVERAGE}%"
  
  if (( $(echo "$COVERAGE < 60" | bc -l) )); then
    echo -e "${RED}❌ Coverage below 60%${NC}"
    FAILED=1
  else
    echo -e "${GREEN}✅ Coverage threshold met${NC}"
  fi
  
  # Generate HTML coverage report
  go tool cover -html=coverage.out -o coverage.html
  echo "📄 Coverage report generated: coverage.html"
else
  echo -e "${RED}❌ Tests failed${NC}"
  FAILED=1
fi

# Check 6: Build both binaries
echo ""
echo "📋 Building binaries..."
echo -n "Building mcp-vikunja... "
if go build -o bin/mcp-vikunja ./cmd/mcp-vikunja 2>/dev/null; then
  echo -e "${GREEN}✅${NC}"
else
  echo -e "${RED}❌${NC}"
  FAILED=1
fi

echo -n "Building vikunja-cli... "
if go build -o bin/vikunja-cli ./cmd/vikunja-cli 2>/dev/null; then
  echo -e "${GREEN}✅${NC}"
else
  echo -e "${RED}❌${NC}"
  FAILED=1
fi

# Final summary
echo ""
echo "═══════════════════════════════════════════"
if [ $FAILED -eq 0 ]; then
  echo -e "${GREEN}✅ All checks passed!${NC}"
  echo ""
  echo "Binaries:"
  ls -lh bin/
  exit 0
else
  echo -e "${RED}❌ Some checks failed${NC}"
  exit 1
fi
