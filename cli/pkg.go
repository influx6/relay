package cli

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var multispaces = regexp.MustCompile(`\s+`)

// GoDeps calls go get for specific package
func GoDeps(targetdir string) (bool, error) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("godeps.Error: %+s", err)
		}
	}()

	cmdline := []string{"go", "get"}

	cmdline = append(cmdline, targetdir)

	//setup the executor and use a shard buffer
	cmd := exec.Command("go", cmdline[1:]...)
	buf := bytes.NewBuffer([]byte{})
	cmd.Stdout = buf
	cmd.Stderr = buf

	err := cmd.Run()

	if buf.Len() > 0 {
		return false, fmt.Errorf("go get failed: %s: %s", buf.String(), err.Error())
	}

	return true, nil
}

// GoRun runs the runs a command
func GoRun(cmd string) string {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("gorun.Error: %+s", err)
		}
	}()
	var cmdline []string
	com := strings.Split(cmd, " ")

	if len(com) < 0 {
		return ""
	}

	if len(com) == 1 {
		cmdline = append(cmdline, com...)
	} else {
		cmdline = append(cmdline, com[0])
		cmdline = append(cmdline, com[1:]...)
	}

	//setup the executor and use a shard buffer
	cmdo := exec.Command(cmdline[0], cmdline[1:]...)
	buf := bytes.NewBuffer([]byte{})
	cmdo.Stdout = buf
	cmdo.Stderr = buf

	_ = cmdo.Run()

	return buf.String()
}

// Gobuild runs the build process and returns true/false and an error
func Gobuild(dir, name string) (bool, error) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("gobuild.Error: %+s", err)
		}
	}()

	cmdline := []string{"go", "build"}

	if runtime.GOOS == "windows" {
		name = fmt.Sprintf("%s.exe", name)
	}

	target := filepath.Join(dir, name)
	cmdline = append(cmdline, "-o", target)

	//setup the executor and use a shard buffer
	cmd := exec.Command("go", cmdline[1:]...)
	buf := bytes.NewBuffer([]byte{})

	msg, err := cmd.CombinedOutput()

	if !cmd.ProcessState.Success() {
		return false, fmt.Errorf("go.build failed: %s: %s", buf.String(), err.Error())
	}

	log.Printf("go.build sucessfully -> %s", msg)

	return true, nil
}

// RunCMD runs the a set of commands from a list while skipping any one-length command
func RunCMD(cmds []string, done func()) chan bool {
	if len(cmds) < 0 {
		panic("commands list cant be empty")
	}

	var relunch = make(chan bool)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("cmdRun.Error: %+s", err)
			}
		}()

	cmdloop:
		for {
			select {
			case do, ok := <-relunch:

				if !ok {
					break cmdloop
				}

				if !do {
					continue
				}

				for _, cox := range cmds {

					cmd := strings.Split(cox, " ")

					if len(cmd) <= 1 {
						continue
					}

					fmt.Printf("-> Executing Command: %s\n", cmd)
					cmdo := exec.Command(cmd[0], cmd[1:]...)
					cmdo.Stdout = os.Stdout
					cmdo.Stderr = os.Stderr

					if err := cmdo.Start(); err != nil {
						log.Printf("-> -> Error executing command: %s -> %s", cmd, err)
					}
				}

				if done != nil {
					done()
				}
			}
		}

	}()
	return relunch
}

// RunGo runs the generated binary file with the arguments expected
func RunGo(gofile string, args []string, stop func()) chan bool {
	var relunch = make(chan bool)

	// if runtime.GOOS == "windows" {
	gofile = filepath.Clean(gofile)
	// }

	go func() {

		// var cmdline = fmt.Sprintf("go run %s", gofile)
		cmdargs := append([]string{"run", gofile}, args...)
		// cmdline = strings.Joinappend([]string{}, "go run", gofile)

		var proc *os.Process

		for dosig := range relunch {
			if proc != nil {
				var err error

				if runtime.GOOS == "windows" {
					err = proc.Kill()
				} else {
					err = proc.Signal(os.Interrupt)
				}

				if err != nil {
					log.Printf("Error in Sending Kill Signal %s", err)
					proc.Kill()
				}
				proc.Wait()
			}

			if !dosig {
				continue
			}

			cmd := exec.Command("go", cmdargs...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Start(); err != nil {
				log.Printf("Error starting process: %s", err)
			}

			proc = cmd.Process
		}

		if stop != nil {
			stop()
		}
	}()
	return relunch
}

// RunBin runs the generated binary file with the arguments expected
func RunBin(bindir, bin string, args []string, stop func()) chan bool {
	var relunch = make(chan bool)
	go func() {
		binfile := fmt.Sprintf("%s/%s", bindir, bin)
		// cmdline := append([]string{bin}, args...)
		var proc *os.Process

		for dosig := range relunch {
			if proc != nil {
				var err error

				if runtime.GOOS == "windows" {
					err = proc.Kill()
				} else {
					err = proc.Signal(os.Interrupt)
				}

				if err != nil {
					log.Printf("Error in Sending Kill Signal %s", err)
					proc.Kill()
				}
				proc.Wait()
			}

			if !dosig {
				continue
			}

			cmd := exec.Command(binfile, args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Start(); err != nil {
				log.Printf("Error starting process: %s", err)
			}

			proc = cmd.Process
		}

		if stop != nil {
			stop()
		}
	}()
	return relunch
}
