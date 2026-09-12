//go:build !windows

package license

// getOSLicenseMetadata fallback cho non-Windows build
func getOSLicenseMetadata() LicenseRecord {
	return LicenseRecord{
		SoftwareName:     "Microsoft Windows",
		Version:          "(chưa xác định - môi trường không phải Windows)",
		ProductID:        "",
		LicensingChannel: "Unknown",
		BIOSOEMKey:       "",
		Legal:            true,
	}
}

func getOfficeLicenseMetadata() []LicenseRecord { return nil }

func readKMSRegistryServer() string { return "" }

func ScanCrackToolsOnWindows(_ []CrackSignature) []CrackHit { return nil }
