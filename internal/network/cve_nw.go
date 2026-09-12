//go:build !windows

package network

func getInstalledSoftware() []string {
	return []string{
		"windows smb 2017",
		"windows rdp 2019",
	}
}
