package main

import (
	"github.com/njenan/simplefilesync/api"

	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	fmt.Println("starting sfs localsync")

	reader := bufio.NewReader(os.Stdin)

	for {
		var line string
		line, err := reader.ReadString('\n')
		if err != nil {
			handleError(err)
			continue
		}

		if line == "" {
			time.Sleep(time.Second)
			continue
		}

		opt := &api.ChangeMessage{}
		err = json.Unmarshal([]byte(line), opt)
		if err != nil {
			handleError(err)
			return
		}

		dests := strings.Split(opt.Arguments["destinations"], ",")

		for _, v := range dests {
			var file *os.File
			parent := filepath.Base(opt.Path)
			err = os.MkdirAll(v+"/"+parent, 0777)
			if err != nil {
				handleError(err)
				break
			}
			filePath := v + "/" + filepath.Base(opt.Path) + "/" + opt.Name
			file, err = os.Create(filepath.Clean(filePath))

			if err != nil {
				handleError(err)
				break
			}

			_, err = file.Write([]byte(opt.Contents))
			if err != nil {
				handleError(err)
			}
			file.Close()

		}

	}
}

func handleError(err error) {
	fmt.Fprintln(os.Stderr, err)
}
