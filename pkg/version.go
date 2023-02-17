package pkg

import "fmt"

// Version component constants for the current build.
const (
	VersionMajor         = 1
	VersionMinor         = 0
	VersionPatch         = 5
	VersionReleaseLevel  = ""
	VersionReleaseNumber = 0
)

// Set the GitVersion via -ldflags="-X 'github.com/trisacrypto/testnet/pkg.GitVersion=$(git rev-parse --short HEAD)'"
var GitVersion string

// Version returns the semantic version for the current build.
func Version() string {
	var versionCore string

	if VersionReleaseLevel != "" {
		if VersionReleaseNumber > 0 {
			versionCore = fmt.Sprintf("%s-%s%d", versionCore, VersionReleaseLevel, VersionReleaseNumber)
		} else {
			versionCore = fmt.Sprintf("%s-%s", versionCore, VersionReleaseLevel)
		}
	} else if GitVersion != "" {
		versionCore = fmt.Sprintf("%s (%s)", versionCore, GitVersion)
	} else {
		versionCore = fmt.Sprintf("%d.%d.%d", VersionMajor, VersionMinor, VersionPatch)
	}

	return versionCore
}
