package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/dedecms/snake"
	"golang.org/x/text/encoding/simplifiedchinese"
)

var srcdir = "./src/uploads"

func main() {

	output := "./public/base-v57/utf-8/verifys.txt"

	// 输出校验文件
	snake.FS(output).MkFile()
	for _, v := range snake.FS(srcdir).Find("*.htm", "*.html", "*.js", "*.php") {
		file := snake.String(v).Replace(snake.FS(srcdir).Get(), "..")
		snake.String(file.MD5()).Add("	", snake.FS(v).MD5()).Add("	", file.Get()).Ln().Write(output, true)
	}

	// 输出补丁包
	updatesrcfile := snake.FS("./update/20210806.file.txt")
	patchname := snake.FS("./public/base-v57/package").Add(fmt.Sprintf("patch-v57sp2&v57sp1&v57-%s", time.Now().Format("20060102"))).Get()
	updatesrcfile.Cp(patchname, true)

	if f, ok := updatesrcfile.Open(); ok {
		for _, v := range f.String().Lines() {
			p := snake.String(v).Split(",")

			dst := snake.FS(patchname).Add("utf-8").Add(snake.String(p[0]).Remove("uploads/").Get())
			gbk := snake.FS(patchname).Add("gb2312").Add(snake.String(p[0]).Remove("uploads/").Get())
			if f, ok := dst.MkFile(); ok {
				defer f.Get().Close()
				if s, ok := snake.FS(srcdir).Add(p[0]).Open(); ok {
					defer s.Get().Close()
					io.Copy(f.Get(), s.Get())
					s.Get()
				}
			}
			b, err := ioutil.ReadFile(snake.FS(srcdir).Add(p[0]).Get())
			if err != nil {
				fmt.Print(err)
			}
			g, _ := simplifiedchinese.GBK.NewEncoder().Bytes(b)
			gbk.ByteWriter(g)
		}
	}

	// 输出更新日志文件到对应目录
	updatesrcfile.Cp("./public/base-v57/utf-8", true)
	snake.FS(srcdir).Cp("./public/base-v57/utf-8/source", true)

}
