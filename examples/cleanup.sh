#!/bin/bash
# Example: Cleanup script to remove all clusters

set -e

echo "=== Cleanup: Removing all ghostctl clusters ==="

# Get list of all clusters
CLUSTERS=$(ghostctl list --output table | tail -n +2 | awk '{print $1}')

if [ -z "$CLUSTERS" ]; then
    echo "No clusters found"
    exit 0
fi

echo "Found clusters to delete:"
echo "$CLUSTERS"
echo ""

# Confirm before deletion
read -p "Are you sure you want to delete all clusters? (yes/no): " confirm
if [ "$confirm" != "yes" ]; then
    echo "Cancelled"
    exit 0
fi

# Delete each cluster
while IFS= read -r cluster; do
    if [ -n "$cluster" ]; then
        echo "Deleting cluster: $cluster"
        ghostctl down "$cluster" --force
    fi
done <<< "$CLUSTERS"

echo "✓ All clusters have been deleted"
