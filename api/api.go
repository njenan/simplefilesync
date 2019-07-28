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

const (
	megabyte     int = 1000 * 1000
	CreateUpdate     = "create/update"
	Remove           = "remove"
)

type SyncOptions struct {
	Integration        string
	Targets            []string
	Arguments          map[string]string
	RemoveDeletedFiles bool
}

type ChangeMessage struct {
	Name      string
	Path      string
	Arguments map[string]string
	Contents  string
	Type      string
}

type SyncHandle struct {
	Cmd *exec.Cmd
}

func (s SyncHandle) Close() error {
	return s.Cmd.Process.Kill()
}

type MyWriter struct{}

func (MyWriter) Write(p []byte) (n int, err error) {
	fmt.Printf("child: %v", string(p))
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
			err = func() error {
				select {
				case event := <-watcher.Events:
					chgMsg := ChangeMessage{}

					if event.Op == fsnotify.Remove {
						chgMsg.Type = Remove
					} else {
						chgMsg.Type = CreateUpdate

						file, err := os.Open(event.Name)
						if err != nil {
							return err
						}
						defer file.Close()

						bytes, err := ioutil.ReadAll(file)
						if err != nil {
							return err
						}

						chgMsg.Contents = string(bytes)
					}

					chgMsg.Arguments = opt.Arguments
					var parentDir string
					parentDir, chgMsg.Name = filepath.Split(event.Name)

					chgMsg.Path, err = subTargetsFromDir(opt.Targets, parentDir)
					if err != nil {
						return err
					}

					msg, err := json.Marshal(chgMsg)
					if err != nil {
						return err
					}

					// fmt.Println(string(msg))

					_, err = writeCloser.Write([]byte(string(msg) + "\n"))
					return err
				}
			}()
			if err != nil {
				handleError(err)
			}
		}
	}()

	for _, v := range opt.Targets {
		var stack []string
		stack = append(stack, v)
		for {
			if len(stack) == 0 {
				break
			}

			var dir string
			dir, stack = stack[len(stack)-1], stack[:len(stack)-1]

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
					stack = append(stack, dir+"/"+f.Name())
				}
			}
		}
	}

	return &SyncHandle{Cmd: cmd}, nil
}

func subTargetsFromDir(targets []string, dir string) (string, error) {
	low := len(dir) - 1

	proto := dir

	for _, v := range targets {
		for i := len(proto) - 1; i > 0; i-- {
			if i >= len(v) {
				if i-1 < low {
					low = i - 1
				}
				continue
			}

			if proto[i] == '/' {
				continue
			}

			if proto[i] != v[i] {
				if i-1 < low {
					low = i - 1
				}
			}
		}
	}

	return filepath.Clean("/" + dir[low+1:] + "/"), nil
}

func handleError(err error) {
	fmt.Fprintln(os.Stderr, "got error", err)
}
