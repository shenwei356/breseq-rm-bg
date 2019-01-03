// Copyright © 2019 Wei Shen <shenwei356@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
)

var help = `
breseq-rm-ctrl

Version: v0.1.0

Author: Wei Shen <shenwei356@gmail.com>

Source code: https://github.com/shenwei356/breseq-rm-ctrl

Usage:

  breseq-rm-ctrl ctrl/output/index.html sample1/output/index.html > sample1/output/index2.html

`

func main() {
	files := os.Args
	if len(files) != 3 {
		checkError(fmt.Errorf(help))
	}
	files = (files[1:])
	checkFiles(".html", files...)

	ctrl := readRecordsFromCtrl(files[0])
	fmt.Fprintf(os.Stderr, "loaded %d records from %s\n", len(ctrl), files[0])

	file := files[1]

	fh, err := os.Open(file)
	checkError(err)
	defer fh.Close()

	scanner := bufio.NewScanner(fh)
	scanner.Split(bufio.ScanLines)

	var line, line2, record string
	var buf, buf2 bytes.Buffer
	var flag bool
	var ok bool
	for scanner.Scan() {
		line = scanner.Text()
		if flag {
			if strings.HasPrefix(line, tagEnd) {
				record = strings.Trim(buf.String(), "\r\n ")

				buf2.Reset()
				for _, line2 = range strings.Split(record, "\n") {
					if strings.Index(line2, "href") < 0 {
						buf2.WriteString(line2)
					}
				}
				if _, ok = ctrl[buf2.String()]; !ok {
					fmt.Println(tagStart)
					fmt.Println(record)
					fmt.Println(tagEnd)
				}
				flag = false
			} else {
				buf.WriteString(line + "\n")
			}
		} else {
			if strings.HasPrefix(line, tagStart) {
				flag = true
				buf.Reset()
			} else {
				fmt.Println(line)
			}
		}

	}
	checkError(scanner.Err())
}

func readRecordsFromCtrl(file string) map[string]struct{} {
	fh, err := os.Open(file)
	checkError(err)
	defer fh.Close()

	scanner := bufio.NewScanner(fh)
	scanner.Split(scanBreseqOutIndexHTML)

	data := make(map[string]struct{}, 100)
	var line, record string
	var buf bytes.Buffer
	for scanner.Scan() {
		record = scanner.Text()
		// fmt.Printf("--%s==\n", record)
		buf.Reset()
		for _, line = range strings.Split(record, "\n") {
			if strings.Index(line, "href") < 0 {
				buf.WriteString(line)
			}
		}

		// fmt.Printf("--%s==\n", buf.String())
		data[buf.String()] = struct{}{}
	}
	checkError(scanner.Err())

	return data
}

var tagEnd = `<!-- End Table Row -->`
var tagStart = `<!-- Print The Table Row -->`

var scanBreseqOutIndexHTML = func(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.Index(data, []byte(tagEnd)); i > 0 {
		if j := bytes.Index(data[0:i], []byte(tagStart)); j >= 0 {
			return i + 1, bytes.Trim(data[j+len(tagStart):i], "\r\n "), nil
		}
		return i + 1, data[:i], nil
	}

	if atEOF {
		if j := bytes.Index(data, []byte(tagStart)); j >= 0 {
			return len(data), bytes.Trim(data[j+len(tagStart):], "\r\n "), nil
		}
		return len(data), nil, nil
	}

	return 0, nil, nil
}
