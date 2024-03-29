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
	"regexp"
	"strings"

	"github.com/shenwei356/util/bytesize"
	"github.com/shenwei356/util/cliutil"
	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.Flags().StringSliceP("bg-files", "b", []string{}, "background mutation file, i.e., index.html produced by breseq")
	RootCmd.Flags().BoolP("bg-union", "u", false, "using union set of background mutations")
	RootCmd.Flags().BoolP("bg-inter", "i", true, "using intersection set of background mutations")

	RootCmd.Flags().StringP("buffer-size", "B", "1G", `size of buffer, supported unit: K, M, G. You need increase the value when "bufio.Scanner: token too long" error reported`)

}

func main() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var bufferSize int

var RootCmd = &cobra.Command{
	Use:   "breseq-rm-bg",
	Short: "remove backgroud mutations from breseq result",
	Long: `breseq-rm-bg -- remove backgroud mutations from breseq result

Version: v0.2.2

Author: Wei Shen <shenwei356@gmail.com>

Source code: https://github.com/shenwei356/breseq-rm-bg

Example:

    breseq-rm-bg -i -b bg1/output/index.html -b bg2/output/index.html \
      sample1/output/index.html > sample1/output/index.filtered.html
	  
`,
	Run: func(cmd *cobra.Command, args []string) {
		files := cliutil.GetFileList(args, true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}

		bufferSizeS := cliutil.GetFlagString(cmd, "buffer-size")

		if bufferSizeS == "" {
			checkError(fmt.Errorf("value of buffer size. supported unit: K, M, G"))
		}

		var err error
		bufferSize, err = bytesize.ParseByteSize(bufferSizeS)
		if err != nil {
			checkError(fmt.Errorf("invalid value of buffer size. supported unit: K, M, G"))
		}

		bgUnion := cliutil.GetFlagBool(cmd, "bg-union")
		bgInter := cliutil.GetFlagBool(cmd, "bg-inter")
		if bgUnion {
			bgInter = false
		}
		if bgInter {
			bgUnion = false
		}

		bgFiles := cliutil.GetFlagStringSlice(cmd, "bg-files")
		if len(bgFiles) == 0 {
			checkError(fmt.Errorf("flag -b/--bg-files needed"))
		}
		checkFiles(".html", bgFiles...)

		file := files[0]
		checkFiles(".html", file)

		// parsing background mutations
		backgroud := readRecordsFromBreseqFiles(bgFiles, bgUnion)

		// parsing and filtering target file
		fh, err := xopen.Ropen(file)
		checkError(err)
		defer fh.Close()

		scanner := bufio.NewScanner(fh)
		scanner.Buffer(make([]byte, bufferSize), int(bufferSize))
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
						if !strings.Contains(line2, "href") {
							buf2.WriteString(line2)
						}
					}
					if _, ok = backgroud[rmExtraInfo(buf2.String())]; !ok {
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
	},
}

func readRecordsFromBreseqFiles(files []string, union bool) map[string]struct{} {
	N := len(files)
	if N == 1 {
		union = true
	}
	mutations := make(map[string]struct{}, 100)
	if union {
		for _, file := range files {
			muts := readRecordsFromBreseqFile(file)
			log.Infof("loaded %d records from %s\n", len(muts), file)
			for m := range muts {
				mutations[m] = struct{}{}
			}
		}
		log.Infof("%d unique records from these %d files were loaded", len(mutations), N)
		return mutations
	}

	mutations2 := make(map[string]int, 100)
	for _, file := range files {
		muts := readRecordsFromBreseqFile(file)
		log.Infof("loaded %d records from %s\n", len(muts), file)
		for m := range muts {
			mutations2[m]++
		}
	}
	for m, n := range mutations2 {
		if n == N {
			mutations[m] = struct{}{}
		}
	}
	log.Infof("%d common records from these %d files were loaded", len(mutations), N)
	return mutations
}

func readRecordsFromBreseqFile(file string) map[string]struct{} {
	fh, err := os.Open(file)
	checkError(err)
	defer fh.Close()

	scanner := bufio.NewScanner(fh)
	scanner.Buffer(make([]byte, bufferSize), int(bufferSize))
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
		data[rmExtraInfo(buf.String())] = struct{}{}
	}
	checkError(scanner.Err())

	return data
}

var reButton = regexp.MustCompile(`<div .+?>|</div>`)
var reButton2 = regexp.MustCompile(`<input .+?>`)

func rmExtraInfo(text string) string {
	if reButton.MatchString(text) {
		text = reButton.ReplaceAllString(text, "")
	}

	if reButton2.MatchString(text) {
		text = reButton2.ReplaceAllString(text, "")
	}
	return text
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
