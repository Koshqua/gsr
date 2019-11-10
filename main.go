package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/urfave/cli"
)

var app = cli.NewApp()
var cmd *exec.Cmd

func main() {
	app.Name = "gsr"
	app.Usage = "Listening to changes on go files and restarting main.go"
	app.Commands = []*cli.Command{
		{
			Name:    "run",
			Aliases: []string{"r"},
			Usage:   "runs tracking on file",
			Action: func(c *cli.Context) error {
				addWatcher(c.Args().First())
				return nil
			},
		},
	}
	app.Run(os.Args)

}
func addWatcher(file string) {
	format := CheckFormat(file)
	if format != "go" {
		fmt.Println("You can run only .go files")
		os.Exit(1)
	}

	//getting current directory
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(pwd)
	watcher, _ := fsnotify.NewWatcher()
	defer watcher.Close()
	//Walking every file and adding listener
	filepath.Walk(pwd, func(path string, fi os.FileInfo, err error) error {
		//checking if it's a .go file
		format := CheckFormat(fi.Name())
		//adding watcher to to .go files in every folder
		if format == "go" {
			watcher.Add(path)
		}
		return nil
	})

	done := make(chan bool)
	go Run(file)
	go ListenExit()
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op.String() == "WRITE" {
					fmt.Println(event)
					Stop()
					Run(file)
				}

			case err := <-watcher.Errors:
				fmt.Print("ERROR: ", err)
			}
		}
	}()
	<-done
}

func Run(file string) {
	cmd = exec.Command("go", "run", file)
	out, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)

	}
	errors, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	errScanner := bufio.NewScanner(errors)
	go func() {
		for errScanner.Scan() {
			fmt.Println(errScanner.Text())
		}
	}()
	err = cmd.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	outScanner := bufio.NewScanner(out)
	go func() {
		for outScanner.Scan() {
			fmt.Println(outScanner.Text())
		}
	}()
}
func Stop() {
	if cmd.Process.Pid != 0 {
		cmd.Process.Kill()
	}
}

func ListenExit() {
	scanner := bufio.NewScanner(os.Stdout)
	for scanner.Scan() {
		if scanner.Text() == "exit" {
			os.Exit(1)
		}
	}
}

func CheckFormat(file string) string {
	fname := strings.Split(file, ".")
	format := ""
	if len(fname) > 1 {
		format = fname[1]
	}
	return format
}
