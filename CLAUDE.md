# gitopsi

## Overview

Go CLI for bootstrapping GitOps repositories.

**Platforms:** Kubernetes, OpenShift, AKS, EKS
**Tools:** ArgoCD, Flux
**Scope:** Infrastructure, Applications, Both

## Commands

```bash
gitopsi init                    # Interactive
gitopsi init --config file.yaml # Config file
```

## Structure

```
cmd/gitopsi/         # Entry
internal/cli/        # Cobra commands
internal/config/     # Viper config
internal/generator/  # Generation logic
internal/platform/   # Platform specifics
templates/           # embed.FS templates
```

## Standards

- Go 1.22+
- gofmt/goimports
- Error wrapping with context
- No code comments
- Table-driven tests
