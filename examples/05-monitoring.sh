#!/bin/bash
# Example 5: Monitoring and observability

set -e

echo "=== Example 5: Cluster monitoring ==="

CLUSTER_NAME="monitoring-cluster"

# Create cluster
echo "Creating monitoring cluster..."
ghostctl up "$CLUSTER_NAME" --template default --ttl 1h

# Get real-time status
echo "Real-time cluster status (press Ctrl+C to stop)..."
ghostctl status "$CLUSTER_NAME" --watch &
WATCH_PID=$!

# Let it run for a bit
sleep 10
kill $WATCH_PID 2>/dev/null || true

# Deploy monitoring application
echo ""
echo "Deploying monitoring stack..."
ghostctl exec "$CLUSTER_NAME" \
  'kubectl create namespace monitoring'

# Get cluster metrics
echo "Cluster metrics:"
ghostctl status "$CLUSTER_NAME" --detailed

# View logs with timestamps
echo "Recent logs with timestamps:"
ghostctl logs "$CLUSTER_NAME" --tail 20 --timestamps --follow=false

echo "✓ Monitoring example complete!"
