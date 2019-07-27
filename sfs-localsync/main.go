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
		fmt.Println("starting read")

		var line string
		line, err := reader.ReadString('\n')
		if err != nil {
			handleError(err)
			continue
		}

		if line == "" {
			fmt.Println("sleeping")
			time.Sleep(time.Second)
			continue
		}

		fmt.Println("read line: " + line)

		opt := &api.ChangeMessage{}
		err = json.Unmarshal([]byte(line), opt)
		if err != nil {
			fmt.Println("1")
			handleError(err)
			return
		}

		filename := filepath.Base(opt.Path)
		dests := strings.Split(opt.Arguments["destinations"], ",")

		fmt.Println("got dests")
		fmt.Println(dests)

		for _, v := range dests {
			var file *os.File

			if opt.Source == "" {
				file, err = os.Create(v + "/" + filename)
			} else {
				parent := filepath.Base(opt.Source)
				fmt.Println("parent is " + parent)
				err = os.MkdirAll(v+"/"+parent, 0777)
				if err != nil {
					handleError(err)
					fmt.Println("2")
					break
				}
				filePath := v + "/" + filepath.Base(opt.Source) + "/" + filename
				fmt.Println("filepath is " + filePath)
				file, err = os.Create(filePath)
			}

			if err != nil {
				handleError(err)
				fmt.Println("3")
				break
			}
			defer file.Close()

			_, err = file.Write([]byte(opt.Contents))
			if err != nil {
				handleError(err)
				fmt.Println("4")
				break
			}

			fmt.Println("done")
		}

	}
}

func handleError(err error) {
	fmt.Fprintln(os.Stderr, err)
}
