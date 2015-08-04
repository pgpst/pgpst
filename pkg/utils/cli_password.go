package utils

import (
	//"bytes"
	"fmt"
	//"io/ioutil"
	"os"
	//"os/exec"
	"runtime"
	//"strings"

	"github.com/pzduniak/speakeasy"
	//"github.com/dchest/uniuri"
)

func AskPassword(ask string) (string, error) {
	password, err := speakeasy.Ask(ask)
	if err != nil {
		if runtime.GOOS == "windows" && os.Getenv("CYGWIN") != "" {
			// We're in a Cygwin shell
			_, err := fmt.Scanln(&password)
			return password, err

			/*
				// Create a temporary file
				file, err := ioutil.TempFile("", "pgpst")
				if err != nil {
					return "", err
				}

				// I'd like to defer file.Close here, but it might not work
				file.Chmod(0777)

				// Prepare an env variable name
				env := "PGPST" + uniuri.New()

				// Write a bash script into it
				if _, err = file.WriteString(`#!/bin/bash
				save_state=$(stty -g)
				stty -echo
				read ` + env + `
				stty "$save_state"
				echo -n $` + env + `
				`); err != nil {
					file.Close()
					return "", err
				}

				// Close it
				file.Close()

				// Locate bash
				bp, err := exec.LookPath("bash")
				if err != nil {
					return "", err
				}

				// Cygwinify the name
				path := file.Name()
				path = strings.Replace(path, ":", "", -1)
				path = strings.Replace(path, "\\", "/", -1)
				path = "/" + strings.ToLower(string(path[0])) + path[1:]

				fmt.Println(path)

				out := &bytes.Buffer{}

				// Run the script
				cmd := exec.Command(bp, "-s", path)
				cmd.Stdin = os.Stdin
				cmd.Stdout = out
				cmd.Stderr = os.Stderr
				if err := cmd.Start(); err != nil {
					return "", err
				}

				x, err := out.ReadString('\n')
				fmt.Printf("%v %v\n", x, err)

				fmt.Println(out.String())

				// Read password from environment
				password := os.Getenv(env)

				if err := os.Setenv(env, ""); err != nil {
					return "", err
				}

				return password, nil
			*/
		} else {
			return "", err
		}
	}
	return password, nil
}
