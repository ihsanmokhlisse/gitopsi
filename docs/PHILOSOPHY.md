# gitopsi Core Philosophy

## Mission Statement

**gitopsi** exists to make GitOps accessible to everyone—from beginners to experts—by generating production-ready GitOps repository structures that embody best practices, security, and simplicity.

## Core Principles

### 1. Simplicity First 🎯

**"Make the right thing the easy thing."**

- **Sensible Defaults**: Every generated configuration should work out-of-the-box with no modifications
- **Progressive Complexity**: Start simple, add complexity only when needed
- **Minimal Configuration**: Require only essential inputs, derive everything else intelligently
- **Clear Abstractions**: Hide complexity without hiding capability

```yaml
# Bad: Requiring many fields
gitops_tool: argocd
argocd_namespace: argocd
argocd_project: default
argocd_sync_policy: automated
argocd_prune: true
argocd_self_heal: true
# ... 20 more fields

# Good: Smart defaults
gitops_tool: argocd  # Everything else derived automatically
```

### 2. Best Practices by Default 📋

**"Don't just work, work correctly."**

Every generated manifest incorporates:

- **Security**: Non-root containers, read-only filesystems, resource limits
- **Reliability**: Health checks, pod disruption budgets, graceful shutdown
- **Observability**: Prometheus annotations, structured logging, tracing headers
- **Compliance**: Network policies, RBAC, audit logging

We don't generate "hello world" manifests—we generate production-ready code.

### 3. Convention Over Configuration 📐

**"If there's one good way, make it the only way."**

- **Standard Directory Structure**: Every repository follows the same layout
- **Naming Conventions**: Resources named consistently and predictably
- **Environment Patterns**: dev/staging/prod follows established patterns
- **Git Workflow**: Trunk-based development with feature branches

```
project-name/
├── applications/          # Application manifests
│   ├── base/             # Base configurations
│   └── overlays/         # Environment-specific
├── infrastructure/        # Platform resources
│   ├── base/             # Base infrastructure
│   └── overlays/         # Environment-specific
├── argocd/               # GitOps tool configuration
│   ├── projects/         # ArgoCD AppProjects
│   ├── applications/     # ArgoCD Applications
│   └── applicationsets/  # ArgoCD ApplicationSets
└── bootstrap/            # Initial setup scripts
```

### 4. Declarative Everything 📝

**"If it's not in Git, it doesn't exist."**

- **No Imperative Commands**: All changes through Git commits
- **Reproducible State**: Repository can recreate the entire system
- **Version Control**: Every change tracked and auditable
- **Review Process**: All changes through pull requests

### 5. Security as a Foundation 🔒

**"Security is not a feature, it's a requirement."**

- **Zero Trust**: Every component assumes the network is hostile
- **Least Privilege**: Minimal permissions for every service
- **Secret Management**: Never store secrets in Git, use references
- **Supply Chain**: Signed images, verified sources, SBOM generation

### 6. Developer Experience 🧑‍💻

**"Developers should focus on features, not infrastructure."**

- **Quick Start**: From zero to working GitOps in minutes
- **Clear Feedback**: Helpful error messages with suggestions
- **Documentation**: Generated docs explain what and why
- **Onboarding**: New team members productive in hours

## Design Decisions

### Why Kustomize Over Helm for Infrastructure?

1. **Transparency**: YAML in, YAML out—no templating surprises
2. **Composability**: Layers of patches, not layers of values
3. **GitOps Friendly**: Native support in ArgoCD and Flux
4. **Auditability**: Easy to see exactly what will be deployed

### Why ApplicationSets Over Applications?

1. **Scalability**: One definition for many deployments
2. **Consistency**: Automatic environment synchronization
3. **Maintainability**: Single source of truth
4. **Flexibility**: Powerful generators for complex topologies

### Why Opinionated Structure?

1. **Team Alignment**: Everyone knows where to find things
2. **Tool Integration**: IDE support, linting, CI/CD
3. **Documentation**: Structure is self-documenting
4. **Best Practices**: Layout embodies years of experience

## User Experience Goals

### For Beginners

- Start with a single command
- Generated documentation explains each file
- Links to learning resources
- Progressive disclosure of advanced features

### For Teams

- Consistent structure across projects
- Clear ownership boundaries
- Built-in PR workflow templates
- Automated validation and testing

### For Enterprises

- Multi-cluster support out of the box
- RBAC and multi-tenancy patterns
- Compliance and audit trails
- Integration with existing tooling

## What gitopsi Is NOT

- **Not a GitOps Controller**: We generate repos, ArgoCD/Flux deploys them
- **Not a Kubernetes Distribution**: We work with any K8s/OpenShift cluster
- **Not a CI/CD System**: We integrate with existing pipelines
- **Not a Secret Manager**: We integrate with Vault, External Secrets, etc.

## Measuring Success

gitopsi is successful when:

1. **Time to GitOps**: Teams can set up GitOps in under 10 minutes
2. **Confidence**: Developers trust generated configurations
3. **Maintenance**: Updates are painless and safe
4. **Community**: Patterns are shared and improved

## Contributing to Philosophy

This document is a living guide. If you believe a principle should change:

1. Open an issue describing the change and rationale
2. Provide real-world examples supporting the change
3. Consider backward compatibility and migration
4. Discuss with maintainers before implementing

---

*"The best tool is the one you don't notice—it just works."*





