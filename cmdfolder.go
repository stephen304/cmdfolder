package cmdfolder

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"strings"

	"github.com/carmark/pseudo-terminal-go/terminal"
	"github.com/mgutz/ansi"
)

/*
Folder is an interface... I don't know what I'm doing.
*/
type Folder interface {
	AddCommand(string, func(string))
	AddFolder(string, Folder)
	Run()
	RunWithTerm(string, *terminal.Terminal)

	// Folder default commands
	Ls(string)
}

/*
DefaultFolder is a struct that holds data and methods for a command folder
*/
type DefaultFolder struct {
	commands   map[string]func(string)
	subfolders map[string]Folder
}

/*
New creates a new command folder
*/
func New() Folder {
	folder := &DefaultFolder{commands: make(map[string]func(string)), subfolders: make(map[string]Folder)}
	folder.AddCommand("ls", folder.Ls)
	return folder
}

/*
Run starts the command environment
*/
func (folder *DefaultFolder) Run() {
	term, err := terminal.NewWithStdInOut()
	if err != nil {
		panic(err)
	}
	defer term.ReleaseFromStdInOut()

	// Make colors
	bluen := ansi.ColorFunc("blue+b")
	magentaen := ansi.ColorFunc("magenta")
	greenen := ansi.ColorFunc("green")
	bolden := ansi.ColorFunc("white+b")

	// Make prompt
	thisUser, _ := user.Current()
	username := thisUser.Username
	thisHost, _ := os.Hostname()
	prompt := bluen(username) + "@" + thisHost + " " + bolden("~%s") + " " + magentaen("[") + greenen("darkcli") + magentaen("]") + " %%"

	// Run it
	folder.RunWithTerm(prompt, term)
}

/*
RunWithTerm is used between folder instances to reuse the terminal object and prompt string
*/
func (folder *DefaultFolder) RunWithTerm(prompt string, term *terminal.Terminal) {
	term.SetPrompt(fmt.Sprintf(prompt, "") + " ")
	line, err := term.ReadLine()
	for {
		if err == io.EOF {
			term.Write([]byte(line))
			fmt.Println()
			return
		}
		if (err != nil && strings.Contains(err.Error(), "control-c break")) || len(line) == 0 {
			line, err = term.ReadLine()
		} else {
			if line == ".." {
				break
			} else if strings.HasPrefix(line, "cd ") {
				if folder.subfolders[line[3:]] != nil {
					folder.subfolders[line[3:]].RunWithTerm(fmt.Sprintf(prompt, "/"+line[3:]+"%s")+"%", term)
					term.SetPrompt(fmt.Sprintf(prompt, "") + " ")
				} else {
					fmt.Println("Folder not found")
				}
			} else if folder.commands[line] != nil {
				folder.commands[line](line)
			} else {
				term.Write([]byte(line + "\r\n"))
			}
			line, err = term.ReadLine()
		}
	}
}

/*
AddCommand adds a function as a command to the folder
*/
func (folder *DefaultFolder) AddCommand(command string, function func(string)) {
	folder.commands[command] = function
}

/*
AddFolder adds another folder instance as a child of the current folder
*/
func (folder *DefaultFolder) AddFolder(name string, subfolder Folder) {
	folder.subfolders[name] = subfolder
}

/*
Ls is the default ls command function which lists the subfolders. It may be overridden.
*/
func (folder *DefaultFolder) Ls(_ string) {
	for name := range folder.subfolders {
		fmt.Println(name)
	}
}
