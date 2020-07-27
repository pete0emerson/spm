package spm

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	// https://github.com/mattn/go-sqlite3/blob/master/_example/simple/simple.go
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func prompt(message string, options []string, defaultOption string) string {
	reader := bufio.NewReader(os.Stdin)
	text := ""
	for {
		fmt.Printf("%s [%s] (%s) > ", message, strings.Join(options, "/"), defaultOption)
		text, _ = reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)
		if stringInSlice(text, options) {
			return text
		} else {
			fmt.Printf("%s is not in [%s]\n", text, strings.Join(options, "/"))
		}
	}
}

func Remove(path string, autoConfirm bool) error {
	remove := false
	if autoConfirm {
		remove = true
	} else {
		question := fmt.Sprintf("Remove %s", path)
		text := prompt(question, []string{"yes", "no"}, "no")
		if text == "yes" {
			remove = true
		}
	}
	if remove {
		log.Infof("Removing %s", path)
		err := os.RemoveAll(path)
		if err != nil {
			return err
		}
	}
	return nil
}

func Install(URI string, destinationPath string, autoConfirm bool) error {
	install := false
	if autoConfirm {
		install = true
	} else {
		question := fmt.Sprintf("Install %s to %s", URI, destinationPath)
		text := prompt(question, []string{"yes", "no"}, "no")
		if text == "yes" {
			install = true
		}

	}
	if install {
		_, err := git.PlainClone(destinationPath, false, &git.CloneOptions{
			URL:      URI,
			Depth:    1,
			Progress: os.Stdout,
		})
		if err != nil {
			return err
		}
		err = os.RemoveAll(destinationPath + "/.git")
		if err != nil {
			return err
		}
	}
	return nil
}
