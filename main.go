package main

import (
	"fmt"
	"time"

	"github.com/dedecms/snake"
	"github.com/dedecms/version/github"
	"golang.org/x/text/encoding/simplifiedchinese"
)

var srcdir = "./src/uploads"
var uplistdir = "./update"
var uplistfile = snake.FS(uplistdir).Add(snake.String(time.Now().Format("20060102")).Add(".file.txt").Get())

func main() {

	if !uplistfile.Exist() {
		generateUpdateList()
	}

	generatePatch()
	generateVerifys()

	// 输出更新日志文件到对应目录
	uplistfile.Cp("./public/base-v57/utf-8", true)
	snake.FS(srcdir).Cp("./public/base-v57/utf-8/source", true)

}

func generateUpdateList() snake.FileSystem {
	uplistfile.Write("")
	commits := github.GetNewCommit()
	for _, v := range commits.NewCommits {
		snake.String(v.Filename).ReplaceOne("uploads/", "").Add(", ").Add(v.Message).Ln().Write(uplistfile.Get(), true)
	}
	return uplistfile
}

func generateVerifys() string {
	var output = "./public/base-v57/utf-8/verifys.txt"
	snake.FS(output).MkFile()
	for _, v := range snake.FS(srcdir).Find("*.htm", "*.html", "*.js", "*.php") {
		file := snake.String(v).Replace(snake.FS(srcdir).Get(), "..")
		snake.String(file.MD5()).Add("	", snake.FS(v).MD5()).Add("	", file.Get()).Ln().Write(output, true)
	}
	return output
}

func generatePatch() {
	// 输出补丁包
	patchname := snake.FS("./public/base-v57/package").Add(fmt.Sprintf("patch-v57sp2&v57sp1&v57-%s.zip", time.Now().Format("20060102"))).Get()
	zip := snake.Zip(patchname)

	if src, ok := uplistfile.Open(); ok {
		body := src.Byte()
		src.Close()
		zip.Add(uplistfile.Base(), body)
	}

	if f, ok := uplistfile.Open(); ok {
		for _, v := range f.String().Lines() {
			item := snake.String(v).Split(",")
			utf8 := snake.FS("utf-8").Add(snake.String(item[0]).Remove("uploads/").Get())
			gbk := snake.FS("gb2312").Add(snake.String(item[0]).Remove("uploads/").Get())

			if src, ok := snake.FS(srcdir).Add(item[0]).Open(); ok {
				body := src.Byte()
				src.Close()
				zip.Add(utf8.Get(), body)
				if gbkbody, err := simplifiedchinese.GBK.NewEncoder().Bytes(body); err == nil {
					zip.Add(gbk.Get(), gbkbody)
				}
			}

		}
	}
	zip.Close()
}
