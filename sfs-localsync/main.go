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
		opt := &api.ChangeMessage{}
		var bytes []byte

		for {
			var line string
			line, err := reader.ReadString('\n')
			if err != nil {
				handleError(err)
				continue
			}

			// TODO can this be removed?
			if line == "" {
				time.Sleep(time.Second)
				continue
			}

			err = json.Unmarshal([]byte(line), opt)
			if err != nil {
				handleError(err)
				return
			}

			if opt.LastChunk {
				opt.Contents = string(append(bytes, []byte(opt.Contents)...))
				break
			} else {
				bytes = append(bytes, []byte(opt.Contents)...)
			}
		}

		dests := strings.Split(opt.Arguments["destinations"], ",")

		for _, v := range dests {
			var file *os.File
			path := filepath.Clean(v + "/" + opt.Path)

			if opt.Type == api.CreateUpdate {
				err := os.MkdirAll(path, 0777)
				if err != nil {
					handleError(err)
					break
				}

				file, err = os.Create(filepath.Clean(path + "/" + opt.Name))
				if err != nil {
					handleError(err)
					break
				}

				_, err = file.Write([]byte(opt.Contents))
				if err != nil {
					handleError(err)
				}
				file.Close()
			} else if opt.Type == api.Remove {
				err := os.Remove(path + "/" + opt.Name)
				if err != nil {
					handleError(err)
					break
				}
			} else {
				fmt.Fprintf(os.Stderr, "invalid message no recognized type %v\n", opt.Type)
			}

		}

	}
}

func handleError(err error) {
	fmt.Fprintln(os.Stderr, err)
}
