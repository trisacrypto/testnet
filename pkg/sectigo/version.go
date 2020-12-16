package sectigo

import "fmt"

// Version component constants for the current build.
const (
	VersionMajor = 1
	VersionMinor = 0
	VersionPatch = 0
)

// Version returns the semantic version for the current build.
func Version() string {
	if VersionPatch > 0 {
		return fmt.Sprintf("%d.%d.%d", VersionMajor, VersionMinor, VersionPatch)
	}
	return fmt.Sprintf("%d.%d", VersionMajor, VersionMinor)
}
