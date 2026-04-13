#!/bin/bash
set -e

VIKUNJA_HOST="${VIKUNJA_HOST:-http://localhost:3456}"
VIKUNJA_EMAIL="${VIKUNJA_EMAIL:-admin@example.com}"
VIKUNJA_PASSWORD="${VIKUNJA_PASSWORD:-admin123}"
VIKUNJA_USERNAME="${VIKUNJA_USERNAME:-admin}"
MAX_RETRIES=30
RETRY_INTERVAL=2

echo "🔧 Vikunja Development Setup"
echo "============================"
echo ""

wait_for_vikunja() {
    echo "⏳ Waiting for Vikunja to be ready..."
    local retries=0
    while [ $retries -lt $MAX_RETRIES ]; do
        if curl -sf "${VIKUNJA_HOST}/" > /dev/null 2>&1; then
            echo "✅ Vikunja is ready!"
            return 0
        fi
        retries=$((retries + 1))
        echo "  Attempt $retries/$MAX_RETRIES - still starting..."
        sleep $RETRY_INTERVAL
    done
    echo "❌ Vikunja did not become ready in time"
    exit 1
}

register_user() {
    local email="$1"
    local username="$2"
    local password="$3"
    
    curl -sf -X POST "${VIKUNJA_HOST}/api/v1/register" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"${email}\",\"username\":\"${username}\",\"password\":\"${password}\"}" \
        > /dev/null 2>&1
    return $?
}

get_token() {
    local email="$1"
    local password="$2"
    
    curl -sf -X POST "${VIKUNJA_HOST}/api/v1/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"${email}\",\"password\":\"${password}\"}" \
        2>/dev/null | grep -o '"token":"[^"]*"' | cut -d'"' -f4
}

check_user_exists() {
    local email="$1"
    local password="$2"
    
    local response=$(curl -sf -X POST "${VIKUNJA_HOST}/api/v1/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"${email}\",\"password\":\"${password}\"}" 2>&1)
    
    if echo "$response" | grep -q '"token"'; then
        return 0
    fi
    return 1
}

save_to_env() {
    local token="$1"
    local env_file="${2:-.env}"
    
    if [ ! -f "$env_file" ]; then
        touch "$env_file"
    fi
    
    if grep -q "^VIKUNJA_TOKEN=" "$env_file" 2>/dev/null; then
        sed -i.bak "s|^VIKUNJA_TOKEN=.*|VIKUNJA_TOKEN=${token}|" "$env_file"
    else
        echo "VIKUNJA_TOKEN=${token}" >> "$env_file"
    fi
    
    rm -f "${env_file}.bak" 2>/dev/null || true
    echo "✅ Token saved to $env_file"
}

wait_for_vikunja

echo ""
echo "📝 Checking if user ${VIKUNJA_EMAIL} exists..."

if check_user_exists "$VIKUNJA_EMAIL" "$VIKUNJA_PASSWORD"; then
    echo "✅ User already exists, getting token..."
    TOKEN=$(get_token "$VIKUNJA_EMAIL" "$VIKUNJA_PASSWORD")
else
    echo "👤 Creating new user..."
    if register_user "$VIKUNJA_EMAIL" "$VIKUNJA_USERNAME" "$VIKUNJA_PASSWORD"; then
        echo "✅ User created successfully"
    else
        echo "⚠️ User registration failed (may already exist)"
    fi
    
    echo "🔑 Generating API token..."
    TOKEN=$(get_token "$VIKUNJA_EMAIL" "$VIKUNJA_PASSWORD")
fi

if [ -z "$TOKEN" ]; then
    echo "❌ Failed to obtain API token"
    exit 1
fi

echo "✅ Token obtained successfully"
echo ""

ENV_FILE=".env"
save_to_env "$TOKEN" "$ENV_FILE"

echo ""
echo "✨ Setup complete!"
echo "   VIKUNJA_HOST=${VIKUNJA_HOST}"
echo "   VIKUNJA_EMAIL=${VIKUNJA_EMAIL}"
echo ""
