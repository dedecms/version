package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dedecms/snake"
	"github.com/dedecms/snake/pkg"
	"github.com/dedecms/version/github"
	"github.com/dedecms/version/log"
	"github.com/kenkyu392/go-safe"
	"golang.org/x/text/encoding/simplifiedchinese"
)

var srcdir = "./source/uploads"
var srcrootdir = "./source"
var uplistdir = "./update"
var uplistfile = snake.FS(uplistdir).Add(snake.String(time.Now().Format("20060102")).Add(".file.txt").Get())
var ver = ""

func main() {

	if err := safe.Do(func() error {
		if !uplistfile.Exist() {
			l := log.Start("发版初始化任务完成")
			getSource()
			generateUpdateList()
			l.Done()
		} else {

			reader := bufio.NewReader(os.Stdin)
			for {
				fmt.Print("请输入本次所发版本的版本号: ")
				ver, _ = reader.ReadString('\n')
				ver = strings.Replace(ver, "\n", "", -1)
				if ver != "" {
					break
				}
			}

			l := log.Start("发版任务完成")
			editPackage()
			generatePatch()
			generateVerifys()
			generatePackage()

			// 输出更新日志文件到对应目录
			sl := log.Start("拷贝更新日志文件: ./public/base-v57/utf-8/" + uplistfile.Base())
			uplistfile.Cp("./public/base-v57/utf-8", true)
			sl.Done()

			sl = log.Start("拷贝源码： ./public/base-v57/utf-8/source")
			snake.FS(srcdir).Cp("./public/base-v57/utf-8/source", true)
			sl.Done()

			l.Done()
		}
		return nil
	}); err != nil {
		fmt.Println("错误: ", err)
	}

}

func getSource() {
	l := log.Start("下载最新DedeCMS源码")
	url := github.GetNewTarPackage()
	resp, err := http.Get(url)
	if err != nil {
		l.Err(err)
	}
	defer resp.Body.Close()
	l.Done()

	l = log.Start("解压源码包覆盖至./source")
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
	l.Done()
}

func generateUpdateList() snake.FileSystem {
	l := log.Start("生成更新日志文件: " + uplistfile.Get())
	uplistfile.Write("")
	commits := github.GetUpdateList()
	for _, v := range commits.NewCommits {
		snake.String(v.Filename).ReplaceOne("uploads/", "").Add(", ").Add(v.Message).Ln().Write(uplistfile.Get(), true)
	}
	l.Done()
	return uplistfile
}

func generateVerifys() string {
	l := log.Start("生成系统指纹文件: ./public/base-v57/utf-8/verifys.txt")
	var output = "./public/base-v57/utf-8/verifys.txt"
	snake.FS(output).MkFile()
	for _, v := range snake.FS(srcdir).Find("*.htm", "*.html", "*.js", "*.php") {
		file := snake.String(v).Replace(snake.FS(srcdir).Get(), "..")
		snake.String(file.MD5()).Add("	", snake.FS(v).MD5()).Add("	", file.Get()).Ln().Write(output, true)
	}
	l.Done()
	return output
}

func generatePatch() {
	l := log.Start("输出补丁包: " + fmt.Sprintf("./public/base-v57/package/patch-v57sp2&v57sp1&v57-%s.zip", time.Now().Format("20060102")))
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
	l.Done()
}

func generatePackage() {

	l := log.Start("生成安装包: ./public/base-v57/package/DedeCMS-V5.7-UTF8-SP2.tar.gz")
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
	l.Done()

	l = log.Start("生成安装包hash文件: ./public/base-v57/package/md5.hash.txt")
	snake.FS("./public/base-v57/package/md5.hash.txt").Write(fmt.Sprintf(`jsonCallback({"DedeCMS-V5.7-UTF8-SP2.tar.gz":"%s"});`, patchname.MD5()))
	l.Done()
}

// 修改包信息
func editPackage() {
	l := log.Start("修改源码包中的版本信息")
	changelog := snake.String("DedeCMS ").Add(snake.String(ver).ToUpper().Get()).Add(" 修正功能列表").DrawBox(64, pkg.Box9Slice{
		Top:         "-",
		TopRight:    "",
		Right:       "#",
		BottomRight: "",
		Bottom:      "-",
		BottomLeft:  "#",
		Left:        "#",
		TopLeft:     "#"}).Ln()

	if u, ok := uplistfile.Open(); ok {
		for _, v := range u.String().Lines() {
			massage := snake.String(v).Split(",")
			output := snake.String(massage[1]).Trim(" ").Trim("	").Get()
			if !changelog.Find(output) {
				changelog.Add(output).Ln()
			}
		}
	}
	changelog.Write(snake.FS(srcrootdir).Add("docs").Add("changelog.txt").Get())
	ver := snake.FS(srcrootdir).Add("uploads", "data", "admin", "ver.txt")
	verifies := snake.FS(srcrootdir).Add("uploads", "data", "admin", "verifies.txt")
	ver.Write(time.Now().Format("20060102"))
	verifies.Write(time.Now().Format("20060102"))
	l.Done()
}
