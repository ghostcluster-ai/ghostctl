#!/bin/bash
# Example 4: Multi-cluster setup for testing

set -e

echo "=== Example 4: Multi-cluster testing setup ==="

# Create multiple clusters with different configurations
create_cluster() {
    local name=$1
    local template=$2
    local ttl=$3
    
    echo "Creating cluster: $name (template: $template, ttl: $ttl)"
    ghostctl up "$name" --template "$template" --ttl "$ttl" --wait
}

# Create base cluster
create_cluster "testing-base" "default" "2h"

# Create specialized clusters
create_cluster "testing-gpu" "gpu" "4h"

# Wait for all clusters
echo "Waiting for all clusters to be ready..."
sleep 10

# List all clusters
echo "Active clusters:"
ghostctl list

# Show status of each cluster
for cluster in testing-base testing-gpu; do
    echo ""
    echo "Status of $cluster:"
    ghostctl status "$cluster" --detailed
done

echo "✓ Multi-cluster setup complete!"
echo ""
echo "Cleanup with:"
echo "  ghostctl down testing-base"
echo "  ghostctl down testing-gpu"
