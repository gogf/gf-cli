package release

import (
	"bufio"
	"fmt"
	"os"

	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/os/gproc"
	"github.com/gogf/gf/text/gstr"
	"github.com/gogf/gf/util/gconv"
)

const (
	VERSION_FILE = "version"

	VERSION_TYPE_MAJOR = "major"
	VERSION_TYPE_MINOR = "minor"
	VERSION_TYPE_PATCH = "patch"

	VERSION_SUFFIX_ALPHA   = "alpha"
	VERSION_SUFFIX_BETA    = "beta"
	VERSION_SUFFIX_RC      = "rc"
	VERSION_SUFFIX_RELEASE = "release"

	VERSION_INIT = "v0.0.0"
)

var (
	VERSION_TYPES    = []string{VERSION_TYPE_MAJOR, VERSION_TYPE_MINOR, VERSION_TYPE_PATCH}
	VERSION_TYPE_MAP = map[string]string{
		"s": VERSION_TYPE_MAJOR,
		"f": VERSION_TYPE_MINOR,
		"b": VERSION_TYPE_PATCH,
	}

	VERSION_SUFFIXS = []string{VERSION_SUFFIX_ALPHA, VERSION_SUFFIX_BETA, VERSION_SUFFIX_RC, VERSION_SUFFIX_RELEASE}
)

func Help() {
	mlog.Print(gstr.TrimLeft(`
USAGE
    gf release

OPTION
    -m, --message   commit message
    -t, --type      semver type, eg: major/minor/patch
    -s, --suffix    release suffix, eg: alpha/beta/rc/release
    -p, --push      push to the remote repository

EXAMPLES
    gf release
    gf release -m "say something"
    gf release -t patch
    gf release -s beta
    gf release -m "say something" -t patch -s alpha -p
`))
}

func Run() {
	parser, err := gcmd.Parse(g.MapStrBool{
		"m,message": true,
		"t,type":    true,
		"s,suffix":  true,
		"p,push":    false,
	})
	if err != nil {
		mlog.Fatal(err)
	}

	versionType := getVersionType(parser)
	versionSuffix := getVersionSuffix(parser)

	if err := gitPull(); err != nil {
		mlog.Fatal(err)
	}

	if err := gitFetch(); err != nil {
		mlog.Fatal(err)
	}

	latestVersion, err := getLatestVersion(versionType)
	if err != nil {
		mlog.Fatal(err)
	}

	newVersion := upgradeVersion(latestVersion, versionType, versionSuffix)
	message := getMessage(parser, newVersion)

	if err := gitCommit(message); err != nil {
		mlog.Fatal(err)
	}

	if err := gitTag(newVersion); err != nil {
		mlog.Fatal(err)
	}

	// confirm info
	mlog.Printf("latest version: %s", latestVersion)
	mlog.Printf("new version: %s", newVersion)
	mlog.Printf("message: %s", message)

	if err := gitPush(parser, newVersion); err != nil {
		mlog.Fatal(err)
	}

	mlog.Print("done!")
}

func getVersionType(parser *gcmd.Parser) (versionType string) {
	versionType = parser.GetOpt("type")

	for !gstr.InArray(VERSION_TYPES, versionType) {
		if versionType != "" {
			if v, ok := VERSION_TYPE_MAP[versionType]; ok {
				versionType = v
				break
			}
		}

		versionType = gcmd.Scan("Version: [major(s)/minor(f)/patch(b)]")
	}

	return
}

func getVersionSuffix(parser *gcmd.Parser) (suffix string) {
	suffix = parser.GetOpt("suffix")

	for !gstr.InArray(VERSION_SUFFIXS, suffix) {
		suffix = gcmd.Scan("Suffix: [alpha/beta/rc/release]")
	}

	if suffix == VERSION_SUFFIX_RELEASE {
		suffix = ""
	}

	return
}

func getLatestVersion(versionType string) (string, error) {
	version, err := gitLatestTag()
	if err != nil {
		return "", err
	}

	if version == "" {
		version = VERSION_INIT
	}

	return version, nil
}

func upgradeVersion(latestVersion, versionType, versionSuffix string) (newVersion string) {
	core := gstr.Split(latestVersion, "-")
	core[0] = gstr.TrimLeft(core[0], "v")
	ver := gstr.Split(core[0], ".")

	switch versionType {
	case VERSION_TYPE_MAJOR:
		newVersion = gstr.JoinAny([]interface{}{gconv.Int(ver[0]) + 1, "0", "0"}, ".")
	case VERSION_TYPE_MINOR:
		newVersion = gstr.JoinAny([]interface{}{ver[0], gconv.Int(ver[1]) + 1, "0"}, ".")
	case VERSION_TYPE_PATCH:
		newVersion = gstr.JoinAny([]interface{}{ver[0], ver[1], gconv.Int(ver[2]) + 1}, ".")
	}

	if versionSuffix != "" {
		newVersion += "-" + versionSuffix
	}

	newVersion = "v" + newVersion

	if err := gfile.PutContents(VERSION_FILE, newVersion); err != nil {
		mlog.Fatal(err)
	}

	return
}

func getMessage(parser *gcmd.Parser, version string) (message string) {
	message = parser.GetOpt("message")

	inputReader := bufio.NewReader(os.Stdin)

	for message == "" {
		fmt.Print("Commit Message:")
		message, _ = inputReader.ReadString('\n')
		message = gstr.Trim(message)
	}

	message = "<" + version + "> " + message

	return
}

func gitPull() error {
	pullCmd := "git pull"
	mlog.Printf("execute: %s", pullCmd)
	_, err := gproc.ShellExec(pullCmd)
	return err
}

func gitFetch() error {
	fetchCmd := "git fetch -p"
	mlog.Printf("execute: %s", fetchCmd)
	_, err := gproc.ShellExec(fetchCmd)
	return err
}

func gitCommit(message string) error {
	commitCmd := fmt.Sprintf("git commit -a -m '%s'", message)
	mlog.Printf("execute: %s", commitCmd)
	_, err := gproc.ShellExec(commitCmd)
	return err
}

func gitTag(version string) error {
	tagCmd := fmt.Sprintf("git tag %s", version)
	mlog.Printf("execute: %s", tagCmd)
	_, err := gproc.ShellExec(tagCmd)
	return err
}

func gitLatestTag() (string, error) {
	tagCmd := "git tag | sort -V | tail -1"
	mlog.Printf("execute: %s", tagCmd)
	r, err := gproc.ShellExec(tagCmd)
	if err != nil {
		return "", err
	}
	return gstr.Trim(r), nil
}

func gitPush(parser *gcmd.Parser, version string) error {
	if !parser.ContainsOpt("push") {
		var yes string
		for yes != "y" && yes != "n" {
			yes = gcmd.Scan("Do you want push to remote[y/n]")
		}
		if yes == "n" {
			return nil
		}
	}
	return nil

	pushCmd := fmt.Sprintf("git push origin %s && git push", version)
	mlog.Printf("execute: %s", pushCmd)
	_, err := gproc.ShellExec(pushCmd)
	return err
}
