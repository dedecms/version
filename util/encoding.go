package util

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"unicode"

	"github.com/yuin/charsetutil"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/transform"
)

// Document Struct
// 内容结构 ...
type Encoding struct {
	text    string
	bytes   []byte
	Charset string
}

func HTMLEncoding(src []byte) (*Encoding, bool) {
	e := &Encoding{
		text:  string(src),
		bytes: src,
	}
	return e.converter()
}

func (e *Encoding) Text() string {
	return e.text
}
func (e *Encoding) Bytes() []byte {
	return e.bytes
}

// Converter Function
// 运行对当前进程进行编码转换成UTF-8 ...
func (e *Encoding) converter() (*Encoding, bool) {

	// 自动获取资源编码 ...
	charset, ok := e.getCharset()

	e.Charset = charset

	// 未获取到资源编码 ...
	if !ok {
		return e, false
	}

	// UTF-8无需转换 ...
	if charset == "UTF-8" {
		return e, true
	}

	if encode := getEncoding(charset); encode != nil {
		if reader, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader(e.bytes), encode.NewDecoder())); err == nil {
			e.bytes = reader
			e.text = string(reader)
			return e, true
		} else {
			fmt.Println(err)
		}
	}

	// 转码失败
	return e, false
}

func getEncoding(charset string) encoding.Encoding {
	if e, err := ianaindex.MIB.Encoding(charset); err == nil && e != nil {
		return e
	}
	return nil
}

/*---------------------------------------------------------------*/

// Charset Function
// 返回当前进程的字符集 ...
func (e *Encoding) getCharset() (string, bool) {

	// 自动获取编码 ...
	encoding, err := charsetutil.GuessBytes(e.bytes)

	// 如果自动获取成功或encoding不为空
	// 则输出编码格式 ...
	if err == nil {
		return strings.ToUpper(encoding.Charset()), true
	}

	if isGBK(e.bytes) {
		return "GBK", true
	}

	if encoding != nil && encoding.Charset() != "WINDOWS-1252" {
		return strings.ToUpper(encoding.Charset()), true
	}

	// 如果内容中出现汉字
	// 则输出GB18030 ...
	if isHan(e.text) {
		return "GBK", true
	}

	// 不符合上述条件
	// 则返回空 ...
	return "", false
}

/*---------------------------------------------------------------*/

// IsHan Function
// 判断是否存在中文 ...
func isHan(str string) bool {
	hanLen := len(regexp.MustCompile(`[\P{Han}]`).ReplaceAllString(str, ""))
	for _, r := range str {
		if unicode.Is(unicode.Scripts["Han"], r) || hanLen > 0 {
			return true
		}
	}
	return false
}

// 判断是否为GBK
func isGBK(data []byte) bool {
	length := len(data)
	var i int = 0
	for i < length {
		if data[i] <= 0xff {
			i++
			continue
		} else {
			//大于127的使用双字节编码
			if data[i] >= 0x81 &&
				data[i] <= 0xfe &&
				data[i+1] >= 0x40 &&
				data[i+1] <= 0xfe &&
				data[i+1] != 0xf7 {
				i += 2
				continue
			} else {
				return false
			}
		}
	}
	return true
}

func preNUm(data byte) int {
	str := fmt.Sprintf("%b", data)
	var i int = 0
	for i < len(str) {
		if str[i] != '1' {
			break
		}
		i++
	}
	return i
}

// 判断是否为UTF8
func isUTF8(data []byte) bool {
	for i := 0; i < len(data); {
		if data[i]&0x80 == 0x00 {
			// 0XXX_XXXX
			i++
			continue
		} else if num := preNUm(data[i]); num > 2 {
			// 110X_XXXX 10XX_XXXX
			// 1110_XXXX 10XX_XXXX 10XX_XXXX
			// 1111_0XXX 10XX_XXXX 10XX_XXXX 10XX_XXXX
			// 1111_10XX 10XX_XXXX 10XX_XXXX 10XX_XXXX 10XX_XXXX
			// 1111_110X 10XX_XXXX 10XX_XXXX 10XX_XXXX 10XX_XXXX 10XX_XXXX
			// preNUm() 返回首个字节的8个bits中首个0bit前面1bit的个数，该数量也是该字符所使用的字节数
			i++
			for j := 0; j < num-1; j++ {
				//判断后面的 num - 1 个字节是不是都是10开头
				if data[i]&0xc0 != 0x80 {
					return false
				}
				i++
			}
		} else {
			//其他情况说明不是utf-8
			return false
		}
	}
	return true
}

// 判断是否在数组内
func InArray(dst string, m []string) bool {
	for _, v := range m {
		if dst == v {
			return true
		}
	}
	return false
}

/*---------------------------------------------------------------*/

/* End of file encoding.go */
/* Location: ./encoding.go */
