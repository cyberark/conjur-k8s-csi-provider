package version

import "fmt"

// ProviderVersion is a SemVer that indicates the baked-in version.
var ProviderVersion = "0.0"

// TagSuffix denotes the specific build type for the client.
var TagSuffix = "dev"

// FullVersionName returns the user-visible aggregation of version and tag.
func FullVersionName() string {
	return fmt.Sprintf("v%s-%s", ProviderVersion, TagSuffix)
}

