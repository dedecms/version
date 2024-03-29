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
	"github.com/dedecms/version/encode"
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
			updateVerFile()
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

			sl := log.Start("拷贝GBK源码： ./public/base-v57/gb2312/source")
			copyGBK()
			sl.Done()

			generateGBKPackage()

			sl = log.Start("拷贝GBK更新日志文件: ./public/base-v57/gb2312/" + uplistfile.Base())
			if f, ok := uplistfile.Open(); ok {
				bytes := f.Byte()
				f.Close()
				if gbk, err := simplifiedchinese.GBK.NewEncoder().Bytes(bytes); err == nil {
					snake.FS("./public/base-v57/gb2312", uplistfile.Base()).ByteWriter(gbk)
				}
			}
			sl.Done()

			generateGBKVerifys()

			sl = log.Start("拷贝UTF-8源码： ./public/base-v57/utf-8/source")
			snake.FS(srcdir).Cp("./public/base-v57/utf-8/source", true)
			sl.Done()

			// generateUTF8Package()
			generateUTF8PackageZip()

			// 输出更新日志文件到对应目录
			sl = log.Start("拷贝UTF-8更新日志文件: ./public/base-v57/utf-8/" + uplistfile.Base())
			uplistfile.Cp("./public/base-v57/utf-8", true)
			sl.Done()

			// 输出更新日志文件到对应目录
			sl = log.Start("输出最后更新文件: ./public/base-v57/latest.txt")
			snake.FS("./public/base-v57/latest.txt").Write(snake.String(time.Now().Unix()).Get())
			sl.Done()

			generateVerifys()

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

func updateVerFile() {
	l := log.Start("修正版本号。")
	snake.FS("./source/uploads/data/admin/ver.txt").Write(time.Now().Format("20060102"))
	snake.FS("./source/uploads/data/admin/verifies.txt").Write(time.Now().Format("20060102"))
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
	l := log.Start("生成UTF-8系统指纹文件: ./public/base-v57/utf-8/verifys.txt")
	var output = "./public/base-v57/utf-8/verifys.txt"
	snake.FS(output).MkFile()
	for _, v := range snake.FS(srcdir).Find("*.htm", "*.html", "*.js", "*.php") {
		file := snake.String(v).Replace(snake.FS(srcdir).Get(), "..")
		snake.String(file.MD5()).Add("	", snake.FS(v).MD5()).Add("	", file.Get()).Ln().Write(output, true)
	}
	l.Done()
	return output
}

func generateGBKVerifys() string {
	l := log.Start("生成GBK系统指纹文件: ./public/base-v57/gb2312/verifys.txt")
	var output = "./public/base-v57/gb2312/verifys.txt"
	snake.FS(output).MkFile()
	src := snake.FS("./public/base-v57/gb2312/source")
	for _, v := range src.Find("*.htm", "*.html", "*.js", "*.php") {
		file := snake.String(v).Replace(src.Get(), "..")
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
		stat, _ := src.Get().Stat()
		src.Close()
		zip.Add(uplistfile.Base(), stat, body)
	}

	if f, ok := uplistfile.Open(); ok {
		for _, v := range f.String().Lines() {
			item := snake.String(v).Split(",")
			utf8 := snake.FS("utf-8").Add(snake.String(item[0]).Remove("uploads/").Get())
			gbk := snake.FS("gb2312").Add(snake.String(item[0]).Remove("uploads/").Get())
			src := snake.FS(srcdir).Add(item[0])
			if file, ok := src.Open(); ok {
				body := file.Byte()
				stat, _ := file.Get().Stat()
				file.Close()
				zip.Add(utf8.Get(), stat, body)

				if !snake.String(src.Get()).Find("ckeditor", true) ||
					(snake.String(src.Get()).Find("ckeditor", true) && snake.String(src.Ext()).ExistSlice([]string{".php"})) ||
					snake.String(src.Get()).Find("ckeditor/lang/zh-cn.js") ||
					snake.String(snake.FS(v).Get()).Find("ckeditor/plugins/multipic/plugin.js") ||
					snake.String(snake.FS(v).Get()).Find("ckeditor/plugins/dedepage/plugin.js") {
					if snake.String(src.Ext()).ExistSlice([]string{".html", ".htm", ".php", ".txt", ".xml", ".js", ".css", ".inc"}) {
						body = getGBKbyte(src.Get(), string(body))
						body, _ = simplifiedchinese.GBK.NewEncoder().Bytes(body)
					}
				}

				zip.Add(gbk.Get(), stat, body)
			}

		}
	}
	zip.Close()
	l.Done()
}

func generateUTF8Package() {

	l := log.Start("生成UTF-8安装包: ./public/base-v57/package/DedeCMS-" + ver + "-UTF8.tar.bz2")
	// 输出安装
	patchname := snake.FS("./public/base-v57/package").Add("DedeCMS-" + ver + "-UTF8.tar.bz2")
	zip := snake.Tar(patchname.Get())
	for _, v := range snake.FS(srcrootdir).Find("*") {
		utf8 := snake.FS(v).ReplaceRoot("")
		if src, ok := snake.FS(v).Open(); ok {
			body := src.Byte()
			stat, _ := src.Get().Stat()
			zip.Add(utf8.Get(), stat, body)
			src.Close()
		}
	}
	zip.Close()
	l.Done()

	l = log.Start("生成UTF-8安装包hash文件: ./public/base-v57/package/md5.hash.txt")
	snake.FS("./public/base-v57/package/md5.hash.txt").Write(fmt.Sprintf(`jsonCallback({"DedeCMS-%s-UTF8.tar.bz2":"%s"});`, ver, patchname.MD5()))
	l.Done()
}

func generateUTF8PackageZip() {
	l := log.Start("生成UTF-8安装包: ./public/base-v57/package/DedeCMS-" + ver + "-UTF8.zip")
	// 输出安装
	patchname := snake.FS("./public/base-v57/package").Add("DedeCMS-" + ver + "-UTF8.zip")
	zip := snake.Zip(patchname.Get())
	for _, v := range snake.FS(srcrootdir).Find("*") {
		utf8 := snake.FS(v).ReplaceRoot("")
		if src, ok := snake.FS(v).Open(); ok {
			body := src.Byte()
			stat, _ := src.Get().Stat()
			zip.Add(utf8.Get(), stat, body)
			src.Close()
		}
	}
	zip.Close()
	l.Done()

	l = log.Start("生成UTF-8安装包hash文件: ./public/base-v57/package/md5.hash.txt")
	snake.FS("./public/base-v57/package/md5.hash.txt").Write(fmt.Sprintf(`jsonCallback({"DedeCMS-%s-UTF8.zip":"%s"});`, ver, patchname.MD5()))
	l.Done()
}

func generateGBKPackage() {

	l := log.Start("生成GBK安装包: ./public/base-v57/package/DedeCMS-" + ver + "-GBK.tar.bz2")
	// 输出安装
	patchname := snake.FS("./public/base-v57/package").Add("DedeCMS-" + ver + "-GBK.tar.bz2")
	zip := snake.Tar(patchname.Get())
	for _, v := range snake.FS(srcrootdir).Find("*") {
		gbk := snake.FS(v).ReplaceRoot("")
		openfile := snake.FS(v)
		if src, ok := openfile.Open(); ok {
			bytes := src.Byte()
			stat, _ := src.Get().Stat()
			src.Close()
			if openfile.IsFile() {
				if !snake.String(snake.FS(v)).Find("ckeditor", true) ||
					(snake.String(snake.FS(v)).Find("ckeditor", true) && snake.String(snake.FS(v).Ext()).ExistSlice([]string{".php"})) ||
					snake.String(snake.FS(v).Get()).Find("ckeditor/lang/zh-cn.js") ||
					snake.String(snake.FS(v).Get()).Find("ckeditor/plugins/multipic/plugin.js") ||
					snake.String(snake.FS(v).Get()).Find("ckeditor/plugins/dedepage/plugin.js") {
					if snake.String(snake.FS(v).Ext()).ExistSlice([]string{".html", ".htm", ".php", ".txt", ".xml", ".js", ".css", ".inc"}) {
						bytes = getGBKbyte(v, string(bytes))
						gbkbody, _ := simplifiedchinese.GBK.NewEncoder().Bytes(bytes)
						bytes = gbkbody
					}
				}
			}

			zip.Add(gbk.Get(), stat, bytes)
		}
	}
	zip.Close()
	l.Done()

}

func copyGBK() {
	outdir := snake.FS("public/base-v57/gb2312/source")
	for _, v := range snake.FS(srcdir).Find("*") {
		outfile := snake.String(v).ReplaceOne(snake.FS(srcdir).Get(), outdir.Get())
		if snake.FS(v).IsDir() {
			snake.FS(outfile.Get()).MkDir()
		}
		if snake.FS(v).IsFile() {
			f, _ := snake.FS(v).Open()
			bytes := f.Byte()
			f.Close()
			if !snake.String(snake.FS(v).Get()).Find("ckeditor", true) ||
				(snake.String(snake.FS(v).Get()).Find("ckeditor", true) && snake.String(snake.FS(v).Ext()).ExistSlice([]string{".php"})) ||
				snake.String(snake.FS(v).Get()).Find("ckeditor/lang/zh-cn.js") ||
				snake.String(snake.FS(v).Get()).Find("ckeditor/plugins/multipic/plugin.js") ||
				snake.String(snake.FS(v).Get()).Find("ckeditor/plugins/dedepage/plugin.js") {
				if snake.String(snake.FS(v).Ext()).ExistSlice([]string{".html", ".htm", ".php", ".txt", ".xml", ".js", ".css", ".inc"}) {
					utf8, _ := encode.GetEncoding(bytes)
					body := getGBKbyte(v, utf8.Text())
					bytes, _ = simplifiedchinese.GBK.NewEncoder().Bytes(body)
				}
			}

			snake.FS(outfile.Get()).ByteWriter(bytes)
		}
	}
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
			if !changelog.Find(output, true) {
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

func getGBKbyte(v, body string) []byte {
	bytes := []byte(body)
	if snake.String(snake.FS(v).Ext()).ExistSlice([]string{".html", ".htm"}) {
		bytes = snake.String(body).
			Replace(`(<meta ((http-equiv|content).*(http-equiv|content)|).*charset.*=.*)(?i)(utf-8|big5)(.*>)`, "${1}gb2312${6}"). // 替换 *.HTML, *.HTM META CHARSET
			Byte()
	} else if snake.String(snake.FS(v).Ext()).ExistSlice([]string{".xml"}) {
		bytes = snake.String(body).
			Replace(`(lang(.*|)=(.*|))(?i)(utf-8|big5)`, "${1}gb2312"). // 替换 *.XML LANG
			Byte()
	} else if snake.String(snake.FS(v).Ext()).ExistSlice([]string{".js"}) {
		bytes = snake.String(body).
			Replace(`((.*|)this.sendlang(.*|)= '(.*|))(?i)(utf-8|big5)`, "${1}gb2312"). // 替换 *.JS this.sendlang
			Byte()
	} else if snake.String(snake.FS(v).Ext()).ExistSlice([]string{".css"}) {
		bytes = snake.String(body).
			Replace(`((.*|)@charset.*("|'))(?i)(utf-8|big5)`, "${1}gb2312"). // 替换 *.CSS charset
			Byte()
	} else if snake.String(snake.FS(v).Ext()).ExistSlice([]string{".php"}) {
		bytes = snake.String(body).
			Replace(`((.*|)\$cfg_db_language(.*)=( |	)("|'))(?i)(utf8)`, "${1}gbk").
			Replace(`((.*|)\$cfg_soft_lang(.*)=( |	)("|'))(?i)(utf-8)`, "${1}gb2312").
			Replace(`((.*|)\$cfg_version(.*)=( |	)("|')V(.*)_)(?i)(utf8|gb2312|big5)`, "${1}GBK").
			Replace(`((.*|)\$s_lang(.*)=(.*)("|'))(?i)(utf-8)`, "${1}gb2312").
			Replace(`((.*|)\$verMsg(.*)=(.*)("|').*V5.7.*)(?i)(utf8)`, "${1}GBK").
			Replace(`((.*|)\$dfDbname(.*)=(.*)("|').*dedecmsv57.*)(?i)(utf8)`, "${1}gbk").
			Replace(`((.*|)return cn_substr_utf8(.*)=(.*)("|')V(.*)_)(?i)(utf8|gb2312|big5)`, "${1}GBK").
			Byte()
	}

	return bytes
}
