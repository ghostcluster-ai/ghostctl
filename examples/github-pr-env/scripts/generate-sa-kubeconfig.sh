#!/bin/bash
# Generate a kubeconfig file for a Kubernetes service account
# This is useful for creating limited-permission credentials for CI/CD

set -e

# Configuration
SERVICE_ACCOUNT_NAME="${1:-ghostctl-ci}"
NAMESPACE="${2:-ghostcluster}"
CLUSTER_NAME="${3:-$(kubectl config current-context)}"

echo "Generating kubeconfig for service account: $SERVICE_ACCOUNT_NAME"
echo "Namespace: $NAMESPACE"
echo "Cluster: $CLUSTER_NAME"
echo ""

# Create namespace if it doesn't exist
kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

# Create service account
kubectl create serviceaccount "$SERVICE_ACCOUNT_NAME" -n "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

# Create cluster role for vCluster operations
cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: ${SERVICE_ACCOUNT_NAME}-role
rules:
  # Permissions for vCluster management
  - apiGroups: [""]
    resources: ["namespaces"]
    verbs: ["get", "list", "create", "delete"]
  - apiGroups: [""]
    resources: ["pods", "services", "configmaps", "secrets", "persistentvolumeclaims", "serviceaccounts"]
    verbs: ["get", "list", "create", "delete", "patch", "update", "watch"]
  - apiGroups: ["apps"]
    resources: ["statefulsets", "deployments", "replicasets"]
    verbs: ["get", "list", "create", "delete", "patch", "update"]
  - apiGroups: ["rbac.authorization.k8s.io"]
    resources: ["roles", "rolebindings", "clusterroles", "clusterrolebindings"]
    verbs: ["get", "list", "create", "delete", "patch", "update"]
  # Additional permissions for vCluster
  - apiGroups: ["networking.k8s.io"]
    resources: ["networkpolicies"]
    verbs: ["get", "list", "create", "delete"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses"]
    verbs: ["get", "list"]
EOF

# Create cluster role binding
cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: ${SERVICE_ACCOUNT_NAME}-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: ${SERVICE_ACCOUNT_NAME}-role
subjects:
  - kind: ServiceAccount
    name: $SERVICE_ACCOUNT_NAME
    namespace: $NAMESPACE
EOF

echo "✓ Service account and RBAC created"

# For Kubernetes 1.24+, we need to manually create a token
SECRET_NAME="${SERVICE_ACCOUNT_NAME}-token"

cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: $SECRET_NAME
  namespace: $NAMESPACE
  annotations:
    kubernetes.io/service-account.name: $SERVICE_ACCOUNT_NAME
type: kubernetes.io/service-account-token
EOF

# Wait for token to be populated
echo "Waiting for token to be generated..."
sleep 2

# Get the token
TOKEN=$(kubectl get secret "$SECRET_NAME" -n "$NAMESPACE" -o jsonpath='{.data.token}' | base64 -d)

# Get cluster information
CLUSTER_SERVER=$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')
CLUSTER_CA=$(kubectl get secret "$SECRET_NAME" -n "$NAMESPACE" -o jsonpath='{.data.ca\.crt}')

# Generate kubeconfig
OUTPUT_FILE="${SERVICE_ACCOUNT_NAME}-kubeconfig.yaml"

cat > "$OUTPUT_FILE" <<EOF
apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: $CLUSTER_CA
    server: $CLUSTER_SERVER
  name: $CLUSTER_NAME
contexts:
- context:
    cluster: $CLUSTER_NAME
    namespace: $NAMESPACE
    user: $SERVICE_ACCOUNT_NAME
  name: ${SERVICE_ACCOUNT_NAME}@${CLUSTER_NAME}
current-context: ${SERVICE_ACCOUNT_NAME}@${CLUSTER_NAME}
users:
- name: $SERVICE_ACCOUNT_NAME
  user:
    token: $TOKEN
EOF

echo ""
echo "✓ Kubeconfig generated: $OUTPUT_FILE"
echo ""
echo "Test the kubeconfig:"
echo "  export KUBECONFIG=$(pwd)/$OUTPUT_FILE"
echo "  kubectl get nodes"
echo ""
echo "Use with Terraform:"
echo "  export TF_VAR_ghostcluster_kubeconfig=\$(cat $(pwd)/$OUTPUT_FILE)"
echo ""
echo "WARNING: This file contains sensitive credentials. Keep it secure!"
echo "         Do not commit it to version control."
