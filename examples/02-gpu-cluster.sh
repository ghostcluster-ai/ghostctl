#!/bin/bash
# Example 2: Create a GPU-enabled cluster for ML workloads

set -e

echo "=== Example 2: Creating GPU cluster for ML workloads ==="

CLUSTER_NAME="ml-experiment-$(date +%s)"

echo "Creating GPU cluster: $CLUSTER_NAME"
ghostctl up "$CLUSTER_NAME" \
  --template gpu \
  --gpu 1 \
  --gpu-type nvidia-a100 \
  --memory 32Gi \
  --cpu 8 \
  --ttl 4h

echo "Waiting for cluster to be ready..."
sleep 10

# Check GPU status
echo "Checking GPU allocation..."
ghostctl exec "$CLUSTER_NAME" 'kubectl describe nodes | grep nvidia'

# Deploy a sample ML workload
echo "Deploying sample ML workload..."
ghostctl exec "$CLUSTER_NAME" 'kubectl create deployment ml-job --image=nvidia/cuda:11.8.0-runtime-ubuntu22.04'

# Check pod status
ghostctl exec "$CLUSTER_NAME" 'kubectl get pods'

echo "✓ GPU cluster '$CLUSTER_NAME' is ready!"
echo "Monitor the workload with:"
echo "  ghostctl logs $CLUSTER_NAME -f"
echo "  ghostctl status $CLUSTER_NAME --watch"
