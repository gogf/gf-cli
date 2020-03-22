package install

import (
	"runtime"

	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/os/genv"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/text/gstr"
	"github.com/gogf/gf/util/gconv"
)

type InstallFolderPath struct {
	path           string
	writable       bool
	binaryFilePath string
	installed      bool
}

// Run does the installation.
func Run() {
	// Ask where to install.
	paths := GetInstallBinaryPaths()
	if len(paths) <= 0 {
		mlog.Printf("No path detected, you can manually install gf by copying the binary to path folder.")
		return
	}
	mlog.Printf("Detected paths: ")
	mlog.Printf("%2s|%8s|%9s|%s", "Id", "Writable", "Installed", "Path")
	for id, aPath := range paths {
		mlog.Printf(
			"%2d|%8t|%9t|%s",
			id, aPath.writable, aPath.installed, aPath.path)
	}
	id := gconv.Int(gcmd.Scanf("Please select install destination [0]: "))

	// Check if out of range.
	if id >= len(paths) || id < 0 {
		mlog.Printf("Invaid install destination Id: %d", id)
		return
	}
	dstPath := paths[id]

	// Install the new binary.
	err := gfile.CopyFile(gfile.SelfPath(), dstPath.binaryFilePath)
	if err != nil {
		mlog.Printf("Install gf binary to '%s' failed: %v", dstPath.path, err)
		mlog.Printf("You can manually install gf by copying the binary to folder: %s", dstPath.path)
	} else {
		mlog.Printf("gf binary is successfully installed to: %s", dstPath.path)
	}

	// Uninstall the old binary.
	for _, aPath := range paths {
		// Do not delete myself.
		if aPath.binaryFilePath != "" &&
			aPath.binaryFilePath != dstPath.binaryFilePath &&
			gfile.SelfPath() != aPath.binaryFilePath {
			gfile.Remove(aPath.binaryFilePath)
		}
	}
}

// IsInstalled returns whether the binary installed.
func IsInstalled() bool {
	paths := GetInstallBinaryPaths()
	for _, aPath := range paths {
		if aPath.installed {
			return true
		}
	}
	return false
}

// GetInstallFolderPaths returns the installation folder paths for the binary.
func GetInstallFolderPaths() []InstallFolderPath {

	var folderPaths []InstallFolderPath

	// Pre generate binaryFileName.
	binaryFileName := "gf" + gfile.Ext(gfile.SelfPath())

	switch runtime.GOOS {
	case "darwin":
		checkPathAndAppendToInstallFolderPath(
			&folderPaths, "/usr/local/bin", binaryFileName)
	default:
		// Search and find the writable directory path.
		envPath := genv.Get("PATH", genv.Get("Path"))
		if gstr.Contains(envPath, ";") {
			for _, v := range gstr.SplitAndTrim(envPath, ";") {
				checkPathAndAppendToInstallFolderPath(
					&folderPaths, v, binaryFileName)
			}
		} else if gstr.Contains(envPath, ":") {
			for _, v := range gstr.SplitAndTrim(envPath, ":") {
				checkPathAndAppendToInstallFolderPath(
					&folderPaths, v, binaryFileName)
			}
		} else if envPath != "" {
			checkPathAndAppendToInstallFolderPath(
				&folderPaths, envPath, binaryFileName)
		} else {
			checkPathAndAppendToInstallFolderPath(
				&folderPaths, "/usr/local/bin", binaryFileName)
		}
	}

	return folderPaths
}

// GetInstallBinaryPaths returns the installation path for the binary.
func GetInstallBinaryPaths() []InstallFolderPath {
	return GetInstallFolderPaths()
}

// Check if path is writable and adds related data to [folderPaths].
func checkPathAndAppendToInstallFolderPath(
	folderPaths *[]InstallFolderPath,
	path string, binaryFileName string) {

	binaryFilePath := gfile.Join(path, binaryFileName)

	if gfile.IsWritable(path) {
		*folderPaths =
			append(
				*folderPaths,
				InstallFolderPath{
					path:           path,
					writable:       true,
					binaryFilePath: binaryFilePath,
					installed:      isInstalled(binaryFilePath),
				})
	} else {
		*folderPaths =
			append(
				*folderPaths,
				InstallFolderPath{
					path:           path,
					writable:       false,
					binaryFilePath: binaryFilePath,
					installed:      isInstalled(binaryFilePath),
				})
	}
}

// Check if this gf binary path exists.
func isInstalled(path string) bool {
	return gfile.Exists(path)
}
