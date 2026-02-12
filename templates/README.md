# ghostctl Templates

This directory contains cluster templates that define standard configurations for vClusters.

## Available Templates

### default
Balanced resources for general development workloads.
- **CPU:** 2 cores
- **Memory:** 4Gi
- **Storage:** 20Gi
- **TTL:** 1h
- **Use Case:** Standard development and testing

### gpu
GPU-accelerated workload for ML/AI training and inference.
- **CPU:** 4 cores
- **Memory:** 16Gi
- **Storage:** 50Gi
- **GPU:** 1x nvidia-t4
- **TTL:** 2h
- **Use Case:** Machine learning, AI training, GPU-accelerated workloads

### minimal
Minimal resources for testing and quick experiments.
- **CPU:** 1 core
- **Memory:** 2Gi
- **Storage:** 10Gi
- **TTL:** 30m
- **Use Case:** Quick tests, CI/CD, lightweight experiments

### large
High-resource cluster for intensive workloads.
- **CPU:** 8 cores
- **Memory:** 32Gi
- **Storage:** 100Gi
- **TTL:** 4h
- **Use Case:** Resource-intensive applications, data processing

## Usage

### List all templates
```bash
ghostctl templates
ghostctl templates --extended        # Show all fields including storage and GPU type
ghostctl templates --format json     # Output as JSON
ghostctl templates --format yaml     # Output as YAML
```

### View template details
```bash
ghostctl templates gpu
ghostctl templates gpu --extended    # Show labels and full details
```

### Filter templates
```bash
ghostctl templates --filter gpu      # Show only templates with "gpu" in name/description
```

### Use templates to create clusters
```bash
# Use default template
ghostctl up my-cluster

# Use specific template
ghostctl up ml-job --template gpu

# Override template values
ghostctl up ml-job --template gpu --gpu 2 --ttl 4h

# Mix template with custom flags
ghostctl up test --template minimal --memory 4Gi
```

## Template Format

Templates are defined in YAML format:

```yaml
name: my-template
description: Description of the template
labels:
  tier: standard
  workload: general

cpu: "2"          # Number of CPU cores
memory: 4Gi       # Memory allocation
storage: 20Gi     # Storage allocation
gpu: 1            # Number of GPUs
gpuType: nvidia-t4  # GPU type
ttl: 1h           # Time-to-live
```

## Creating Custom Templates

1. Create a new YAML file in this directory (e.g., `custom.yaml`)
2. Define the template properties using the format above
3. The template name will be derived from the filename or the `name` field
4. Use it with: `ghostctl up my-cluster --template custom`

## Multi-Template Files

You can also define multiple templates in a single `templates.yaml` file:

```yaml
templates:
  - name: template1
    description: First template
    cpu: "2"
    memory: 4Gi
  - name: template2
    description: Second template
    cpu: "4"
    memory: 8Gi
```

## Override Behavior

When using templates with CLI flags:
- Template values are applied as defaults
- CLI flags override template values
- Unspecified values use template defaults

Example:
```bash
# GPU template has: cpu=4, memory=16Gi, gpu=1
ghostctl up ml --template gpu --gpu 2
# Result: cpu=4 (from template), memory=16Gi (from template), gpu=2 (from flag)
```
