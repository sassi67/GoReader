package lib

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

/*FileReader ... */
type FileReader struct {
	path string
	out  chan<- uint64
}

/*NewFileReader ... */
func NewFileReader(filePath string, outChannel chan<- uint64) *FileReader {
	return &FileReader{
		path: filePath,
		out:  outChannel,
	}
}

/*StringSearch ... */
func (fr *FileReader) StringSearch(strSearch string) {
	var counter uint64 = 0
	f, errFile := os.Open(fr.path)
	if errFile != nil {
		fmt.Println(errFile)
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), strSearch) {
			counter++
		}
	}

	if errScan := scanner.Err(); errScan != nil {
		fmt.Println(errScan)
		counter = 0
	}

	fr.out <- counter
}
