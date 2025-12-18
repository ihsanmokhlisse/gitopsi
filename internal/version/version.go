// Package version provides Kubernetes API version compatibility management.
package version

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// KubernetesVersion represents a parsed Kubernetes version.
type KubernetesVersion struct {
	Major int
	Minor int
	Patch int
	Raw   string
}

// APIVersionMapping maps resource kinds to their API versions for different K8s versions.
type APIVersionMapping struct {
	Kind           string
	PreferredAPI   string
	DeprecatedAPIs []DeprecatedAPI
	IntroducedIn   string // Kubernetes version where this API was introduced
	RemovedIn      string // Kubernetes version where this API was removed (empty if still available)
}

// DeprecatedAPI represents a deprecated API version.
type DeprecatedAPI struct {
	APIVersion   string
	DeprecatedIn string // Kubernetes version where this was deprecated
	RemovedIn    string // Kubernetes version where this was removed
	Replacement  string // The replacement API version
}

// Mapper provides version-aware API version selection.
type Mapper struct {
	targetVersion *KubernetesVersion
	platform      string // kubernetes, openshift
	mappings      map[string]APIVersionMapping
}

// DefaultAPIVersions maps resource kinds to their current preferred API versions.
var DefaultAPIVersions = map[string]string{
	// Core resources
	"Namespace":             "v1",
	"ConfigMap":             "v1",
	"Secret":                "v1",
	"Service":               "v1",
	"ServiceAccount":        "v1",
	"PersistentVolumeClaim": "v1",
	"PersistentVolume":      "v1",
	"Pod":                   "v1",
	"Node":                  "v1",
	"Event":                 "v1",
	"Endpoints":             "v1",
	"LimitRange":            "v1",
	"ResourceQuota":         "v1",

	// Apps resources
	"Deployment":  "apps/v1",
	"StatefulSet": "apps/v1",
	"DaemonSet":   "apps/v1",
	"ReplicaSet":  "apps/v1",

	// Networking
	"Ingress":       "networking.k8s.io/v1",
	"NetworkPolicy": "networking.k8s.io/v1",
	"IngressClass":  "networking.k8s.io/v1",

	// Batch
	"Job":     "batch/v1",
	"CronJob": "batch/v1",

	// RBAC
	"Role":               "rbac.authorization.k8s.io/v1",
	"RoleBinding":        "rbac.authorization.k8s.io/v1",
	"ClusterRole":        "rbac.authorization.k8s.io/v1",
	"ClusterRoleBinding": "rbac.authorization.k8s.io/v1",

	// Autoscaling
	"HorizontalPodAutoscaler": "autoscaling/v2",

	// Policy
	"PodDisruptionBudget": "policy/v1",

	// Storage
	"StorageClass":        "storage.k8s.io/v1",
	"VolumeSnapshot":      "snapshot.storage.k8s.io/v1",
	"CSIDriver":           "storage.k8s.io/v1",
	"CSINode":             "storage.k8s.io/v1",
	"VolumeSnapshotClass": "snapshot.storage.k8s.io/v1",

	// ArgoCD
	"Application":    "argoproj.io/v1alpha1",
	"ApplicationSet": "argoproj.io/v1alpha1",
	"AppProject":     "argoproj.io/v1alpha1",

	// Flux
	"GitRepository":  "source.toolkit.fluxcd.io/v1",
	"Kustomization":  "kustomize.toolkit.fluxcd.io/v1",
	"HelmRepository": "source.toolkit.fluxcd.io/v1",
	"HelmRelease":    "helm.toolkit.fluxcd.io/v2",
	"HelmChart":      "source.toolkit.fluxcd.io/v1",
	"OCIRepository":  "source.toolkit.fluxcd.io/v1beta2",
	"Bucket":         "source.toolkit.fluxcd.io/v1beta2",

	// OpenShift specific
	"Route":            "route.openshift.io/v1",
	"DeploymentConfig": "apps.openshift.io/v1",
	"BuildConfig":      "build.openshift.io/v1",
	"ImageStream":      "image.openshift.io/v1",
	"Project":          "project.openshift.io/v1",
}

// apiVersionHistory contains historical API version information.
var apiVersionHistory = []APIVersionMapping{
	{
		Kind:         "Ingress",
		PreferredAPI: "networking.k8s.io/v1",
		DeprecatedAPIs: []DeprecatedAPI{
			{
				APIVersion:   "networking.k8s.io/v1beta1",
				DeprecatedIn: "1.19",
				RemovedIn:    "1.22",
				Replacement:  "networking.k8s.io/v1",
			},
			{
				APIVersion:   "extensions/v1beta1",
				DeprecatedIn: "1.14",
				RemovedIn:    "1.22",
				Replacement:  "networking.k8s.io/v1",
			},
		},
		IntroducedIn: "1.19",
	},
	{
		Kind:         "CronJob",
		PreferredAPI: "batch/v1",
		DeprecatedAPIs: []DeprecatedAPI{
			{
				APIVersion:   "batch/v1beta1",
				DeprecatedIn: "1.21",
				RemovedIn:    "1.25",
				Replacement:  "batch/v1",
			},
		},
		IntroducedIn: "1.21",
	},
	{
		Kind:         "HorizontalPodAutoscaler",
		PreferredAPI: "autoscaling/v2",
		DeprecatedAPIs: []DeprecatedAPI{
			{
				APIVersion:   "autoscaling/v2beta2",
				DeprecatedIn: "1.23",
				RemovedIn:    "1.26",
				Replacement:  "autoscaling/v2",
			},
			{
				APIVersion:   "autoscaling/v2beta1",
				DeprecatedIn: "1.22",
				RemovedIn:    "1.25",
				Replacement:  "autoscaling/v2",
			},
		},
		IntroducedIn: "1.23",
	},
	{
		Kind:         "PodDisruptionBudget",
		PreferredAPI: "policy/v1",
		DeprecatedAPIs: []DeprecatedAPI{
			{
				APIVersion:   "policy/v1beta1",
				DeprecatedIn: "1.21",
				RemovedIn:    "1.25",
				Replacement:  "policy/v1",
			},
		},
		IntroducedIn: "1.21",
	},
	{
		Kind:         "PodSecurityPolicy",
		PreferredAPI: "policy/v1beta1",
		DeprecatedAPIs: []DeprecatedAPI{
			{
				APIVersion:   "policy/v1beta1",
				DeprecatedIn: "1.21",
				RemovedIn:    "1.25",
				Replacement:  "", // No direct replacement, use Pod Security Admission
			},
		},
		IntroducedIn: "1.3",
		RemovedIn:    "1.25",
	},
	{
		Kind:         "RuntimeClass",
		PreferredAPI: "node.k8s.io/v1",
		DeprecatedAPIs: []DeprecatedAPI{
			{
				APIVersion:   "node.k8s.io/v1beta1",
				DeprecatedIn: "1.20",
				RemovedIn:    "1.22",
				Replacement:  "node.k8s.io/v1",
			},
		},
		IntroducedIn: "1.20",
	},
	{
		Kind:         "EndpointSlice",
		PreferredAPI: "discovery.k8s.io/v1",
		DeprecatedAPIs: []DeprecatedAPI{
			{
				APIVersion:   "discovery.k8s.io/v1beta1",
				DeprecatedIn: "1.21",
				RemovedIn:    "1.25",
				Replacement:  "discovery.k8s.io/v1",
			},
		},
		IntroducedIn: "1.21",
	},
	{
		Kind:         "CSIStorageCapacity",
		PreferredAPI: "storage.k8s.io/v1",
		DeprecatedAPIs: []DeprecatedAPI{
			{
				APIVersion:   "storage.k8s.io/v1beta1",
				DeprecatedIn: "1.24",
				RemovedIn:    "1.27",
				Replacement:  "storage.k8s.io/v1",
			},
		},
		IntroducedIn: "1.24",
	},
	{
		Kind:         "FlowSchema",
		PreferredAPI: "flowcontrol.apiserver.k8s.io/v1beta3",
		DeprecatedAPIs: []DeprecatedAPI{
			{
				APIVersion:   "flowcontrol.apiserver.k8s.io/v1beta2",
				DeprecatedIn: "1.26",
				RemovedIn:    "1.29",
				Replacement:  "flowcontrol.apiserver.k8s.io/v1beta3",
			},
			{
				APIVersion:   "flowcontrol.apiserver.k8s.io/v1beta1",
				DeprecatedIn: "1.23",
				RemovedIn:    "1.26",
				Replacement:  "flowcontrol.apiserver.k8s.io/v1beta2",
			},
		},
		IntroducedIn: "1.26",
	},
}

// ParseVersion parses a Kubernetes version string.
func ParseVersion(versionStr string) (*KubernetesVersion, error) {
	// Handle various version formats
	// v1.28.0, 1.28.0, 1.28, v1.28
	versionStr = strings.TrimPrefix(versionStr, "v")
	versionStr = strings.TrimPrefix(versionStr, "V")

	// Remove any suffix like -eks-1234, +k3s1, etc.
	re := regexp.MustCompile(`^(\d+)\.(\d+)(?:\.(\d+))?`)
	matches := re.FindStringSubmatch(versionStr)
	if matches == nil {
		return nil, fmt.Errorf("invalid version format: %s", versionStr)
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", matches[1])
	}

	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %s", matches[2])
	}

	patch := 0
	if len(matches) > 3 && matches[3] != "" {
		patch, err = strconv.Atoi(matches[3])
		if err != nil {
			return nil, fmt.Errorf("invalid patch version: %s", matches[3])
		}
	}

	return &KubernetesVersion{
		Major: major,
		Minor: minor,
		Patch: patch,
		Raw:   versionStr,
	}, nil
}

// String returns the version as a string.
func (v *KubernetesVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Compare compares two versions.
// Returns: -1 if v < other, 0 if v == other, 1 if v > other
func (v *KubernetesVersion) Compare(other *KubernetesVersion) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}
	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}
	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}
	return 0
}

// IsAtLeast checks if the version is at least the specified version.
func (v *KubernetesVersion) IsAtLeast(major, minor int) bool {
	if v.Major > major {
		return true
	}
	if v.Major == major && v.Minor >= minor {
		return true
	}
	return false
}

// NewMapper creates a new version mapper for the target version.
func NewMapper(targetVersion, platform string) (*Mapper, error) {
	var version *KubernetesVersion
	var err error

	if targetVersion != "" {
		version, err = ParseVersion(targetVersion)
		if err != nil {
			return nil, fmt.Errorf("invalid target version: %w", err)
		}
	}

	vm := &Mapper{
		targetVersion: version,
		platform:      platform,
		mappings:      make(map[string]APIVersionMapping),
	}

	// Build mappings from history
	for _, mapping := range apiVersionHistory {
		vm.mappings[mapping.Kind] = mapping
	}

	return vm, nil
}

// NewVersionMapper creates a new version mapper (alias for NewMapper for backward compatibility).
func NewVersionMapper(targetVersion, platform string) (*Mapper, error) {
	return NewMapper(targetVersion, platform)
}

// GetAPIVersion returns the appropriate API version for a resource kind.
func (vm *Mapper) GetAPIVersion(kind string) string {
	// Check if we have version-specific mapping
	if mapping, ok := vm.mappings[kind]; ok && vm.targetVersion != nil {
		return vm.selectAPIVersion(&mapping)
	}

	// Fall back to default
	if api, ok := DefaultAPIVersions[kind]; ok {
		return api
	}

	return ""
}

// selectAPIVersion selects the appropriate API version based on target version.
func (vm *Mapper) selectAPIVersion(mapping *APIVersionMapping) string {
	if vm.targetVersion == nil {
		return mapping.PreferredAPI
	}

	// Check if preferred API is available in target version
	if mapping.IntroducedIn != "" {
		introducedVersion, err := ParseVersion(mapping.IntroducedIn)
		if err == nil && vm.targetVersion.Compare(introducedVersion) < 0 {
			// Target version is older than when preferred API was introduced
			// Find the most recent deprecated API that's still available
			for _, deprecated := range mapping.DeprecatedAPIs {
				if deprecated.RemovedIn == "" {
					return deprecated.APIVersion
				}
				removedVersion, err := ParseVersion(deprecated.RemovedIn)
				if err == nil && vm.targetVersion.Compare(removedVersion) < 0 {
					return deprecated.APIVersion
				}
			}
		}
	}

	return mapping.PreferredAPI
}

// DeprecationResult contains information about deprecated APIs in manifests.
type DeprecationResult struct {
	Kind           string   `json:"kind"`
	Name           string   `json:"name"`
	CurrentAPI     string   `json:"current_api"`
	DeprecatedIn   string   `json:"deprecated_in,omitempty"`
	RemovedIn      string   `json:"removed_in,omitempty"`
	ReplacementAPI string   `json:"replacement_api,omitempty"`
	Severity       string   `json:"severity"` // warning, error
	Message        string   `json:"message"`
	FilePath       string   `json:"file_path,omitempty"`
	Suggestions    []string `json:"suggestions,omitempty"`
}

// CheckDeprecation checks if an API version is deprecated for the target version.
func (vm *Mapper) CheckDeprecation(kind, apiVersion string) *DeprecationResult {
	mapping, ok := vm.mappings[kind]
	if !ok {
		return nil
	}

	for _, deprecated := range mapping.DeprecatedAPIs {
		if deprecated.APIVersion == apiVersion {
			result := &DeprecationResult{
				Kind:           kind,
				CurrentAPI:     apiVersion,
				ReplacementAPI: deprecated.Replacement,
				DeprecatedIn:   deprecated.DeprecatedIn,
				RemovedIn:      deprecated.RemovedIn,
				Suggestions:    []string{},
			}

			// Determine severity
			if vm.targetVersion != nil && deprecated.RemovedIn != "" {
				removedVersion, err := ParseVersion(deprecated.RemovedIn)
				if err == nil && vm.targetVersion.Compare(removedVersion) >= 0 {
					result.Severity = "error"
					result.Message = fmt.Sprintf("%s %s was removed in Kubernetes %s", kind, apiVersion, deprecated.RemovedIn)
				} else {
					result.Severity = "warning"
					result.Message = fmt.Sprintf("%s %s is deprecated (deprecated in %s, removed in %s)", kind, apiVersion, deprecated.DeprecatedIn, deprecated.RemovedIn)
				}
			} else {
				result.Severity = "warning"
				result.Message = fmt.Sprintf("%s %s is deprecated since Kubernetes %s", kind, apiVersion, deprecated.DeprecatedIn)
			}

			if deprecated.Replacement != "" {
				result.Suggestions = append(result.Suggestions, fmt.Sprintf("Use %s instead", deprecated.Replacement))
			} else {
				result.Suggestions = append(result.Suggestions, "This API has no direct replacement. Consider alternative approaches.")
			}

			return result
		}
	}

	return nil
}

// GetAllDeprecatedAPIs returns all known deprecated APIs for the target version.
func (vm *Mapper) GetAllDeprecatedAPIs() []DeprecationResult {
	var results []DeprecationResult

	for kind, mapping := range vm.mappings {
		for _, deprecated := range mapping.DeprecatedAPIs {
			result := &DeprecationResult{
				Kind:           kind,
				CurrentAPI:     deprecated.APIVersion,
				ReplacementAPI: deprecated.Replacement,
				DeprecatedIn:   deprecated.DeprecatedIn,
				RemovedIn:      deprecated.RemovedIn,
			}

			// Determine if it's an error or warning for target version
			if vm.targetVersion != nil && deprecated.RemovedIn != "" {
				removedVersion, err := ParseVersion(deprecated.RemovedIn)
				if err == nil && vm.targetVersion.Compare(removedVersion) >= 0 {
					result.Severity = "error"
					result.Message = fmt.Sprintf("Removed in Kubernetes %s", deprecated.RemovedIn)
				} else {
					result.Severity = "warning"
					result.Message = fmt.Sprintf("Deprecated in %s, removed in %s", deprecated.DeprecatedIn, deprecated.RemovedIn)
				}
			} else {
				result.Severity = "warning"
				result.Message = fmt.Sprintf("Deprecated since Kubernetes %s", deprecated.DeprecatedIn)
			}

			results = append(results, *result)
		}
	}

	return results
}

// OpenShiftToKubernetesVersion maps OpenShift versions to Kubernetes versions.
var OpenShiftToKubernetesVersion = map[string]string{
	"4.16": "1.29",
	"4.15": "1.28",
	"4.14": "1.27",
	"4.13": "1.26",
	"4.12": "1.25",
	"4.11": "1.24",
	"4.10": "1.23",
	"4.9":  "1.22",
	"4.8":  "1.21",
	"4.7":  "1.20",
	"4.6":  "1.19",
}

// GetKubernetesVersionForOpenShift returns the Kubernetes version for an OpenShift version.
func GetKubernetesVersionForOpenShift(openshiftVersion string) (string, bool) {
	// Handle full version strings like 4.14.0
	parts := strings.Split(openshiftVersion, ".")
	if len(parts) >= 2 {
		shortVersion := parts[0] + "." + parts[1]
		if k8sVersion, ok := OpenShiftToKubernetesVersion[shortVersion]; ok {
			return k8sVersion, true
		}
	}
	return "", false
}

// GetOpenShiftVersionForKubernetes returns the OpenShift version for a Kubernetes version.
func GetOpenShiftVersionForKubernetes(k8sVersion string) (string, bool) {
	for ocp, k8s := range OpenShiftToKubernetesVersion {
		if strings.HasPrefix(k8sVersion, k8s) {
			return ocp, true
		}
	}
	return "", false
}

// SupportedVersionRange represents the supported Kubernetes version range.
type SupportedVersionRange struct {
	MinVersion string
	MaxVersion string
}

// GetSupportedVersions returns the supported Kubernetes version range for gitopsi.
func GetSupportedVersions() SupportedVersionRange {
	return SupportedVersionRange{
		MinVersion: "1.21",
		MaxVersion: "1.31",
	}
}

// IsVersionSupported checks if a Kubernetes version is supported.
// Returns whether the version is supported and a message explaining why if not.
func IsVersionSupported(versionStr string) (supported bool, message string) {
	v, err := ParseVersion(versionStr)
	if err != nil {
		return false, err.Error()
	}

	supportedRange := GetSupportedVersions()

	minV, err := ParseVersion(supportedRange.MinVersion)
	if err != nil {
		return false, fmt.Sprintf("internal error: invalid MinVersion: %v", err)
	}

	maxV, err := ParseVersion(supportedRange.MaxVersion)
	if err != nil {
		return false, fmt.Sprintf("internal error: invalid MaxVersion: %v", err)
	}

	if v.Compare(minV) < 0 {
		return false, fmt.Sprintf("version %s is older than minimum supported version %s", versionStr, supportedRange.MinVersion)
	}

	if v.Compare(maxV) > 0 {
		return false, fmt.Sprintf("version %s is newer than maximum supported version %s - some features may not work correctly", versionStr, supportedRange.MaxVersion)
	}

	return true, ""
}
