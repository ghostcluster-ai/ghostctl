package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	ClustersFileName     = "clusters.json"
	DefaultDir           = ".ghost"
	KubeconfigsDirName   = "kubeconfigs"
)

// ClusterMetadata represents metadata about a managed cluster
type ClusterMetadata struct {
	Name            string            `json:"name"`
	Namespace       string            `json:"namespace"`
	CreatedAt       time.Time         `json:"createdAt"`
	TTL             string            `json:"ttl,omitempty"`
	KubeconfigPath  string            `json:"kubeconfigPath"`
	HostCluster     string            `json:"hostCluster"`
	Template        string            `json:"template,omitempty"`
	CPU             string            `json:"cpu,omitempty"`
	Memory          string            `json:"memory,omitempty"`
	Storage         string            `json:"storage,omitempty"`
	GPU             int               `json:"gpu,omitempty"`
	GPUType         string            `json:"gpuType,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
}

// Store manages the metadata store
type Store struct {
	path string
}

// NewStore creates a new metadata store
func NewStore() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	basePath := filepath.Join(home, DefaultDir)
	if err := os.MkdirAll(basePath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create .ghost directory: %w", err)
	}

	// Create kubeconfigs directory
	kubeDir := filepath.Join(basePath, KubeconfigsDirName)
	if err := os.MkdirAll(kubeDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create kubeconfigs directory: %w", err)
	}

	path := filepath.Join(basePath, ClustersFileName)
	return &Store{path: path}, nil
}

// GetClusterPath returns the kubeconfig path for a cluster
func GetClusterPath(name string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, DefaultDir, KubeconfigsDirName, name+".yaml"), nil
}

// all returns all clusters from the store
func (s *Store) all() (map[string]*ClusterMetadata, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]*ClusterMetadata), nil
		}
		return nil, fmt.Errorf("failed to read metadata store: %w", err)
	}

	var clusters map[string]*ClusterMetadata
	if err := json.Unmarshal(data, &clusters); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return clusters, nil
}

// save persists all clusters to the store
func (s *Store) save(clusters map[string]*ClusterMetadata) error {
	data, err := json.MarshalIndent(clusters, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0600); err != nil {
		return fmt.Errorf("failed to write metadata store: %w", err)
	}

	return nil
}

// Add adds a new cluster to the store
func (s *Store) Add(meta *ClusterMetadata) error {
	clusters, err := s.all()
	if err != nil {
		return err
	}

	meta.CreatedAt = time.Now()
	clusters[meta.Name] = meta

	return s.save(clusters)
}

// Get retrieves a cluster from the store
func (s *Store) Get(name string) (*ClusterMetadata, error) {
	clusters, err := s.all()
	if err != nil {
		return nil, err
	}

	meta, exists := clusters[name]
	if !exists {
		return nil, fmt.Errorf("cluster not found: %s", name)
	}

	return meta, nil
}

// Remove removes a cluster from the store
func (s *Store) Remove(name string) error {
	clusters, err := s.all()
	if err != nil {
		return err
	}

	if _, exists := clusters[name]; !exists {
		return fmt.Errorf("cluster not found: %s", name)
	}

	delete(clusters, name)
	return s.save(clusters)
}

// List returns all clusters from the store
func (s *Store) List() ([]*ClusterMetadata, error) {
	clusters, err := s.all()
	if err != nil {
		return nil, err
	}

	var result []*ClusterMetadata
	for _, meta := range clusters {
		result = append(result, meta)
	}

	return result, nil
}

// Exists checks if a cluster exists in the store
func (s *Store) Exists(name string) bool {
	_, err := s.Get(name)
	return err == nil
}
