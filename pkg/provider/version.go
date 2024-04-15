package provider

import "fmt"

// ProviderVersion field is a SemVer that should indicate the baked-in version
// of the k8s-csi-provider
var ProviderVersion = "0.0"

// TagSuffix field denotes the specific build type for the client. It may
// be replaced by compile-time variables if needed to provide the git
// commit information in the final binary.
// In fixed versions, we don't want the tag to be present
var TagSuffix = "dev"

// FullVersionName is the user-visible aggregation of version and tag
// of this codebase
var FullVersionName = fmt.Sprintf("v%s-%s", ProviderVersion, TagSuffix)
