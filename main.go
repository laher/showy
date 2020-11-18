package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"
)

type Options struct {
	// Example of verbosity with level
	Verbose []bool `short:"v" long:"verbose" description:"Verbose output"`

	// Example of optional value
	User string `short:"u" long:"user" description:"User name" optional:"yes" optional-value:"pancake"`

	// Example of map with multiple default values
	Users map[string]string `long:"users" description:"User e-mail map" default:"system:system@example.org" default:"admin:admin@example.org"`
}

type VimHelpCmd struct {
	Options     *Options
	RuntimePath string `short:"r" long:"vimruntime" description:"path of runtime" default:"/usr/share/nvim/runtime"`
	MaxLines    int    `short:"l" long:"max-lines" description:"most lines to display" default:"20"`
	Key         string `short:"k" long:"key" description:"Key of help item" required:"yes"`
}

func (c *VimHelpCmd) Execute(args []string) error {
	b, err := ioutil.ReadFile(filepath.Join(c.RuntimePath, "doc", "tags"))
	if err != nil {
		return err
	}
	lines := strings.Split(string(b), "\n")
	type match struct {
		key    string
		file   string
		lookup string
	}
	matches := [][]match{{}, {}, {}}
	for _, l := range lines {
		parts := strings.Split(l, "\t")
		if len(parts) == 3 {
			m := match{key: parts[0], file: parts[1], lookup: parts[2]}
			if m.key == c.Key || m.key == "<"+c.Key+">" {
				matches[0] = append(matches[0], m)
			} else if strings.HasPrefix(m.key, c.Key) || strings.HasPrefix(m.key, "<"+c.Key) {
				matches[1] = append(matches[1], m)
			} else if strings.Contains(m.key, c.Key) {
				matches[2] = append(matches[2], m)
			}
		}
	}
	for i, s := range matches {
		for _, m := range s {
			fmt.Println("match level ", i)
			b, err := ioutil.ReadFile(filepath.Join(c.RuntimePath, "doc", m.file))
			if err != nil {
				return err
			}
			text := string(b)
			index := strings.Index(text, m.lookup[1:])
			if index < 0 {
				return errors.New(m.lookup + " not found")
			}
			lines = strings.Split(text[index:], "\n")
			if len(lines) > c.MaxLines {
				lines = lines[:c.MaxLines]
			}
			fmt.Printf("\u001b[1m%s\u001b[0m\n", lines[0])
			fmt.Println(strings.Join(lines[1:], "\n"))
			return nil
		}
	}
	return errors.New("not found or some error")
}

type CliHelpCmd struct {
	Options *Options
	Key     string `short:"k" long:"key" description:"Key of item" required:"yes"`
}

func (c *CliHelpCmd) Execute(args []string) error {
	fmt.Printf("Show: key=%v\n", c.Key)
	fmt.Printf("\u001b[1m\u001b[7m%s\u001b[0m\n", c.Key)
	cmd := exec.Command("man", c.Key)
	b := bytes.NewBuffer(nil)
	cmd.Stdout = b
	//cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		/*
			if cmd := exec.Command("which", os.Args[2]); true {
				cmd.Stdout = os.Stdout
				if err := cmd.Run(); err != nil {
					panic(err)
				}
			}
		*/
		cmd := exec.Command(c.Key, "--help")
		// reset
		b = bytes.NewBuffer(nil)
		cmd.Stdout = b
		b2 := bytes.NewBuffer(nil)
		cmd.Stderr = b2
		//cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			fmt.Printf("No help found for %s\n", c.Key)
			lines := strings.Split(b2.String(), "\n")
			fmt.Printf("\u001b[1m%s\u001b[0m\n", lines[0])
			fmt.Println(strings.Join(lines[1:], "\n"))
			return err
		}
	}
	lines := strings.Split(b.String(), "\n")
	fmt.Printf("\u001b[1m%s\u001b[0m\n", lines[0])
	fmt.Println(strings.Join(lines[1:], "\n"))
	return nil
}

type VimToplevel struct {
	Options *Options
	Key     string `short:"k" long:"key" description:"Key of item" required:"yes"`
}

func (c *VimToplevel) Execute(args []string) error {
	if info, ok := topLevel[c.Key]; ok {
		lines := strings.Split(info.preview, "\n")
		fmt.Printf("\u001b[1m\u001b[7m%s\u001b[0m\n", c.Key)
		fmt.Printf("\u001b[1m%s\u001b[0m\n", lines[0])
		fmt.Println(strings.Join(lines[1:], "\n"))
	}
	/*
			for k, _ := range topLevel {
				fmt.Println(k)
			}
		default:
		}
		return nil
	*/
	return nil
}

type CfgCmd struct {
	Options *Options
}

func (c *CfgCmd) Execute(args []string) error {
	fmt.Printf("Options: %+v\n", c.Options)
	return nil
}
func main() {
	var (
		options    = &Options{}
		parser     = flags.NewParser(options, flags.Default)
		vimHelpCmd = &VimHelpCmd{Options: options}
		cliHelp    = &CliHelpCmd{Options: options}
		cfgCmd     = &CfgCmd{Options: options}
	)
	parser.AddCommand("vim:help", "preview a vim help entry", "show vim:help", vimHelpCmd)
	parser.AddCommand("cli:help", "show manpage/help for a CLI command", "preview help for a command", cliHelp)
	parser.AddCommand("config", "show config", "show config", cfgCmd)

	if _, err := parser.Parse(); err != nil {
		switch flagsErr := err.(type) {
		case *flags.Error:
			if flagsErr.Type == flags.ErrHelp {
				os.Exit(0)
			}
			os.Exit(1)
		default:
			os.Exit(1)
		}
	}
}

type cmdInfo struct {
	preview string
}

var topLevel = map[string]cmdInfo{"text objects": cmdInfo{
	preview: `Do the stuffs with the text objects

 * The Operators
 * The Motions and Text Objects
 * The Niceness`,
}, "configuration": cmdInfo{
	preview: `Learning about configuration

 * The doing the config
 * The writing the config`,
}, "All-the-Things": cmdInfo{
	preview: `A big old fuzzy menu of goodness

 * FZF functionality
 * Lots of helpers
 * IDE-like stuffs`,
}}
