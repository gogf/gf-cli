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

type installFolderPath struct {
	path           string
	writable       bool
	binaryFilePath string
	installed      bool
}

// Run does the installation.
func Run() {
	// Ask where to install.
	paths := getInstallBinaryPaths()
	if len(paths) <= 0 {
		mlog.Printf("No path detected, you can manually install gf by copying the binary to path folder.")
		return
	}
	mlog.Printf("Detected paths: ")
	mlog.Printf("%2s|%8s|%9s|%s", "Id", "Writable", "Installed", "Path")

	// Print all paths status and determine the default selectedID value.
	var selectedID int = 0
	for id, aPath := range paths {
		mlog.Printf(
			"%2d|%8t|%9t|%s",
			id, aPath.writable, aPath.installed, aPath.path)
		if aPath.writable && selectedID == 0 {
			selectedID = id
		}
	}

	// Get input and update selectedID.
	input := gcmd.Scanf("Please select install destination [%d]: ", selectedID)
	if input != "" {
		selectedID = gconv.Int(input)
	}

	// Check if out of range.
	if selectedID >= len(paths) || selectedID < 0 {
		mlog.Printf("Invaid install destination Id: %d", selectedID)
		return
	}

	// Get selected destination path.
	dstPath := paths[selectedID]

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

// IsInstalled returns whether the binary is installed.
func IsInstalled() bool {
	paths := getInstallBinaryPaths()
	for _, aPath := range paths {
		if aPath.installed {
			return true
		}
	}
	return false
}

// GetInstallFolderPaths returns the installation folder paths for the binary.
func getInstallFolderPaths() []installFolderPath {

	var folderPaths []installFolderPath

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
func getInstallBinaryPaths() []installFolderPath {
	return getInstallFolderPaths()
}

// Check if path is writable and adds related data to [folderPaths].
func checkPathAndAppendToInstallFolderPath(
	folderPaths *[]installFolderPath,
	path string, binaryFileName string) {

	binaryFilePath := gfile.Join(path, binaryFileName)
	*folderPaths =
		append(
			*folderPaths,
			installFolderPath{
				path:           path,
				writable:       gfile.IsWritable(path),
				binaryFilePath: binaryFilePath,
				installed:      isInstalled(binaryFilePath),
			})

}

// Check if this gf binary path exists.
func isInstalled(path string) bool {
	return gfile.Exists(path)
}
