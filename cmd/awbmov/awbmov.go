package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Env:
// [DEBUG=n]
// [OUTPUT=1]
// [KEEP=1]
// [VQ=1-31], default 20
var (
	gDebug  int
	gOutput bool
	gKeep   bool
	gVQ     = "20"
)

func execCommand(debug int, output bool, cmdAndArgs []string, env map[string]string) (string, error) {
	// Execution time
	dtStart := time.Now()
	// STDOUT pipe size
	pipeSize := 0x100

	// Command & arguments
	command := cmdAndArgs[0]
	arguments := cmdAndArgs[1:]
	if debug > 0 {
		var args []string
		for _, arg := range cmdAndArgs {
			argLen := len(arg)
			if argLen > 0x200 {
				arg = arg[0:0x100] + "..." + arg[argLen-0x100:argLen]
			}
			if strings.Contains(arg, " ") {
				args = append(args, "'"+arg+"'")
			} else {
				args = append(args, arg)
			}
		}
		fmt.Printf("%s\n", strings.Join(args, " "))
	}
	cmd := exec.Command(command, arguments...)

	// Environment setup (if any)
	if len(env) > 0 {
		newEnv := os.Environ()
		for key, value := range env {
			newEnv = append(newEnv, key+"="+value)
		}
		cmd.Env = newEnv
		if debug > 0 {
			fmt.Printf("Environment Override: %+v\n", env)
			if debug > 2 {
				fmt.Printf("Full Environment: %+v\n", newEnv)
			}
		}
	}

	// Capture STDOUT (non buffered - all at once when command finishes), only used on error and when no buffered/piped version used
	// Which means it is used on error when debug <= 1
	// In debug > 1 mode, we're displaying STDOUT during execution, and storing results to 'outputStr'
	// Capture STDERR (non buffered - all at once when command finishes)
	var (
		stdOut    bytes.Buffer
		stdErr    bytes.Buffer
		outputStr string
	)
	cmd.Stderr = &stdErr
	if debug <= 1 {
		cmd.Stdout = &stdOut
	}

	// Pipe command's STDOUT during execution (if debug > 1)
	// Or just starts command when no STDOUT debug
	if debug > 1 {
		stdOutPipe, e := cmd.StdoutPipe()
		if e != nil {
			return "", e
		}
		e = cmd.Start()
		if e != nil {
			return "", e
		}
		buffer := make([]byte, pipeSize, pipeSize)
		nBytes, e := stdOutPipe.Read(buffer)
		for e == nil && nBytes > 0 {
			fmt.Printf("%s", buffer[:nBytes])
			outputStr += string(buffer[:nBytes])
			nBytes, e = stdOutPipe.Read(buffer)
		}
		if e != io.EOF {
			return "", e
		}
	} else {
		e := cmd.Start()
		if e != nil {
			return "", e
		}
	}
	// Wait for command to finish
	err := cmd.Wait()

	// If error - then output STDOUT, STDERR and error info
	if err != nil {
		if debug <= 1 {
			outStr := stdOut.String()
			if len(outStr) > 0 {
				fmt.Printf("%v\n", outStr)
			}
		}
		errStr := stdErr.String()
		if len(errStr) > 0 {
			fmt.Printf("STDERR:\n%v\n", errStr)
		}
		if err != nil {
			return stdOut.String(), err
		}
	}

	// If debug > 1 display STDERR contents as well (if any)
	if debug > 1 {
		errStr := stdErr.String()
		if len(errStr) > 0 {
			fmt.Printf("Errors:\n%v\n", errStr)
		}
	}
	if debug > 0 {
		info := strings.Join(cmdAndArgs, " ")
		lenInfo := len(info)
		if lenInfo > 0x280 {
			info = info[0:0x140] + "..." + info[lenInfo-0x140:lenInfo]
		}
		dtEnd := time.Now()
		fmt.Printf("%s: %+v\n", info, dtEnd.Sub(dtStart))
	}
	outStr := ""
	if output {
		if debug <= 1 {
			outStr = stdOut.String()
		} else {
			outStr = outputStr
		}
	}
	return outStr, nil
}

func awbmov(fn string) (err error) {
	fmt.Printf("processing '%s'\n", fn)
	// ffmpeg -i "$1" -qmin 1 -qmax "${VQ}" "${root}_%06d.png" || exit 1
	// ffmpeg -i "$1" -vn -acodec aac -ac 2 -ar 48000 -f mp4 -y "${root}.aac" || exit 2
	fnAry := strings.Split(fn, ".")
	root := fn
	if len(fnAry) > 1 {
		root = strings.Join(fnAry[:len(fnAry)-1], ".")
	}
	ch := make(chan error)
	go func(ch chan error) {
		var (
			res string
			err error
		)
		defer func() {
			ch <- err
		}()
		res, err = execCommand(
			gDebug,
			gOutput,
			[]string{"ffmpeg", "-i", fn, "-qmin", "1", "-qmax", gVQ, root + "_%06d.png"},
			nil,
		)
		if err != nil && res != "" {
			fmt.Printf("%s:\n%s\n", fn, res)
		}
	}(ch)
	go func(ch chan error) {
		var (
			res string
			err error
		)
		defer func() {
			ch <- err
		}()
		res, err = execCommand(
			gDebug,
			gOutput,
			[]string{"ffmpeg", "-i", fn, "-vn", "-acodec", "aac", "-ac", "2", "-ar", "48000", "-f", "mp4", "-y", root + ".aac"},
			nil,
		)
		if err != nil && res != "" {
			fmt.Printf("%s:\n%s\n", fn, res)
		}
	}(ch)
	err = <-ch
	if err != nil {
		return
	}
	err = <-ch
	if err != nil {
		return
	}
	return
}

func main() {
	if os.Getenv("DEBUG") != "" {
		gDebug, _ = strconv.Atoi(os.Getenv("DEBUG"))
	}
	if os.Getenv("VQ") != "" {
		iVQ, _ := strconv.Atoi(os.Getenv("VQ"))
		if iVQ > 0 {
			gVQ = strconv.Itoa(iVQ)
		}
	}
	gOutput = os.Getenv("OUTPUT") != ""
	gKeep = os.Getenv("KEEP") != ""
	for _, arg := range os.Args[1:] {
		err := awbmov(arg)
		if err != nil {
			fmt.Printf("'%s': error: %+v\n", arg, err)
		}
	}
}
