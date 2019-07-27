package api

import (
	"github.com/fsnotify/fsnotify"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type SyncOptions struct {
	Integration string
	Targets     []string
	Arguments   map[string]string
}

type ChangeMessage struct {
	Source    string
	Path      string
	Arguments map[string]string
	Contents  string
}

type SyncHandle struct {
	Cmd *exec.Cmd
}

func (s SyncHandle) Close() error {
	return s.Cmd.Process.Kill()
}

type MyWriter struct{}

func (MyWriter) Write(p []byte) (n int, err error) {
	fmt.Print("child: " + string(p))
	return len(p), nil
}

func Sync(opt SyncOptions) (*SyncHandle, error) {
	cmd := exec.Command(opt.Integration)

	cmd.Stderr = MyWriter{}
	cmd.Stdout = MyWriter{}

	writeCloser, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	// TODO find a better way to wait
	time.Sleep(time.Second)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				file, err := os.Open(event.Name)
				if err != nil {
					handleError(err)
					break
				}
				defer file.Close()

				bytes, err := ioutil.ReadAll(file)
				if err != nil {
					handleError(err)
					break
				}

				chgMsg := ChangeMessage{}
				chgMsg.Path = event.Name
				chgMsg.Arguments = opt.Arguments
				chgMsg.Contents = string(bytes)
				if len(opt.Targets) != 1 {
					chgMsg.Source, _ = filepath.Split(event.Name)
				}

				msg, err := json.Marshal(chgMsg)
				if err != nil {
					handleError(err)
					break
				}

				fmt.Println(string(msg))

				_, err = writeCloser.Write(msg)
				if err != nil {
					handleError(err)
				}
			}
		}
	}()

	for _, v := range opt.Targets {
		var folders []string
		folders = append(folders, v)
		for {
			if len(folders) == 0 {
				break
			}

			var dir string
			dir, folders = folders[len(folders)-1], folders[:len(folders)-1]

			err = watcher.Add(dir)
			if err != nil {
				return nil, err
			}

			files, err := ioutil.ReadDir(dir)
			if err != nil {
				return nil, err
			}

			for _, f := range files {
				if f.IsDir() {
					folders = append(folders, f.Name())
				}
			}
		}
	}

	return &SyncHandle{Cmd: cmd}, nil
}

func handleError(err error) {
	fmt.Fprintln(os.Stderr, "got error", err)
}
