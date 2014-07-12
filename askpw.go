package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	PWSAFE_BIN  = "pwsafe"
	ASKPW_ENV   = "ASKPW_ENTRY"
	PROMPT_TEXT = "Which entry to select (blank to skip): "
	USAGE       = `askpw [OPTION]...
Prompt for the entry key to read from a password manager.

    --bin=PATH              the absolute path to the invoked binary
    --entry=NAME            this will not ask for the entry via prompt
    --stderr                ask for the entry key on stderr
    --version               display the askpw version and exit
    --help                  display the usage/help message and exit
Alternatively the entry can also be set via the environment (` + ASKPW_ENV + `)
The default command is ` + PWSAFE_BIN
)

const (
	ACTION_COMMAND int = iota
	ACTION_VERSION     = iota
	ACTION_HELP        = iota
)

var (
	ARG_BIN     = flag{"bin", "b", true}
	ARG_ENTRY   = flag{"entry", "e", true}
	ARG_ERROR   = flag{"stderr", "2", false}
	ARG_VERSION = flag{"version", "v", false}
	ARG_HELPALT = flag{"help", "?", false}
	ARG_HELP    = flag{"help", "h", false}
)

var (
	VERSION string = "1.1.2"

	DEBUG bool = false
)

type flag struct {
	long   string
	short  string
	valued bool
}

type arguments struct {
	action int
	err    bool
	bin    string
	entry  string
	pass   []string
}

func (a *arguments) append(value string) {
	a.pass = append(a.pass, value)
}

func (f *flag) matches(value string) bool {
	var l, s string = "--" + f.long, "-" + f.short

	if f.valued {
		return (strings.HasPrefix(value, l) || strings.HasPrefix(value, s))
	} else {
		return value == l || value == s
	}
}

func (f *flag) value(value string) (string, bool) {
	parts := strings.SplitAfterN(value, "=", 2)

	if len(parts) > 1 && len(parts[1]) > 0 {
		return parts[1], true
	}

	return "", false
}

func main() {
	var args arguments
	var path string
	var entry string
	var err error

	args.bin = PWSAFE_BIN

	if 0 == len(os.Args) {
		debug("no arguments provided")
	} else if err = parse(os.Args[1:], &args); nil != err {
		warn("Unable to parse command line:", err)
		os.Exit(9)
	}

	if ACTION_VERSION == args.action {
		version()
		os.Exit(0)
	} else if ACTION_HELP == args.action {
		help()
		os.Exit(0)
	}

	if 0 == len(args.entry) {
		args.entry = os.Getenv(ASKPW_ENV)
	}

	if entry, err = prompt(args.entry, args.err); nil != err {
		warn("Invalid password entry:", err)
		os.Exit(3)
	} else if 0 == len(entry) {
		debug("empty entry name. aborting.")
		os.Exit(0)
	} else {
		args.entry = entry
	}

	if path, err = resolve(args.bin); nil != err {
		warn("Invalid manager command:", err)
		os.Exit(4)
	} else {
		args.bin = path
	}

	if err = run(args); nil != err {
		warn("Password manager error:", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func parse(cmd []string, args *arguments) error {
	var visitor func(string)

	noop := func(arg string) {}
	pass := func(arg string) {
		debug("pass-through argument:", arg)
		args.append(arg)
	}
	proc := func(arg string) {
		param := strings.SplitAfterN(arg, "=", 2)

		switch {
		case "--" == arg:
			visitor = pass // use different visitor
		case ARG_VERSION.matches(arg):
			debug("displaying command version")
			args.action = ACTION_VERSION
			visitor = noop // ignore all remaining args
		case ARG_HELPALT.matches(arg):
			fallthrough
		case ARG_HELP.matches(arg):
			debug("displaying usage message")
			args.action = ACTION_HELP
			visitor = noop // ignore all remaining args
		case ARG_ERROR.matches(arg):
			debug("prompting on stderr")
			args.err = true
		case ARG_BIN.matches(arg):
			debug("binary argument:", arg)
			replace(&args.bin, param, 1)
		case ARG_ENTRY.matches(arg):
			debug("entry argument:", arg)
			replace(&args.entry, param, 1)
		default:
			pass(arg)
		}
	}

	args.action = ACTION_COMMAND
	visitor = proc

	for _, arg := range cmd {
		visitor(arg)
	}

	return nil
}

func prompt(current string, stderr bool) (entry string, err error) {
	if 0 == len(current) {
		if stderr {
			fmt.Fprint(os.Stderr, PROMPT_TEXT)
		} else {
			fmt.Fprint(os.Stderr, PROMPT_TEXT)
		}

		entry, err = readln()
	} else {
		debug("entry already defined as", current)
		entry, err = current, nil
	}

	return
}

func readln() (string, error) {
	var buf *bufio.Reader = bufio.NewReader(os.Stdin)

	if data, err := buf.ReadBytes('\n'); nil == err {
		/*
		   data = bytes.TrimRight(data, "\n")

		   if 0 < len(data) && data[len(data)-1] == 13 { //'\r'
		       data = bytes.TrimRight(data, "\r")
		   }
		*/
		data = bytes.TrimSpace(data)

		return string(data[:]), nil
	} else {
		return "", err
	}
}

func replace(value *string, argv []string, index int) {
	if len(argv) > index && len(argv[index]) > 0 {
		*value = argv[index]
	}
}

func resolve(value string) (string, error) {
	if path, err := exec.LookPath(value); nil != err {
		// the LookPath error is somewhat bloated
		return "", errors.New("Unable to resolve " + value)
	} else {
		return path, nil
	}
}

func run(data arguments) error {
	var arg []string = append(data.pass, data.entry)
	var cmd *exec.Cmd = exec.Command(data.bin, arg...)

	debug("command:", cmd.Args)

	// connecte everything with the sub-process
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return errors.New(data.bin + " did not exit successfully")
	}

	return nil
}

func version() {
	fmt.Println("askpw", VERSION)
}

func help() {
	fmt.Println(USAGE)
}

func warn(message string, err error) {
	fmt.Println(message, err)
}

func debug(data ...interface{}) {
	if DEBUG {
		fmt.Println(data)
	}
}
