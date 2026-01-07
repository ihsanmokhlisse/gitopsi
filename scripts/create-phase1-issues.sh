#!/bin/bash
# Script to create Phase 1 GitHub Issues
# Run: gh auth login (first time)
# Then: ./scripts/create-phase1-issues.sh

set -e

REPO="ihsanmokhlisse/gitopsi"
MILESTONE="v0.1.0 - MVP"

echo "Creating Phase 1 issues for gitopsi..."

# Create milestone first
gh api repos/$REPO/milestones -f title="$MILESTONE" -f description="Phase 1 MVP - Basic K8s + ArgoCD generation" -f state="open" 2>/dev/null || echo "Milestone may already exist"

# Issue 1: CLI Structure
gh issue create \
  --repo $REPO \
  --title "[FEATURE] CLI structure with Cobra" \
  --label "enhancement,phase-1" \
  --body "## Summary
Set up the basic CLI structure using Cobra framework.

## Acceptance Criteria
- [ ] Create \`cmd/gitopsi/main.go\` entry point
- [ ] Create \`internal/cli/root.go\` with root command
- [ ] Implement \`version\` command
- [ ] Add global flags: \`--verbose\`, \`--config\`, \`--output\`, \`--dry-run\`
- [ ] Proper error handling and exit codes

## Technical Notes
- Use cobra.Command structure
- Version info via ldflags at build time

## Related
Part of Phase 1 MVP (v0.1.0)"

echo "✓ Created Issue: CLI structure"

# Issue 2: Config Parsing
gh issue create \
  --repo $REPO \
  --title "[FEATURE] Config file parsing with Viper" \
  --label "enhancement,phase-1" \
  --body "## Summary
Implement configuration file parsing using Viper.

## Acceptance Criteria
- [ ] Create \`internal/config/config.go\` with Config struct
- [ ] Create \`internal/config/loader.go\` for YAML loading
- [ ] Create \`internal/config/validate.go\` for validation
- [ ] Support \`gitops.yaml\` and \`gitops.yml\` filenames
- [ ] Environment variable override support
- [ ] Clear validation error messages

## Config Schema
\`\`\`yaml
project:
  name: string (required)
  description: string
platform: kubernetes|openshift|aks|eks
scope: infrastructure|application|both
gitops_tool: argocd|flux|both
environments: []
applications: []
\`\`\`

## Related
Part of Phase 1 MVP (v0.1.0)"

echo "✓ Created Issue: Config parsing"

# Issue 3: Interactive Prompts
gh issue create \
  --repo $REPO \
  --title "[FEATURE] Interactive prompts with Survey" \
  --label "enhancement,phase-1" \
  --body "## Summary
Implement interactive prompts using Survey v2 for the init command.

## Acceptance Criteria
- [ ] Create \`internal/prompt/prompt.go\`
- [ ] Prompt for project name
- [ ] Prompt for platform selection (select)
- [ ] Prompt for scope selection (select)
- [ ] Prompt for GitOps tool selection (select)
- [ ] Prompt for environments (multi-select)
- [ ] Prompt for documentation generation (confirm)
- [ ] Return populated Config struct

## User Experience
\`\`\`
? Project name: my-platform
? Platform: [Kubernetes, OpenShift, AKS, EKS]
? Scope: [infrastructure, application, both]
? GitOps tool: [argocd, flux, both]
? Environments: [dev, staging, prod]
? Generate documentation? Yes
\`\`\`

## Related
Part of Phase 1 MVP (v0.1.0)"

echo "✓ Created Issue: Interactive prompts"

# Issue 4: Init Command
gh issue create \
  --repo $REPO \
  --title "[FEATURE] Init command implementation" \
  --label "enhancement,phase-1" \
  --body "## Summary
Implement the \`gitopsi init\` command that orchestrates repository generation.

## Acceptance Criteria
- [ ] Create \`internal/cli/init.go\`
- [ ] Support interactive mode (no flags)
- [ ] Support config file mode (\`--config\`)
- [ ] Support dry-run mode (\`--dry-run\`)
- [ ] Create output directory structure
- [ ] Call appropriate generators
- [ ] Display summary on completion

## Command Interface
\`\`\`bash
gitopsi init                    # Interactive
gitopsi init --config file.yaml # Config file
gitopsi init --dry-run          # Preview
\`\`\`

## Related
Depends on: Config parsing, Interactive prompts
Part of Phase 1 MVP (v0.1.0)"

echo "✓ Created Issue: Init command"

# Issue 5: Kubernetes Manifests
gh issue create \
  --repo $REPO \
  --title "[FEATURE] Kubernetes manifest generation" \
  --label "enhancement,phase-1" \
  --body "## Summary
Generate basic Kubernetes manifests for applications.

## Acceptance Criteria
- [ ] Create \`internal/generator/generator.go\` interface
- [ ] Create \`internal/generator/applications.go\`
- [ ] Create \`internal/platform/kubernetes.go\`
- [ ] Generate Deployment manifest
- [ ] Generate Service manifest
- [ ] Generate ConfigMap manifest
- [ ] Generate Ingress manifest
- [ ] Use Go text/template

## Templates Required
- \`templates/kubernetes/deployment.yaml.tmpl\`
- \`templates/kubernetes/service.yaml.tmpl\`
- \`templates/kubernetes/configmap.yaml.tmpl\`
- \`templates/kubernetes/ingress.yaml.tmpl\`

## Related
Part of Phase 1 MVP (v0.1.0)"

echo "✓ Created Issue: Kubernetes manifests"

# Issue 6: Kustomize Structure
gh issue create \
  --repo $REPO \
  --title "[FEATURE] Kustomize base/overlay structure" \
  --label "enhancement,phase-1" \
  --body "## Summary
Generate Kustomize structure with base and environment overlays.

## Acceptance Criteria
- [ ] Generate \`base/kustomization.yaml\`
- [ ] Generate \`overlays/{env}/kustomization.yaml\` for each environment
- [ ] Generate environment-specific patches
- [ ] Proper resource references
- [ ] Support namespace transformation

## Generated Structure
\`\`\`
applications/
├── base/
│   ├── kustomization.yaml
│   └── {app}/
│       ├── deployment.yaml
│       ├── service.yaml
│       └── kustomization.yaml
└── overlays/
    ├── dev/
    │   └── kustomization.yaml
    ├── staging/
    │   └── kustomization.yaml
    └── prod/
        └── kustomization.yaml
\`\`\`

## Related
Depends on: Kubernetes manifests
Part of Phase 1 MVP (v0.1.0)"

echo "✓ Created Issue: Kustomize structure"

# Issue 7: ArgoCD Generation
gh issue create \
  --repo $REPO \
  --title "[FEATURE] ArgoCD Application generation" \
  --label "enhancement,phase-1" \
  --body "## Summary
Generate ArgoCD Application and AppProject resources.

## Acceptance Criteria
- [ ] Create \`internal/generator/argocd.go\`
- [ ] Generate ArgoCD Application manifest
- [ ] Generate ArgoCD AppProject manifest
- [ ] Support sync policy configuration
- [ ] Generate app-of-apps pattern
- [ ] Proper destination configuration

## Templates Required
- \`templates/argocd/application.yaml.tmpl\`
- \`templates/argocd/project.yaml.tmpl\`
- \`templates/argocd/app-of-apps.yaml.tmpl\`

## Generated Structure
\`\`\`
argocd/
├── projects/
│   └── applications.yaml
└── applications/
    └── {app}.yaml
bootstrap/
└── argocd/
    └── app-of-apps.yaml
\`\`\`

## Related
Part of Phase 1 MVP (v0.1.0)"

echo "✓ Created Issue: ArgoCD generation"

# Issue 8: Template System
gh issue create \
  --repo $REPO \
  --title "[FEATURE] Template embedding with embed.FS" \
  --label "enhancement,phase-1" \
  --body "## Summary
Set up template embedding using Go's embed.FS.

## Acceptance Criteria
- [ ] Create \`internal/templates/templates.go\` with embed directive
- [ ] Organize templates by category (kubernetes/, argocd/, docs/)
- [ ] Helper functions for template loading
- [ ] Template execution with data binding
- [ ] Error handling for missing templates

## Directory Structure
\`\`\`
templates/
├── kubernetes/
│   ├── deployment.yaml.tmpl
│   ├── service.yaml.tmpl
│   └── ...
├── argocd/
│   ├── application.yaml.tmpl
│   └── ...
└── docs/
    └── README.md.tmpl
\`\`\`

## Related
Part of Phase 1 MVP (v0.1.0)"

echo "✓ Created Issue: Template system"

# Issue 9: File Output
gh issue create \
  --repo $REPO \
  --title "[FEATURE] File output system" \
  --label "enhancement,phase-1" \
  --body "## Summary
Implement file writing system for generated content.

## Acceptance Criteria
- [ ] Create \`internal/output/writer.go\`
- [ ] Create directories recursively
- [ ] Write files with proper permissions
- [ ] Support dry-run mode (print only)
- [ ] Handle file conflicts
- [ ] Summary of files created

## Interface
\`\`\`go
type Writer interface {
    WriteFile(path string, content []byte) error
    CreateDir(path string) error
    SetDryRun(enabled bool)
}
\`\`\`

## Related
Part of Phase 1 MVP (v0.1.0)"

echo "✓ Created Issue: File output"

# Issue 10: README Generation
gh issue create \
  --repo $REPO \
  --title "[FEATURE] README.md generation" \
  --label "enhancement,phase-1" \
  --body "## Summary
Generate README.md documentation for created repositories.

## Acceptance Criteria
- [ ] Create \`internal/generator/docs.go\`
- [ ] Generate README.md with project info
- [ ] Include directory structure
- [ ] Include quick start instructions
- [ ] Include environment details
- [ ] Makefile generation

## Template Content
- Project name and description
- Prerequisites
- Directory structure
- Quick start commands
- Environment details

## Related
Part of Phase 1 MVP (v0.1.0)"

echo "✓ Created Issue: README generation"

# Issue 11: Unit Tests
gh issue create \
  --repo $REPO \
  --title "[FEATURE] Unit tests for Phase 1" \
  --label "enhancement,phase-1,testing" \
  --body "## Summary
Write unit tests for all Phase 1 components.

## Acceptance Criteria
- [ ] Config validation tests
- [ ] Template rendering tests
- [ ] Generator tests with golden files
- [ ] CLI command tests
- [ ] >80% code coverage

## Test Files
- \`internal/config/config_test.go\`
- \`internal/generator/generator_test.go\`
- \`internal/templates/templates_test.go\`
- \`internal/cli/init_test.go\`

## Golden Files
Store expected outputs in \`testdata/golden/\`

## Related
Part of Phase 1 MVP (v0.1.0)"

echo "✓ Created Issue: Unit tests"

echo ""
echo "✅ All Phase 1 issues created!"
echo ""
echo "View issues: gh issue list --repo $REPO"
