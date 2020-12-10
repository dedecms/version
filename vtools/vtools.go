package vtools

import (
	"path/filepath"
	"regexp"

	"github.com/mycalf/util"
)

func Run(filename string) bool {
	file := util.OS(filename)
	run := false
	for _, v := range Conf.Suffix {
		if file.Suffix() == v {
			run = true
			replaces(filename)
		}
	}

	if run == false {
		for _, v := range Conf.Charset {
			p := filepath.Join("./out", v, filename)
			if err := file.Copy(p); err == nil {
				run = true
			} else {
				Warning.Println(filename + ": " + err.Error())
			}
		}
	}

	return run
}

// Run 根据替换规则，替换并生成新的文件。
func replaces(filename string) bool {
	// if src, ok := util.OS(filename).Cat(); ok {
	if src, ok := util.OS(filename).Read(); ok {

		for k, v := range Conf.Regexp {
			r := regexp.MustCompile(k)
			src = r.ReplaceAllString(src, v)
		}

		// src = util.Text(src).Parse(Conf.Parse)

		for k, v := range Conf.Replaces {
			src = util.Text(src).Replace(k, v)
		}

		if src != "" {
			for _, v := range Conf.Charset {
				util.OS("./out").Add(v).Add(filename).WriteCharset(src, v)
			}
		}
		return true
	}
	Warning.Println(filename + ": 转换失败!")
	return false
}
