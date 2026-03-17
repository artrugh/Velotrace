#!/bin/sh

# Configuration
CONTRACTS_DIR="./libs/api-contract"
CONVERT_CMD="npx swagger2openapi"
GEN_CMD="npx openapi-typescript"

# Function to sync a single service (Convert + Generate)
sync_service() {
    swagger_file="$1"
    dir=$(dirname "$swagger_file")
    gen_dir="$dir/generated"
    mkdir -p "$gen_dir" 2>/dev/null
    
    openapi_file="$gen_dir/openapi.yaml"
    
    # 1. Convert Swagger 2.0 to OpenAPI 3.0
    echo "🔄 Converting $swagger_file to OpenAPI 3.0..."
    $CONVERT_CMD "$swagger_file" --outfile "$openapi_file" --patch
    
    # 2. Extract service name (e.g., identity-api -> identity)
    service_name=$(basename "$dir" | sed 's/-api//')
    output_file="$gen_dir/$service_name.ts"
    
    # 3. Generate TypeScript types
    echo "⚙️  Generating types for $service_name in $gen_dir..."
    $GEN_CMD "$openapi_file" -o "$output_file"
}

# Function to sync all services
sync_all() {
    echo "📡 Starting full API contract synchronization..."
    find "$CONTRACTS_DIR" -name "swagger.yaml" | while read -r yaml_file; do
        sync_service "$yaml_file"
    done
    echo "✅ All contracts synchronized."
}

# Check if we should watch
if [ "$1" = "--watch" ]; then
    sync_all
    echo "👀 Watching for changes in $CONTRACTS_DIR... (Ctrl+C to stop)"
    
    LAST_HASH=""
    while true; do
        # We watch the source swagger.yaml files
        CURRENT_HASH=$(find "$CONTRACTS_DIR" -name "swagger.yaml" -exec stat -c %Y {} + 2>/dev/null)
        
        if [ "$CURRENT_HASH" != "$LAST_HASH" ]; then
            if [ "$LAST_HASH" != "" ]; then
                echo "🔄 Change detected in swagger.yaml! Regenerating everything..."
                sync_all
            fi
            LAST_HASH="$CURRENT_HASH"
        fi
        sleep 2
    done
else
    # Just run once
    sync_all
fi
