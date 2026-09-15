//go:build !windows

package hardware

// RecentFilesResult stub cho non-Windows
type RecentFilesResult struct {
	RecentFilesCount   int
	SuspiciousRecent   []string
	ExternalDriveRefs  []string
	UNCPathRefs        []string
	Summary            string
}

// scanRecentFilesAndJumpLists stub cho non-Windows
func scanRecentFilesAndJumpLists() RecentFilesResult {
	return RecentFilesResult{}
}
