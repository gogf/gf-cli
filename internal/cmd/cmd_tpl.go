package cmd

import (
	"context"

	"github.com/gogf/gf-cli/v2/utility/mlog"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gtag"
	"github.com/gogf/gf/v2/util/gutil"
)

var (
	Tpl = cTpl{}
)

type cTpl struct {
	g.Meta `name:"tpl" brief:"{cTplBrief}" dc:"{cTplDc}"`
}

const (
	cTplBrief = `template parsing and building commands`
	cTplDc    = `
The "tpl" command is used for template parsing and building purpose.
It can parse either template file or folder with multiple types of values support,
like json/xml/yaml/toml/ini.
`
	cTplParseBrief = `parse either template file or folder with multiple types of values`
	cTplParseEg    = `
gf tpl parse -p ./template -v values.json -r
gf tpl parse -p ./template -v values.json -o ./template.parsed
`
	cTplSupportValuesFilePattern = `*.json,*.xml,*.yaml,*.yml,*.toml,*.ini`
)

type (
	cTplParseInput struct {
		g.Meta  `name:"parse" brief:"{cTplParseBrief}" eg:"{cTplParseEg}"`
		Path    string `name:"path"    short:"p" brief:"template file or folder path" v:"required"`
		Values  string `name:"values"  short:"v" brief:"template values file/folder, support file types like: json/xml/yaml/toml/ini" v:"required"`
		Output  string `name:"output"  short:"o" brief:"output file/folder path"`
		Replace bool   `name:"replace" short:"r" brief:"replace original files" orphan:"true"`
	}
	cTplParseOutput struct{}
)

func init() {
	gtag.Sets(g.MapStrStr{
		`cTplBrief`:      cTplBrief,
		`cTplDc`:         cTplDc,
		`cTplParseEg`:    cTplParseEg,
		`cTplParseBrief`: cTplParseBrief,
	})
}

func (c *cTpl) Parse(ctx context.Context, in cTplParseInput) (out *cTplParseOutput, err error) {
	if in.Output == "" && in.Replace == false {
		return nil, gerror.New(`parameter output and replace should not be both empty`)
	}
	valuesMap, err := c.loadValues(ctx, in.Values)
	if err != nil {
		return nil, err
	}
	if len(valuesMap) == 0 {
		return nil, gerror.Newf(`empty values loaded from values file/folder "%s"`, in.Values)
	}
	err = c.parsePath(ctx, in.Path, in.Output, valuesMap, in.Replace)
	if err == nil {
		mlog.Print("done!")
	}
	return
}

func (c *cTpl) parsePath(ctx context.Context, path string, output string, values g.Map, replace bool) (err error) {
	if !gfile.Exists(path) {
		return gerror.Newf(`path "%s" does not exist`, path)
	}
	var (
		files        []string
		relativePath string
		outputPath   string
	)
	path = gfile.RealPath(path)
	if gfile.IsDir(path) {
		files, err = gfile.ScanDirFile(path, "*", true)
		if err != nil {
			return err
		}
		for _, file := range files {
			relativePath = gstr.Replace(file, path, "")
			if output != "" {
				outputPath = gfile.Join(output, relativePath)
			}
			if err = c.parseFile(ctx, file, outputPath, values, replace); err != nil {
				return
			}
		}
		return
	}
	if output != "" {
		outputPath = output
	}
	err = c.parseFile(ctx, path, outputPath, values, replace)
	return
}

func (c *cTpl) parseFile(ctx context.Context, file string, output string, values g.Map, replace bool) (err error) {
	output = gstr.ReplaceByMap(output, g.MapStrStr{
		`\\`: `\`,
		`//`: `/`,
	})
	content, err := g.View().Parse(ctx, file, values)
	if err != nil {
		return err
	}
	if output != "" {
		mlog.Printf(`parse file "%s" to "%s"`, file, output)
		return gfile.PutContents(output, content)
	}
	if replace {
		mlog.Printf(`parse and replace file "%s"`, file)
		return gfile.PutContents(file, content)
	}
	return nil
}

func (c *cTpl) loadValues(ctx context.Context, valuesPath string) (data g.Map, err error) {
	if !gfile.Exists(valuesPath) {
		return nil, gerror.Newf(`values file/folder "%s" does not exist`, valuesPath)
	}
	var j *gjson.Json
	if gfile.IsDir(valuesPath) {
		var valueFiles []string
		valueFiles, err = gfile.ScanDirFile(valuesPath, cTplSupportValuesFilePattern, true)
		if err != nil {
			return nil, err
		}
		data = make(g.Map)
		for _, file := range valueFiles {
			if j, err = gjson.Load(file); err != nil {
				return nil, err
			}
			gutil.MapMerge(data, j.Map())
		}
		return
	}
	if j, err = gjson.Load(valuesPath); err != nil {
		return nil, err
	}
	data = j.Map()
	return
}
