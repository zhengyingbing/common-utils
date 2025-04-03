package utils

import (
	"bufio"
	"io"
	"os"
)

type LineCallback func(err error, line int, content string) bool

func ReadLine(filePath string, callback LineCallback) {
	file, err := os.Open(filePath)
	if err != nil {
		callback(err, 0, "")
		return
	}
	defer file.Close()
	rd := bufio.NewReader(file)
	line := 1
	for {
		b, _, err := rd.ReadLine()
		if err == io.EOF {
			callback(io.EOF, line, "")
			break
		}
		if callback(nil, line, string(b)) {
			break
		}
		line++
	}
}
