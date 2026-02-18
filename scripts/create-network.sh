#!/bin/bash

# Create shared Docker network for WordPress Platform
# This network will be used by Traefik and all WordPress containers
# for external access and reverse proxy functionality

set -e

NETWORK_NAME="wp-public"

echo "Creating Docker network: $NETWORK_NAME"

# Check if network already exists
if docker network ls --format '{{.Name}}' | grep -q "^${NETWORK_NAME}$"; then
    echo "Network $NETWORK_NAME already exists"
    exit 0
fi

# Create the network
docker network create \
    --driver bridge \
    --attachable \
    $NETWORK_NAME

echo "Network $NETWORK_NAME created successfully"
echo ""
echo "Network details:"
docker network inspect $NETWORK_NAME --format '{{json .}}' | python3 -m json.tool 2>/dev/null || docker network inspect $NETWORK_NAME
