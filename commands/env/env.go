package env

import (
	"bytes"
	"github.com/gogf/gf-cli/library/mlog"
	"github.com/gogf/gf/os/gproc"
	"github.com/gogf/gf/text/gregex"
	"github.com/gogf/gf/text/gstr"
	"github.com/olekukonko/tablewriter"
)

func Run() {
	result, err := gproc.ShellExec("go env")
	if err != nil {
		mlog.Fatal(err)
	}
	if result == "" {
		mlog.Fatal(`retrieving Golang environment variables failed, did you install Golang?`)
	}
	var (
		lines  = gstr.Split(result, "\n")
		buffer = bytes.NewBuffer(nil)
	)
	array := make([][]string, 0)
	for _, line := range lines {
		line = gstr.Trim(line)
		if line == "" {
			continue
		}
		match, _ := gregex.MatchString(`(.+?)=(.+)`, line)
		if len(match) < 3 {
			mlog.Fatalf(`invalid Golang environment variable: "%s"`, line)
		}
		array = append(array, []string{match[1], match[2]})
	}
	tw := tablewriter.NewWriter(buffer)
	tw.AppendBulk(array)
	tw.Render()
	mlog.Print(buffer.String())
}
