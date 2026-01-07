# gitopsi Configuration Examples

This document provides complete, ready-to-use configuration examples for various scenarios.

## Table of Contents

- [Basic Examples](#basic-examples)
- [Platform-Specific Examples](#platform-specific-examples)
- [Scope-Specific Examples](#scope-specific-examples)
- [Real-World Scenarios](#real-world-scenarios)
- [Enterprise Examples](#enterprise-examples)

---

## Basic Examples

### Minimal Configuration

The simplest possible configuration:

```yaml
# minimal.yaml
project:
  name: my-app

platform: kubernetes
scope: both
gitops_tool: argocd

environments:
  - name: dev
  - name: prod
```

```bash
gitopsi init --config minimal.yaml
```

### Standard Development Setup

A typical development environment:

```yaml
# dev-setup.yaml
project:
  name: dev-platform
  description: "Development Platform"

platform: kubernetes
scope: both
gitops_tool: argocd

environments:
  - name: dev
  - name: staging
  - name: prod

infrastructure:
  namespaces: true
  rbac: true
  network_policies: true
  resource_quotas: true

applications:
  - name: sample-app
    image: nginx:latest
    port: 80
    replicas: 1

docs:
  readme: true
  architecture: true
  onboarding: true
```

---

## Platform-Specific Examples

### Kubernetes (Generic)

```yaml
# kubernetes.yaml
project:
  name: k8s-platform
  description: "Standard Kubernetes GitOps Platform"

platform: kubernetes
scope: both
gitops_tool: argocd

environments:
  - name: dev
    cluster: https://192.168.1.100:6443
  - name: staging
    cluster: https://192.168.1.101:6443
  - name: prod
    cluster: https://192.168.1.102:6443

infrastructure:
  namespaces: true
  rbac: true
  network_policies: true
  resource_quotas: true

applications:
  - name: frontend
    image: myregistry/frontend:latest
    port: 3000
    replicas: 2
  - name: backend
    image: myregistry/backend:latest
    port: 8080
    replicas: 3

docs:
  readme: true
  architecture: true
  onboarding: true
```

### OpenShift

```yaml
# openshift.yaml
project:
  name: ocp-platform
  description: "OpenShift GitOps Platform"

platform: openshift
scope: both
gitops_tool: argocd

environments:
  - name: dev
    cluster: https://api.dev.ocp.example.com:6443
  - name: staging
    cluster: https://api.staging.ocp.example.com:6443
  - name: prod
    cluster: https://api.prod.ocp.example.com:6443

infrastructure:
  namespaces: true
  rbac: true
  network_policies: true
  resource_quotas: true

applications:
  - name: app
    image: image-registry.openshift-image-registry.svc:5000/myproject/app:latest
    port: 8080
    replicas: 2

docs:
  readme: true
  architecture: true
  onboarding: true
```

### Azure Kubernetes Service (AKS)

```yaml
# aks.yaml
project:
  name: aks-platform
  description: "Azure Kubernetes Service GitOps Platform"

platform: aks
scope: both
gitops_tool: argocd

output:
  type: git
  url: https://dev.azure.com/myorg/myproject/_git/aks-platform
  branch: main

environments:
  - name: dev
    cluster: https://aks-dev-dns.hcp.westus2.azmk8s.io:443
  - name: staging
    cluster: https://aks-staging-dns.hcp.westus2.azmk8s.io:443
  - name: prod
    cluster: https://aks-prod-dns.hcp.westus2.azmk8s.io:443

infrastructure:
  namespaces: true
  rbac: true
  network_policies: true
  resource_quotas: true

applications:
  - name: api
    image: myacr.azurecr.io/api:latest
    port: 8080
    replicas: 3
  - name: web
    image: myacr.azurecr.io/web:latest
    port: 3000
    replicas: 2

docs:
  readme: true
  architecture: true
  onboarding: true
```

### Amazon EKS

```yaml
# eks.yaml
project:
  name: eks-platform
  description: "Amazon EKS GitOps Platform"

platform: eks
scope: both
gitops_tool: argocd

output:
  type: git
  url: https://github.com/myorg/eks-platform.git
  branch: main

environments:
  - name: dev
    cluster: https://ABCD1234.gr7.us-east-1.eks.amazonaws.com
  - name: staging
    cluster: https://EFGH5678.gr7.us-east-1.eks.amazonaws.com
  - name: prod
    cluster: https://IJKL9012.gr7.us-east-1.eks.amazonaws.com

infrastructure:
  namespaces: true
  rbac: true
  network_policies: true
  resource_quotas: true

applications:
  - name: api
    image: 123456789012.dkr.ecr.us-east-1.amazonaws.com/api:latest
    port: 8080
    replicas: 3
  - name: web
    image: 123456789012.dkr.ecr.us-east-1.amazonaws.com/web:latest
    port: 3000
    replicas: 2

docs:
  readme: true
  architecture: true
  onboarding: true
```

---

## Scope-Specific Examples

### Infrastructure Only

For cluster operators who only manage infrastructure:

```yaml
# infrastructure-only.yaml
project:
  name: cluster-infrastructure
  description: "Cluster Infrastructure Management"

platform: kubernetes
scope: infrastructure
gitops_tool: argocd

environments:
  - name: dev
  - name: staging
  - name: prod

infrastructure:
  namespaces: true
  rbac: true
  network_policies: true
  resource_quotas: true

docs:
  readme: true
  architecture: true
  onboarding: true
```

### Applications Only

For application teams who don't manage infrastructure:

```yaml
# applications-only.yaml
project:
  name: my-applications
  description: "Application Deployments"

platform: kubernetes
scope: application
gitops_tool: argocd

environments:
  - name: dev
  - name: staging
  - name: prod

applications:
  - name: api
    image: myregistry/api:latest
    port: 8080
    replicas: 3
  - name: web
    image: myregistry/web:latest
    port: 3000
    replicas: 2
  - name: worker
    image: myregistry/worker:latest
    port: 9000
    replicas: 1

docs:
  readme: true
  architecture: true
  onboarding: true
```

---

## Real-World Scenarios

### E-Commerce Platform

```yaml
# ecommerce.yaml
project:
  name: ecommerce-platform
  description: "E-Commerce GitOps Platform"

platform: kubernetes
scope: both
gitops_tool: argocd

output:
  type: git
  url: https://github.com/mycompany/ecommerce-gitops.git
  branch: main

environments:
  - name: dev
    cluster: https://dev.k8s.mycompany.com
  - name: staging
    cluster: https://staging.k8s.mycompany.com
  - name: prod
    cluster: https://prod.k8s.mycompany.com

infrastructure:
  namespaces: true
  rbac: true
  network_policies: true
  resource_quotas: true

applications:
  - name: storefront
    image: mycompany/storefront:latest
    port: 3000
    replicas: 3
  - name: catalog-api
    image: mycompany/catalog-api:latest
    port: 8080
    replicas: 3
  - name: cart-api
    image: mycompany/cart-api:latest
    port: 8081
    replicas: 2
  - name: checkout-api
    image: mycompany/checkout-api:latest
    port: 8082
    replicas: 2
  - name: payment-service
    image: mycompany/payment-service:latest
    port: 8083
    replicas: 2
  - name: notification-service
    image: mycompany/notification-service:latest
    port: 8084
    replicas: 2

docs:
  readme: true
  architecture: true
  onboarding: true
```

### SaaS Application

```yaml
# saas-platform.yaml
project:
  name: saas-platform
  description: "Multi-Tenant SaaS Platform"

platform: kubernetes
scope: both
gitops_tool: argocd

output:
  type: git
  url: https://github.com/mysaas/platform-gitops.git
  branch: main

environments:
  - name: dev
  - name: staging
  - name: prod-us
  - name: prod-eu

infrastructure:
  namespaces: true
  rbac: true
  network_policies: true
  resource_quotas: true

applications:
  - name: api-gateway
    image: mysaas/api-gateway:latest
    port: 8080
    replicas: 3
  - name: auth-service
    image: mysaas/auth-service:latest
    port: 8081
    replicas: 2
  - name: tenant-service
    image: mysaas/tenant-service:latest
    port: 8082
    replicas: 2
  - name: billing-service
    image: mysaas/billing-service:latest
    port: 8083
    replicas: 2
  - name: web-app
    image: mysaas/web-app:latest
    port: 3000
    replicas: 2

docs:
  readme: true
  architecture: true
  onboarding: true
```

### Data Processing Pipeline

```yaml
# data-pipeline.yaml
project:
  name: data-pipeline
  description: "Data Processing Pipeline"

platform: kubernetes
scope: both
gitops_tool: argocd

environments:
  - name: dev
  - name: staging
  - name: prod

infrastructure:
  namespaces: true
  rbac: true
  network_policies: true
  resource_quotas: true

applications:
  - name: data-ingestion
    image: mycompany/data-ingestion:latest
    port: 8080
    replicas: 2
  - name: data-processor
    image: mycompany/data-processor:latest
    port: 8081
    replicas: 5
  - name: data-aggregator
    image: mycompany/data-aggregator:latest
    port: 8082
    replicas: 2
  - name: api-server
    image: mycompany/api-server:latest
    port: 8083
    replicas: 3

docs:
  readme: true
  architecture: true
  onboarding: true
```

---

## Enterprise Examples

### Multi-Team Platform

```yaml
# multi-team.yaml
project:
  name: enterprise-platform
  description: "Enterprise Multi-Team Platform"

platform: kubernetes
scope: infrastructure
gitops_tool: argocd

output:
  type: git
  url: https://github.com/enterprise/platform-gitops.git
  branch: main

environments:
  - name: dev
    cluster: https://dev.k8s.enterprise.com
  - name: staging
    cluster: https://staging.k8s.enterprise.com
  - name: prod-primary
    cluster: https://prod-primary.k8s.enterprise.com
  - name: prod-dr
    cluster: https://prod-dr.k8s.enterprise.com

infrastructure:
  namespaces: true
  rbac: true
  network_policies: true
  resource_quotas: true

docs:
  readme: true
  architecture: true
  onboarding: true
```

### Regulated Industry (Healthcare/Finance)

```yaml
# regulated-industry.yaml
project:
  name: regulated-platform
  description: "Regulated Industry Platform - HIPAA/PCI Compliant"

platform: kubernetes
scope: both
gitops_tool: argocd

output:
  type: git
  url: https://github.com/regulated-org/platform-gitops.git
  branch: main

environments:
  - name: dev
    cluster: https://dev.secure.example.com
  - name: staging
    cluster: https://staging.secure.example.com
  - name: prod
    cluster: https://prod.secure.example.com

infrastructure:
  namespaces: true
  rbac: true
  network_policies: true
  resource_quotas: true

applications:
  - name: secure-api
    image: secure-registry.example.com/secure-api:latest
    port: 8443
    replicas: 3
  - name: audit-service
    image: secure-registry.example.com/audit-service:latest
    port: 8444
    replicas: 2

docs:
  readme: true
  architecture: true
  onboarding: true
```

---

## Using Examples

1. **Copy the example** that matches your use case
2. **Customize** the values for your organization:
   - Replace `project.name` with your project name
   - Update `output.url` with your Git repository
   - Adjust `environments` with your cluster URLs
   - Configure `applications` with your container images
3. **Run gitopsi**:

```bash
gitopsi init --config your-config.yaml
```

4. **Review with dry-run first**:

```bash
gitopsi init --config your-config.yaml --dry-run --verbose
```

---

## Need Help?

- [Usage Guide](USAGE.md) - Detailed configuration options
- [README](../README.md) - Quick start guide
- [GitHub Issues](https://github.com/ihsanmokhlisse/gitopsi/issues) - Report issues or request features
