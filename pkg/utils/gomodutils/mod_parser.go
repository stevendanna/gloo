package gomodutils

import (
	"io/ioutil"
	"os/exec"
	"strings"

	"golang.org/x/mod/modfile"
)

var (
	DefaultGoModPathGetter CurrentGoModContentGetter = func() (filePath, fileContents string, err error) {
		gomod, err := exec.Command("go", "env", "GOMOD").CombinedOutput()
		if err != nil {
			return "", "", err
		}

		filePath = strings.TrimSpace(string(gomod))
		fileContentsBytes, err := ioutil.ReadFile(filePath)
		return filePath, string(fileContentsBytes), err
	}
)

// returns the path to the go mod file for the current project as well as its contents
type CurrentGoModContentGetter func() (filePath, fileContents string, err error)

type ModParser struct {
	// if nil, will use DefaultGoModPathGetter
	CurrentGoModContentGetter CurrentGoModContentGetter
}

// parse the go mod file contents given by its CurrentGoModContentGetter field
func (m *ModParser) Parse() (modFile *modfile.File, err error) {
	contentGetter := m.CurrentGoModContentGetter
	if contentGetter == nil {
		contentGetter = DefaultGoModPathGetter
	}

	filePath, modContent, err := contentGetter()
	if err != nil {
		return nil, err
	}

	return modfile.Parse(filePath, []byte(modContent), nil)
}
