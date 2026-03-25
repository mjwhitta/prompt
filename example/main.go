package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mjwhitta/prompt"
)

var home string

func cmdHandler(input []string, p *prompt.Prompt) error {
	var args []string
	var cmd string
	var e error
	var files []os.DirEntry

	if len(input) == 0 {
		return nil
	}

	cmd = input[0]
	args = input[1:]

	switch cmd {
	case "": // Do nothing
	case "cd":
		if len(args) == 0 {
			args = []string{home}
		}

		e = os.Chdir(args[0])
	case "clear":
		fmt.Print("\x1b[1;1H\x1b[2J")
	case "clearhist":
		p.History = nil
	case "exit", "q", "quit":
		p.Stop = true
		return nil
	case "ls":
		if len(args) == 0 {
			args = []string{"."}
		}

		files, e = os.ReadDir(args[0])
		for _, file := range files {
			fmt.Println(file.Name())
		}
	default:
		e = fmt.Errorf("unknown command: %s", cmd)
	}

	return e
}

func init() {
	home, _ = os.UserHomeDir()
}

func main() {
	var p *prompt.Prompt = prompt.New()

	p.Handler = cmdHandler
	p.OnStateChange = updatePrompt
	// p.Splitter = prompt.SimpleSplit

	if e := p.Run(); e != nil {
		fmt.Println(e.Error())
	}
}

func updatePrompt(p *prompt.Prompt) {
	var prefix string

	prefix, _ = os.Getwd()
	if (home != "") && strings.HasPrefix(prefix, home) {
		prefix = "~" + strings.TrimPrefix(prefix, home)
	}

	prefix = fmt.Sprintf(
		"\x1b[32m%s\x1b[0m\n \x1b[31m$\x1b[0m ",
		prefix,
	)

	p.SetPrompt(prefix)
}
