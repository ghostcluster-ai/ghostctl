#!/bin/bash
# Cleanup script for stale PR environments
# Run this periodically to clean up abandoned vClusters

set -e

NAMESPACE="${1:-ghostcluster}"
DRY_RUN="${DRY_RUN:-false}"

echo "PR Environment Cleanup"
echo "====================="
echo "Namespace: $NAMESPACE"
echo "Dry run: $DRY_RUN"
echo ""

# Check if ghostctl is installed
if ! command -v ghostctl &> /dev/null; then
    echo "Error: ghostctl not found. Please install it first."
    exit 1
fi

# List all PR environments
echo "Listing PR environments..."
PR_CLUSTERS=$(ghostctl list | grep '^pr-' | awk '{print $1}' || true)

if [ -z "$PR_CLUSTERS" ]; then
    echo "No PR environments found."
    exit 0
fi

echo "Found PR environments:"
echo "$PR_CLUSTERS"
echo ""

# Check each cluster
for CLUSTER in $PR_CLUSTERS; do
    echo "Checking: $CLUSTER"
    
    # Extract PR number from cluster name
    PR_NUM=$(echo "$CLUSTER" | sed 's/pr-//')
    
    # Check if PR exists in GitHub (requires gh CLI)
    if command -v gh &> /dev/null; then
        PR_STATE=$(gh pr view "$PR_NUM" --json state -q .state 2>/dev/null || echo "NOT_FOUND")
        
        if [ "$PR_STATE" = "CLOSED" ] || [ "$PR_STATE" = "MERGED" ] || [ "$PR_STATE" = "NOT_FOUND" ]; then
            echo "  → PR #$PR_NUM is $PR_STATE - marking for cleanup"
            
            if [ "$DRY_RUN" = "true" ]; then
                echo "  → [DRY RUN] Would delete: $CLUSTER"
            else
                echo "  → Deleting: $CLUSTER"
                ghostctl down "$CLUSTER" || echo "  → Failed to delete $CLUSTER"
            fi
        else
            echo "  → PR #$PR_NUM is still open ($PR_STATE)"
        fi
    else
        echo "  → gh CLI not found, skipping PR state check"
        echo "  → Install with: brew install gh"
    fi
    
    echo ""
done

echo "Cleanup complete!"
echo ""
echo "Current PR environments:"
ghostctl list | grep '^pr-' || echo "None"
