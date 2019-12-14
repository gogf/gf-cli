package install

import (
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/os/gfile"
	"runtime"
)

// Run does the installation.
func Run() {
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
	folderPath := "/usr/local/bin"
	if "windows" == runtime.GOOS {
		folderPath = "C:\\Windows"
	}
	return folderPath
}

// getInstallBinaryPath returns the installation path for the binary.
func GetInstallBinaryPath() string {
	return gfile.Join(GetInstallFolderPath(), "gf"+gfile.Ext(gfile.SelfPath()))
}
