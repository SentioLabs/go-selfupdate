package selfupdate

import (
	"strings"

	"golang.org/x/mod/semver"
)

// devVersion is what development builds normalize to. It sorts below every
// real release so a dev binary is always offered an update.
const devVersion = "v0.0.0-dev"

// NormalizeVersion adds a leading v and maps "" and "dev" to v0.0.0-dev.
func NormalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	if v == "" || v == "dev" {
		return devVersion
	}
	if !strings.HasPrefix(v, "v") {
		return "v" + v
	}
	return v
}

// Compare is semver.Compare over normalized versions. The result is 0 when
// current and latest are equal, positive when latest is newer (an update is
// available) and negative when current is newer than latest.
func Compare(current, latest string) int {
	return semver.Compare(NormalizeVersion(latest), NormalizeVersion(current))
}
