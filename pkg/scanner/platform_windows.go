//go:build windows

package scanner

// NewPlatformScanner returns a Windows scanner with default options.
func NewPlatformScanner() Scanner {
	return NewWindowsScanner()
}

// NewPlatformScannerWithOptions returns a Windows scanner with custom options.
func NewPlatformScannerWithOptions(opts Options) Scanner {
	return NewWindowsScannerWithOptions(opts)
}
