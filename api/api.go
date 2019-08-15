package api

import (
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"strings"

	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	CreateUpdate            = "create/update"
	Remove                  = "remove"
	DefaultMaxFileSizeBytes = 10 * 1000
	PlaceholderExtension    = ".placeholder"
)

var placeheld = make(map[string]bool)

type SyncOptions struct {
	Integration        string
	Targets            []string
	Arguments          map[string]string
	RemoveDeletedFiles bool
	MaxFileSizeBytes   int
	UsePlaceholders    bool
}

type ChangeMessage struct {
	Name      string
	Path      string
	Arguments map[string]string
	Contents  string
	Type      string
	LastChunk bool
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
	if opt.MaxFileSizeBytes == 0 {
		opt.MaxFileSizeBytes = DefaultMaxFileSizeBytes
	}

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
					chgMsgs := []*ChangeMessage{}

					var file *os.File

					fmt.Printf("event detected %v:%v\n", event.Op.String(), event.Name)

					if event.Op == fsnotify.Remove {
						// if the file has been swapped for a placeholder then ignore the delete and exit
						if placeheld[event.Name] {
							return nil
						}

						chgMsg := &ChangeMessage{}
						chgMsg.Type = Remove
						chgMsg.LastChunk = true
						chgMsgs = append(chgMsgs, chgMsg)
					} else {
						var err error
						if isPlaceholder(event.Name) {
							return nil
						}

						file, err = os.Open(event.Name)
						if err != nil {
							return errors.Wrapf(err, "error while opening file %v", event.Name)
						}
						defer file.Close()

						bytes, err := ioutil.ReadAll(file)
						if err != nil {
							return err
						}

						if len(bytes) > opt.MaxFileSizeBytes {
							for {
								var clip int
								if len(bytes) < opt.MaxFileSizeBytes {
									clip = len(bytes)
								} else {
									clip = opt.MaxFileSizeBytes
								}

								chgMsg := &ChangeMessage{}
								chgMsg.Type = CreateUpdate

								encoded := make([]byte, base64.StdEncoding.EncodedLen(len(bytes[:clip])))
								base64.StdEncoding.Encode(encoded, bytes[:clip])

								chgMsg.Contents = string(encoded)
								bytes = bytes[clip:]

								chgMsgs = append(chgMsgs, chgMsg)

								if len(bytes) == 0 {
									chgMsg.LastChunk = true
									break
								}
							}
						} else {
							chgMsg := &ChangeMessage{}
							chgMsg.Type = CreateUpdate

							encoded := make([]byte, base64.StdEncoding.EncodedLen(len(bytes)))
							base64.StdEncoding.Encode(encoded, bytes)

							chgMsg.Contents = string(encoded)
							chgMsg.LastChunk = true

							chgMsgs = append(chgMsgs, chgMsg)
						}
					}

					parentDir, base := filepath.Split(event.Name)

					for _, chgMsg := range chgMsgs {
						if isPlaceholder(base) {
							chgMsg.Name = base[:strings.LastIndex(base, PlaceholderExtension)]
						} else {
							chgMsg.Name = base
						}

						chgMsg.Arguments = opt.Arguments
						sub, err := subTargetsFromDir(opt.Targets, parentDir)
						if err != nil {
							return err
						}

						chgMsg.Path = sub

						msg, err := json.Marshal(chgMsg)
						if err != nil {
							return err
						}

						// fmt.Println(string(msg))

						_, err = writeCloser.Write([]byte(string(msg) + "\n"))
						if err != nil {
							return err
						}
					}

					fmt.Printf("file written in %v chunks\n", len(chgMsgs))

					if event.Op != fsnotify.Remove {
						if opt.UsePlaceholders && !isPlaceholder(event.Name) {
							fmt.Println("swapping for placeholder")

							placeheld[event.Name] = true

							err = os.Remove(file.Name())
							if err != nil {
								return err
							}

							_, err = os.Create(file.Name() + PlaceholderExtension)
							if err != nil {
								return err
							}
						}
					}

					return err
				}
			}()
			fmt.Println("done with event")
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

func isPlaceholder(name string) bool {
	fmt.Printf("name is %v\nlast index is %v\nlen(name) is %v\nlen(placeholder.ext) is %v\n", name, strings.LastIndex(name, PlaceholderExtension), len(name), len(PlaceholderExtension))

	isPlaceholder := strings.LastIndex(name, PlaceholderExtension)+len(PlaceholderExtension) == len(name)

	fmt.Printf("isPlaceholder is %v\n", isPlaceholder)

	return isPlaceholder
}
