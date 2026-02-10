#!/bin/bash
# Example 1: Setup Ghostcluster and create a cluster

set -e

echo "=== Example 1: Setting up Ghostcluster ==="

# Initialize the controller
echo "Initializing Ghostcluster controller..."
ghostctl init --namespace ghostcluster --skip-validation

# Wait a moment for the controller to be ready
sleep 5

# Create a default cluster
echo "Creating default cluster..."
ghostctl up my-first-cluster --template default --wait

# Check cluster status
echo "Checking cluster status..."
ghostctl status my-first-cluster

# List all clusters
echo "Listing all clusters..."
ghostctl list

echo "✓ Cluster 'my-first-cluster' is ready!"
echo "Try these commands:"
echo "  ghostctl exec my-first-cluster 'kubectl get pods'"
echo "  ghostctl logs my-first-cluster -f"
echo "  ghostctl down my-first-cluster"
