package prompt

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/term"
)

type (
	// HandlerFunc is function to handle user input. Any returned
	// errors are printed by the prompt loop and the loop continues.
	// Use p.Stop to exit the loop.
	HandlerFunc func(input []string, p *Prompt) error

	// OnStateChangeFunc allows for updating of Prompt state at the
	// beginning of every prompt loop.
	OnStateChangeFunc func(p *Prompt)

	// Prompt is a struct containing relevant metadata to simulate a
	// cli prompt. The default HistSize is 1000. Modify to your needs.
	// Set to -1 to disable history. The default prompt is "> ". Use
	// (*Prompt).SetPrompt() to modify it to fit your needs. The
	// default Splitter is ShellSplit.
	Prompt struct {
		Comment       string
		Handler       HandlerFunc
		History       []string
		HistSize      int
		OnStateChange OnStateChangeFunc
		Splitter      SplitterFunc
		Stop          bool

		input  []byte
		offset int
		prefix string
	}

	// SplitterFunc is used to split user input.
	SplitterFunc func(input string) ([][]string, error)
)

// New will return a sane-default Prompt instance.
func New() *Prompt {
	return (&Prompt{}).init()
}

// Cmd will execute single commands, although you can join commands
// with semi-colons.
func (p *Prompt) Cmd(cmd string, interactive ...bool) error {
	p.init()
	p.Stop = false
	p.OnStateChange(p)

	if p.processInput(cmd); p.Stop {
		return nil
	}

	if (len(interactive) > 0) && interactive[0] {
		return p.Run()
	}

	return nil
}

func (p *Prompt) init() *Prompt {
	if p.HistSize == 0 {
		p.HistSize = 1000
	}

	if p.OnStateChange == nil {
		p.OnStateChange = DoNothing
	}

	if p.prefix == "" {
		p.prefix = "> "
	}

	if p.Splitter == nil {
		p.Splitter = ShellSplit
	}

	return p
}

func (p *Prompt) processInput(input string) {
	var cmds [][]string
	var e error

	if input = strings.TrimSpace(input); input == "" {
		return
	}

	if (p.Comment != "") && strings.HasPrefix(input, p.Comment) {
		return
	}

	if cmds, e = p.Splitter(input); e != nil {
		fmt.Println(e.Error())
		return
	}

	if input != "" {
		if len(p.History) > 0 {
			if p.History[len(p.History)-1] != input {
				p.History = append(p.History, input)
			}
		} else {
			p.History = append(p.History, input)
		}

		if p.HistSize > 0 {
			if len(p.History) > p.HistSize {
				p.History = p.History[1:]
			}
		}
	}

	for _, cmd := range cmds {
		if len(cmd) == 0 {
			continue
		}

		if p.Handler != nil {
			if e = p.Handler(cmd, p); e != nil {
				fmt.Println(e.Error())
			}
		}

		if p.Stop {
			break
		}

		p.OnStateChange(p)

		// Ensure OnStateChangeFunc didn't screw anything up
		p.init()
	}
}

//nolint:cyclop,gocyclo // no shit
func (p *Prompt) readString() (input string, e error) {
	var b byte
	var diff int
	var direction int
	var hist int
	var old *term.State
	var stdin int = int(os.Stdin.Fd())
	var tmp []byte

	if old, e = term.MakeRaw(stdin); e != nil {
		e = fmt.Errorf("failed to make stdin raw: %w", e)
		return "", e
	}
	defer func() {
		var err error = term.Restore(stdin, old)

		if (e == nil) && (err != nil) {
			e = err
			input = ""
		}
	}()

	p.input = nil
	p.offset = 0

	for {
		// Reset state
		diff = 0
		direction = 0

		if b, e = readByte(); e != nil {
			return "", e
		}

		// fmt.Printf("\r\nread 0x%x\r\n\r\n", b)

		switch b {
		case 0x3, 0x4: //nolint:mnd // ^C and ^D
			p.Stop = true
			return "", nil
		case 0x9: //nolint:mnd // tab
			// Do nothing for now, no tab-completion support yet
		case 0xa, 0xd: //nolint:mnd // newline (raw)
			p.offset = 0
			fmt.Printf("%s\n\r", p.render(0, 0))

			input = strings.TrimSpace(string(p.input))

			return input, nil
		case 0xc: //nolint:mnd // ^L
			fmt.Print("\x1b[1;1H\x1b[2J")
		case 0x1b: //nolint:mnd // escape code
			// Read [
			if _, e = readByte(); e != nil {
				return "", e
			}

			// Read arrow key
			if b, e = readByte(); e != nil {
				return "", e
			}

			switch b {
			case 0x41: //nolint:mnd // Up
				if hist == 0 {
					tmp = p.input
				}

				if hist < len(p.History) {
					hist += 1

					diff = len(p.History[len(p.History)-hist])
					diff -= len(p.input)
					p.input = []byte(p.History[len(p.History)-hist])
					p.offset = 0
				}
			case 0x42: //nolint:mnd // Down
				if hist > 0 {
					hist -= 1

					if hist == 0 {
						diff = len(tmp)
						diff -= len(p.input)
						p.input = tmp
					} else {
						diff = len(p.History[len(p.History)-hist])
						diff -= len(p.input)
						p.input = []byte(
							p.History[len(p.History)-hist],
						)
					}

					p.offset = 0
				}
			case 0x43: //nolint:mnd // Right
				if p.offset > 0 {
					direction = 1
					p.offset -= 1
				}
			case 0x44: //nolint:mnd // Left
				if p.offset < len(p.input) {
					direction = -1
					p.offset += 1
				}
			}
		case 0x7f: //nolint:mnd // backspace
			if len(p.input) > 0 {
				diff = -1

				if p.offset == 0 {
					p.input = p.input[:len(p.input)-1]
				} else if p.offset < len(p.input) {
					p.input = slices.Delete(
						p.input,
						len(p.input)-p.offset-1,
						len(p.input)-p.offset,
					)
				}

				if p.offset > len(p.input) {
					p.offset = len(p.input)
				}
			}
		default:
			diff = 1

			if p.offset == 0 {
				p.input = append(p.input, b)
			} else {
				p.input = slices.Insert(
					p.input,
					len(p.input)-p.offset,
					b,
				)
			}
		}

		fmt.Print(p.render(diff, direction))
	}
}

func (p *Prompt) render(diff int, direction int) string {
	type cursor struct {
		col  int
		len  int
		pos  int
		row  int
		rows int
	}

	var e error
	var lines []string
	var newCursor cursor
	var oldCursor cursor
	var prefix string
	var sb strings.Builder
	var stdin int = int(os.Stdin.Fd())
	var termWidth int

	termWidth, _, e = term.GetSize(stdin)
	if (e != nil) || (termWidth <= 0) {
		termWidth = 80 // sane default
	}

	prefix = plain(p.prefix)
	lines = strings.Split(prefix, "\n\r")

	// Account for newlines in prefix
	if len(lines) > 0 {
		for range lines[0 : len(lines)-1] {
			newCursor.len += termWidth
		}

		newCursor.len += len(lines[len(lines)-1])
	}

	newCursor.len += len(p.input)
	newCursor.pos = newCursor.len - p.offset
	newCursor.col = newCursor.pos % termWidth
	newCursor.row = newCursor.pos / termWidth
	newCursor.rows = newCursor.len / termWidth

	if p.offset > 0 {
		newCursor.rows = (newCursor.len - 1) / termWidth
	}

	oldCursor.len = newCursor.len - diff

	if diff == 0 {
		oldCursor.len = newCursor.len - direction
	}

	oldCursor.pos = oldCursor.len - p.offset
	oldCursor.col = oldCursor.pos % termWidth
	oldCursor.row = oldCursor.pos / termWidth
	oldCursor.rows = oldCursor.len / termWidth

	if p.offset > 0 {
		oldCursor.rows = (oldCursor.len - 1) / termWidth
	}

	// Reset prompt based off old input
	if oldCursor.row > 0 {
		fmt.Fprintf(&sb, "\x1b[%dA", oldCursor.row)
	}

	// Display new prompt value
	fmt.Fprintf(&sb, "\r%s%s\x1b[J\r", p.prefix, p.input)

	// Move cursor to correct location
	if (newCursor.col == 0) && (diff != 0) {
		fmt.Fprint(&sb, "\n")
	}

	if newCursor.row < newCursor.rows {
		fmt.Fprintf(&sb, "\x1b[%dA", newCursor.rows-newCursor.row)
	}

	if newCursor.col > 0 {
		fmt.Fprintf(&sb, "\x1b[%dC", newCursor.col)
	}

	return sb.String()
}

// Run will start the interactive prompt.
func (p *Prompt) Run() error {
	var e error
	var input string

	p.init()
	p.Stop = false
	p.OnStateChange(p)

	for !p.Stop {
		fmt.Print(p.prefix)

		if input, e = p.readString(); e != nil {
			return e
		} else if p.Stop {
			break
		}

		p.processInput(input)
	}

	return nil
}

// Script will execute multiple commands, as if read from a file.
func (p *Prompt) Script(lines []string, interactive ...bool) error {
	p.init()
	p.Stop = false
	p.OnStateChange(p)

	for _, line := range lines {
		if line = strings.TrimSpace(line); line == "" {
			continue
		}

		if (p.Comment != "") && strings.HasPrefix(line, p.Comment) {
			continue
		}

		fmt.Println(p.prefix + line)

		if p.processInput(line); p.Stop {
			return nil
		}
	}

	if (len(interactive) > 0) && interactive[0] {
		return p.Run()
	}

	return nil
}

// ScriptFile will read commands from a file.
func (p *Prompt) ScriptFile(path string, interactive ...bool) error {
	var b []byte
	var e error

	if b, e = os.ReadFile(filepath.Clean(path)); e != nil {
		return fmt.Errorf("failed to read script: %w", e)
	}

	b = bytes.TrimSpace(b)

	return p.Script(strings.Split(string(b), "\n"), interactive...)
}

// SetPrompt will set the prompt's prefix.
func (p *Prompt) SetPrompt(prefix string) {
	prefix = strings.ReplaceAll(prefix, "\r\n", "\n\r")
	prefix = strings.ReplaceAll(prefix, "\n", "\n\r")
	p.prefix = strings.ReplaceAll(prefix, "\n\r\r", "\n\r")
}
