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
	if IsInstalled() {
		shell := CurrentShell()
		if shell == "zsh" {
			err := AliasZSHConf("gf", "gf")
			if err != nil {
				mlog.Print("your shell is zsh, you might need rename your alias by command `echo alias gf=gf>>~/.zshrc && source ~/.zshrc` to resolve the conflicts between `gf` and `git fetch`")
			}
		}
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

// currentShell returns current shell name
func CurrentShell() string {
	if runtime.GOOS != "windows" {
		osPid := gproc.PPidOS()
		psCmd := fmt.Sprintf("ps -p %d -ocomm=", osPid)
		ps, err := gproc.ShellExec(psCmd)
		if err != nil {
			return ""
		}
		return gstr.Trim(ps)
	}
	return ""
}

// aliasZSHConf append `alias command=dstCommand` to .zshrc if this file exist
func AliasZSHConf(command, dstCommand string) error {
	homePath, err := gfile.Home()
	if err != nil {
		return err
	}
	confPath := fmt.Sprintf("%s/.zshrc", homePath)
	if gfile.Exists(confPath) {
		content := gfile.GetContents(confPath)
		if !gstr.ContainsI(content, "alias gf=gf") {
			if err := gfile.PutContentsAppend(confPath, fmt.Sprintf("alias %s=%s", command, dstCommand)); err != nil {
				return err
			}
		}
	}
	return nil
}
