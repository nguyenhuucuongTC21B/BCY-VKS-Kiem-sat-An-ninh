//go:build !windows

package license

type InstalledSoftware struct {
	DisplayName    string
	DisplayVersion string
	Publisher      string
	ProductID      string
	InstallDate    string
	InstallLocation string
	UninstallString string
}

func scanInstalledSoftware() []InstalledSoftware { return nil }

func isAuditWorthy(name, publisher string) bool { return false }

func isLikelyLegalByDefault(name, publisher string) bool { return true }
