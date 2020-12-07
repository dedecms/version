package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mycalf/util"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// func Find(dst string) bool {
// 	path, _ := util.OS("./"+workspace+"/DedeCMS").Find("*", "R")
// 	for _, p := range path {
// 		o := util.OS(p)
// 		if o.IsDir() == false &&
// 			o.Suffix() == ".php" || o.Suffix() == ".js" || o.Suffix() == ".htm" || o.Suffix() == ".css" || o.Suffix() == ".md" {
// 			if util.Text(o.Cat()).Find(dst) {
// 				return true
// 			}
// 		}
// 	}
// 	return false
// }

// Config 配置
type Config struct {
	Replaces map[string]string
	Parse    map[string]interface{}
	Charset  []string
	DelFile  []string
	AddFile  map[string]string
	Regexp   map[string]string
	Suffix   []string
}

func loadConfig() Config {
	filename, _ := filepath.Abs("./config.yml")
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		panic(err)
	}
	return config
}

func main() {
	conf := loadConfig()
	filename := "index.html"

	if src, ok := util.OS(filename).Read(); ok {

		for k, v := range conf.Regexp {
			r := regexp.MustCompile(k)
			src = r.ReplaceAllString(src, v)
		}

		for _, v := range conf.Parse {
			src = util.Text(src).Parse(v)
		}

		for k, v := range conf.Replaces {
			src = util.Text(src).Replace(k, v)
		}

		for _, v := range conf.Charset {
			util.OS("./out").Add(v).Add(filename).WriteCharset(src, v)
		}
	}

	// exampleWriteGBK("./index.html")
	// path, _ := util.OS("./"+workspace+"/DedeCMS").Find("*", "R")

	// for _, p := range path {
	// 	o := util.OS(p)
	// 	if o.IsDir() == false &&
	// 		o.Suffix() == ".gif" || o.Suffix() == ".png" || o.Suffix() == ".jpg" || o.Suffix() == ".jpge" || o.Suffix() == ".js" || o.Suffix() == ".css" {
	// 		if Find(o.File()) == false {
	// 			fmt.Println(o.Path(), "Delete.")
	// 			o.Rm()
	// 		}
	// 	}
	// }
	// util.Text("sss").Find()

	// 	packageList, err := util.OS("./package").Ls("*.tar.gz")

	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	for _, v := range packageList {
	// 		fmt.Println("./"+v, util.OS("./"+v).FileName(".tar.gz"))
	// 		if len(util.Text(util.OS("./"+v).FileName(".tar.gz")).Split("BIG5")) > 1 {
	// 			archiver.Unarchive("./"+v, "./task/"+util.OS("./"+v).FileName(".tar.gz"))
	// 		} else {
	// 			archiver.Unarchive("./"+v, "./task/")
	// 		}
	// 	}

	// 	taskList, err := util.OS("./task").Ls()

	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	for _, v := range taskList {
	// 		util.OS("./" + v + "/dede/images/login-bg.jpg").Rm()
	// 		copy("./replace/login-bg.jpg", "./"+v+"/uploads/dede/images/login-bg.jpg")
	// 		fmt.Println("./out/" + util.OS(v).FileName("") + ".tar.gz")
	// 		err := archiver.Archive([]string{v}, "./out/"+util.OS(v).FileName("")+".tar.gz")
	// 		fmt.Println(err)
	// 	}

	// }

	// func copy(src, dst string) (int64, error) {
	// 	sourceFileStat, err := os.Stat(src)
	// 	if err != nil {
	// 		return 0, err
	// 	}

	// 	if !sourceFileStat.Mode().IsRegular() {
	// 		return 0, fmt.Errorf("%s is not a regular file", src)
	// 	}

	// 	source, err := os.Open(src)
	// 	if err != nil {
	// 		return 0, err
	// 	}
	// 	defer source.Close()

	// 	destination, err := os.Create(dst)
	// 	if err != nil {
	// 		return 0, err
	// 	}

	// 	defer destination.Close()
	// 	nBytes, err := io.Copy(destination, source)

	// 	return nBytes, err
}

// DB Database 初始化 ...
func DB() *gorm.DB {
	db, _ := _db()
	return db
}

// DB .
func _db() (db *gorm.DB, err error) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: 5000 * time.Second, // Slow SQL threshold
		},
	)
	dsn := "root:198342calf@tcp(127.0.0.1:3306)/img?charset=utf8mb4&parseTime=True&loc=Local"
	return gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,   // data source name
		DefaultStringSize:         256,   // default size for string fields
		DisableDatetimePrecision:  true,  // disable datetime precision, which not supported before MySQL 5.6
		DontSupportRenameIndex:    true,  // drop & create when rename index, rename index not supported before MySQL 5.7, MariaDB
		DontSupportRenameColumn:   true,  // `change` when rename column, rename column not supported before MySQL 8, MariaDB
		SkipInitializeWithVersion: false, // auto configure based on currently MySQL version
	}), &gorm.Config{
		Logger: newLogger,
	})

}
func Read(path string) {
	dat, err := ioutil.ReadFile(path)
	fmt.Println(string(dat), err)
}

func Query(src string) *goquery.Document {
	f, err := os.Open(src)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		log.Fatal(err)
	}
	return doc

}
