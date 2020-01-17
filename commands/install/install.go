package install

import (
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/os/genv"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/os/gproc"
	"github.com/gogf/gf/text/gstr"
	"runtime"
)

// Run does the installation.
func Run() {
	// Uninstall the old binary.
	if path := gproc.SearchBinary("gf"); path != "" {
		// Do not delete myself.
		if gfile.SelfPath() != path {
			gfile.Remove(path)
		}
	}
	// Install the new binary.
	path := GetInstallBinaryPath()
	err := gfile.CopyFile(gfile.SelfPath(), path)
	if err != nil {
		mlog.Printf("install gf binary to '%s' failed: %v", path, err)
		mlog.Printf("you can manually install gf by copying the binary to folder: %s", GetInstallFolderPath())
	} else {
		mlog.Printf("gf binary is successfully installed to: %s", path)
	}
}

// IsInstalled returns whether the binary installed.
func IsInstalled() bool {
	return gfile.Exists(GetInstallBinaryPath())
}

// getInstallFolderPath returns the installation folder path for the binary.
func GetInstallFolderPath() string {
	folderPath := ""
	switch runtime.GOOS {
	case "darwin":
		folderPath = "/usr/local/bin"
	default:
		// Search and find the writable directory path.
		envPath := genv.Get("PATH", genv.Get("Path"))
		if gstr.Contains(envPath, ";") {
			for _, v := range gstr.SplitAndTrim(envPath, ";") {
				if gfile.IsWritable(v) {
					return v
				}
			}
		} else if gstr.Contains(envPath, ":") {
			for _, v := range gstr.SplitAndTrim(envPath, ":") {
				if gfile.IsWritable(v) {
					return v
				}
			}
		}
		folderPath = "/usr/local/bin"
	}
	return folderPath
}

// getInstallBinaryPath returns the installation path for the binary.
func GetInstallBinaryPath() string {
	return gfile.Join(GetInstallFolderPath(), "gf"+gfile.Ext(gfile.SelfPath()))
}
