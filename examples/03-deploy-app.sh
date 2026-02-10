#!/bin/bash
# Example 3: Deploy an application to a cluster

set -e

echo "=== Example 3: Deploying an application ==="

CLUSTER_NAME="app-cluster"

# Create the cluster
echo "Creating cluster: $CLUSTER_NAME"
ghostctl up "$CLUSTER_NAME" --template default --ttl 2h

# Create a namespace for the app
echo "Creating namespace 'myapp'..."
ghostctl exec "$CLUSTER_NAME" 'kubectl create namespace myapp'

# Deploy a sample application (nginx)
echo "Deploying nginx application..."
ghostctl exec "$CLUSTER_NAME" \
  'kubectl create deployment nginx -n myapp --image=nginx:latest'

# Scale the deployment
echo "Scaling deployment to 3 replicas..."
ghostctl exec "$CLUSTER_NAME" \
  'kubectl scale deployment nginx -n myapp --replicas=3'

# Create a service
echo "Exposing service..."
ghostctl exec "$CLUSTER_NAME" \
  'kubectl expose deployment nginx -n myapp --type=ClusterIP --port=80'

# Check deployment status
echo "Deployment status:"
ghostctl exec "$CLUSTER_NAME" \
  'kubectl get deployment -n myapp'

echo "✓ Application deployed!"
echo "View logs with:"
echo "  ghostctl logs $CLUSTER_NAME -n myapp"
