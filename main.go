package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dedecms/snake"
	"github.com/dedecms/version/github"
	"golang.org/x/text/encoding/simplifiedchinese"
)

var srcdir = "./source/uploads"
var srcrootdir = "./source"
var uplistdir = "./update"
var uplistfile = snake.FS(uplistdir).Add(snake.String(time.Now().Format("20060102")).Add(".file.txt").Get())

func main() {

	getSrc()

	// if !uplistfile.Exist() {
	generateUpdateList()
	// }

	generatePatch()
	generateVerifys()
	generatePackage()

	// 输出更新日志文件到对应目录
	uplistfile.Cp("./public/base-v57/utf-8", true)
	snake.FS(srcdir).Cp("./public/base-v57/utf-8/source", true)

}

func getSrc() {
	url := github.GetNewTarPackage()
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	gr, _ := gzip.NewReader(resp.Body)
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return
			}
		}
		name := snake.FS(hdr.Name).ReplaceRoot(srcrootdir)
		if !snake.String(name.Get()).Find(".gitignore", true) && !snake.String(name.Get()).Find(".index", true) {
			if hdr.Typeflag == tar.TypeDir && !name.Exist() {
				name.MkDir()
			} else if hdr.Typeflag == tar.TypeReg {
				var buf strings.Builder
				io.Copy(&buf, tr)
				name.Write(buf.String())
			}
		}

	}

	gr.Close()
}

func generateUpdateList() snake.FileSystem {
	uplistfile.Write("")
	commits := github.GetUpdateList()
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
	patchname := snake.FS("./public/base-v57/package").Add(fmt.Sprintf("patch-v57sp2&v57sp1&v57-%s.tar.gz", time.Now().Format("20060102"))).Get()
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

func generatePackage() {
	// 输出安装
	patchname := snake.FS("./public/base-v57/package").Add("DedeCMS-V5.7-UTF8-SP2.tar.gz")
	zip := snake.Tar(patchname.Get())
	for _, v := range snake.FS(srcrootdir).Find("*") {
		utf8 := snake.FS(snake.String(v).Remove(snake.FS(srcrootdir).Get()).Trim("/").Trim(`\`).Get())
		if src, ok := snake.FS(v).Open(); ok {
			body := src.Byte()
			stat, _ := src.Get().Stat()
			src.Close()
			zip.Add(utf8.Get(), stat, body)
		}
	}
	zip.Close()
	snake.FS("./public/base-v57/package/md5.hash.txt").Write(fmt.Sprintf(`jsonCallback({"DedeCMS-V5.7-UTF8-SP2.tar.gz":"%s"});`, patchname.MD5()))
}
