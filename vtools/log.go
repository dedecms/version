package vtools

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var (
	// Trace 记录所有日志
	Trace *log.Logger
	// Info 重要的信息
	Info *log.Logger
	// Warning 需要注意的信息
	Warning *log.Logger
	// Error 非常严重的问题
	Error *log.Logger
	// Conf 系统配置
	Conf Config
)

// Config 配置
type Config struct {
	Replaces  map[string]string
	Parse     map[string]interface{}
	Charset   []string
	DelFile   []string
	AddFile   map[string]string
	Regexp    map[string]string
	Suffix    []string
	SourceDIR string
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

func init() {

	Conf = loadConfig()

	file, err := os.OpenFile("./log/run.log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open error log file:", err)
	}

	Trace = log.New(ioutil.Discard,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(os.Stdout,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(os.Stdout,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(io.MultiWriter(file, os.Stderr),
		"ERROR: ",
		log.Ldate|log.Ltime)
}
