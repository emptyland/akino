package gunit

import (
	"bytes"
	"io/ioutil"
	"strings"
)

// lines cache
// fileName -> []line
type linesCache map[string][]string

func newLinesCache() linesCache {
	rv := make(map[string][]string)
	return linesCache(rv)
}

func (self linesCache) Put(fileName string) ([]string, error) {
	lines, found := self[fileName]
	if found {
		return lines, nil
	}

	return readLines(fileName)
}

const lineDelimString = "\n"

var lineDelim = byte(lineDelimString[0])

func readLines(fileName string) ([]string, error) {
	var originData []byte
	var err error

	if originData, err = ioutil.ReadFile(fileName); err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(originData)
	lines := make([]string, 0, 128)
	for err == nil {
		var line string

		line, err = buf.ReadString(lineDelim)
		line = strings.TrimSuffix(line, lineDelimString)
		lines = append(lines, line)
	}
	return lines, err
}
