package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Env:
// [DEBUG=n]
// [OUTPUT=1]
// [KEEP=1]
// [VQ=1-31], default 20
// [FPS=29.97], default autodetect
// [N_CPUS=16], default autodetect
// [IJQUAL=99], default 99
// [JQUAL=90], default 90
// [JPEG_NO_DEFAULT=1]
// [NO_JPEG=1]
// [NO_NORMALIZE=1]
// [NO_CONVERT=1] - it will still use convert for png -> jpeg conversion
// [WBRSC=white-balance-source.jpg]
// [ACM=1]
var (
	gThrN              int
	gDebug             int
	gOutput            bool
	gKeep              bool
	gJpegNoDefault     bool
	gNoJpeg            bool
	gNoNorm            bool
	gNoConvert         bool
	gFPS               string
	gVQ                = "20"
	gJqual             = "90"
	gIJqual            = "99"
	gWBSrc             string
	gDefaultJpegEnvMap = map[string]string{
		"RR":  "1",
		"RG":  "0",
		"RB":  "0",
		"GR":  "0",
		"GG":  "1",
		"GB":  "0",
		"BR":  "0",
		"BG":  "0",
		"BB":  "1",
		"RLO": ".3",
		"RHI": ".3",
		"GLO": ".3",
		"GHI": ".3",
		"BLO": ".3",
		"BHI": ".3",
		"NA":  "1",
	}
)

func toYMDHMSMi(dt time.Time) string {
	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d.%03d", dt.Year(), dt.Month(), dt.Day(), dt.Hour(), dt.Minute(), dt.Second(), dt.Nanosecond()/1000000)
}

func printf(format string, args ...interface{}) (n int, err error) {
	return fmt.Printf("%s: "+format, append([]interface{}{toYMDHMSMi(time.Now())}, args...)...)
}

func getThreadsNum() (thrN int) {
	nCPUsStr := os.Getenv("N_CPUS")
	nCPUs := 0
	if nCPUsStr != "" {
		var err error
		nCPUs, err = strconv.Atoi(nCPUsStr)
		if err != nil || nCPUs < 0 {
			nCPUs = 0
		}
	}
	if nCPUs > 0 {
		n := runtime.NumCPU()
		if nCPUs > n {
			nCPUs = n
		}
		runtime.GOMAXPROCS(nCPUs)
		thrN = nCPUs
		return
	}
	thrN = runtime.NumCPU()
	runtime.GOMAXPROCS(thrN)
	return
}

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
		printf("%s\n", strings.Join(args, " "))
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
			printf("Environment Override: %+v\n", env)
			if debug > 2 {
				printf("Full Environment: %+v\n", newEnv)
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
			printf("%s", buffer[:nBytes])
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
				printf("%v\n", outStr)
			}
		}
		errStr := stdErr.String()
		if len(errStr) > 0 {
			printf("STDERR:\n%v\n", errStr)
		}
		if err != nil {
			return stdOut.String(), err
		}
	}

	// If debug > 1 display STDERR contents as well (if any)
	if debug > 1 {
		errStr := stdErr.String()
		if len(errStr) > 0 {
			printf("Errors:\n%v\n", errStr)
		}
	}
	if debug > 0 {
		info := strings.Join(cmdAndArgs, " ")
		lenInfo := len(info)
		if lenInfo > 0x280 {
			info = info[0:0x140] + "..." + info[lenInfo-0x140:lenInfo]
		}
		dtEnd := time.Now()
		printf("%s: %+v\n", info, dtEnd.Sub(dtStart))
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
	printf("processing '%s'\n", fn)
	fnAry := strings.Split(fn, ".")
	root := fn
	if len(fnAry) > 1 {
		root = strings.Join(fnAry[:len(fnAry)-1], ".")
	}
	ch := make(chan error)
	autoDetect := false
	if gFPS == "" {
		autoDetect = true
		go func(ch chan error) {
			var (
				res string
				err error
			)
			defer func() {
				ch <- err
			}()
			// result=`ffprobe -v error -select_streams v -of default=noprint_wrappers=1:nokey=1 -show_entries stream=r_frame_rate "${1}"`
			// FPS=`echo "scale=3; ${result}" | bc`
			res, err = execCommand(
				gDebug,
				true,
				[]string{"ffprobe", "-v", "error", "-select_streams", "v", "-of", "default=noprint_wrappers=1:nokey=1", "-show_entries", "stream=r_frame_rate", fn},
				nil,
			)
			if err != nil && res != "" {
				printf("%s:\n%s\n", fn, res)
			}
			if err == nil && res != "" {
				var (
					i1 int
					i2 int
					n  int
				)
				n, err = fmt.Sscanf(res, "%d/%d", &i1, &i2)
				if err != nil {
					return
				}
				if n == 2 && i1 > 0 && i2 > 0 {
					fFPS := float64(i1) / float64(i2)
					gFPS = fmt.Sprintf("%.3f", fFPS)
				}
			}
		}(ch)
	}
	go func(ch chan error) {
		var (
			res string
			err error
		)
		defer func() {
			ch <- err
		}()
		// ffmpeg -i "$1" -qmin 1 -qmax "${VQ}" "${root}_%06d.png"
		res, err = execCommand(
			gDebug,
			gOutput,
			[]string{"ffmpeg", "-i", fn, "-qmin", "1", "-qmax", gVQ, root + "_%06d.png"},
			nil,
		)
		if err != nil && res != "" {
			printf("%s:\n%s\n", fn, res)
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
		// ffmpeg -i "$1" -vn -acodec aac -ac 2 -ar 48000 -f mp4 -y "${root}.aac"
		res, err = execCommand(
			gDebug,
			gOutput,
			[]string{"ffmpeg", "-i", fn, "-vn", "-acodec", "aac", "-ac", "2", "-ar", "48000", "-f", "mp4", "-y", root + ".aac"},
			nil,
		)
		if err != nil && res != "" {
			printf("%s:\n%s\n", fn, res)
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
	if autoDetect {
		err = <-ch
		if err != nil {
			return
		}
	}
	if gFPS == "" {
		err = fmt.Errorf("no fps specified and autodetect failed, please specify fps via FPS=29.97")
		return
	}
	printf("using %s fps, %d threads\n", gFPS, gThrN)
	fileExists := func(name string) (bool, error) {
		// printf("check '%s'\n", name)
		_, err := os.Stat(name)
		if err == nil {
			return true, nil
		}
		if errors.Is(err, os.ErrNotExist) {
			//if strings.Contains(err.Error(), "no such file or directory") {
			return false, nil
		}
		return false, err
	}
	var (
		frameSize string
		wbColor   string
		cwb       bool
	)
	getWBColor := func() error {
		var res string
		// color=`convert "${WBSRC}" -resize 1x1! -modulate 100,100,0 -format "%[pixel:u.p{0,0}]" info:`
		res, err = execCommand(
			gDebug,
			true,
			[]string{"convert", gWBSrc, "-resize", "1x1!", "-modulate", "100,100,0", "-format", "%[pixel:u.p{0,0}]", "info:"},
			nil,
		)
		if err != nil {
			if res != "" {
				printf("%s:\n%s\n", fn, res)
			}
			return err
		}
		if err == nil && res != "" {
			wbColor = res
			cwb = true
		}
		return nil
	}
	getFrameSize := func(fn string) error {
		var res string
		// size=`convert "$f" -format "%wx%h" info:`
		res, err = execCommand(
			gDebug,
			true,
			[]string{"convert", fn, "-format", "%wx%h", "info:"},
			nil,
		)
		if err != nil {
			if res != "" {
				printf("%s:\n%s\n", fn, res)
			}
			return err
		}
		if err == nil && res != "" {
			frameSize = res
		}
		return nil
	}
	var jpegEnvMap map[string]string
	if !gNoJpeg {
		if gJpegNoDefault {
			jpegEnvMap = map[string]string{"Q": gJqual}
		} else {
			jpegEnvMap = gDefaultJpegEnvMap
			jpegEnvMap["Q"] = gJqual
		}
	}
	processFrame := func(ch chan error, fn string) {
		var (
			err error
			res string
		)
		fnAry := strings.Split(fn, ".")
		root := fn
		if len(fnAry) > 1 {
			root = strings.Join(fnAry[:len(fnAry)-1], ".")
		}
		jfn := root + ".jpeg"
		qual := gIJqual
		if gNoJpeg {
			jfn = "co_" + jfn
			qual = gJqual
		}
		defer func() {
			if !gKeep {
				_ = os.Remove(fn)
				if !gNoJpeg {
					_ = os.Remove(jfn)
				}
			}
			ch <- err
		}()
		// printf("processing frame: '%s'\n", fn)
		if gNoConvert {
			res, err = execCommand(
				gDebug,
				gOutput,
				[]string{"convert", fn, "-quality", qual, jfn},
				nil,
			)
		} else {
			var cmdAndArgs []string
			if cwb {
				// convert "${f}" -colorspace sRGB \( -clone 0 -fill "$color" -colorize 50% \) -compose colorize -composite -colorspace sRGB -quality "${IJQUAL}" "${jf}"
				cmdAndArgs = []string{
					"convert", fn, "-colorspace", "sRGB", "(", "-clone", "0", "-fill", wbColor, "-colorize", "50%", ")",
					"-compose", "colorize", "-composite", "-colorspace", "sRGB", "-quality", qual,
				}
				if gNoJpeg && !gNoNorm {
					// cmdAndArgs = append(cmdAndArgs, []string{"-auto-gamma", "-auto-level"}...)
					cmdAndArgs = append(cmdAndArgs, "-normalize")
				}
				cmdAndArgs = append(cmdAndArgs, jfn)
				res, err = execCommand(
					gDebug,
					gOutput,
					cmdAndArgs,
					nil,
				)
			} else {
				// convert "${f}" \( -clone 0 -resize 1x1! -resize $size! -modulate 100,100,0 \) \( -clone 0 -fill "gray(50%)" -colorize 100 \) -compose colorize -composite -quality "${IJQUAL}" "${jf}"
				cmdAndArgs = []string{
					"convert", fn, "(", "-clone", "0", "-resize", "1x1!", "-resize", frameSize + "!", "-modulate", "100,100,0", ")",
					"(", "-clone", "0", "-fill", "gray(50%)", "-colorize", "100", ")",
					"-compose", "colorize", "-composite", "-quality", qual,
				}
				if gNoJpeg && !gNoNorm {
					// cmdAndArgs = append(cmdAndArgs, []string{"-auto-gamma", "-auto-level"}...)
					cmdAndArgs = append(cmdAndArgs, "-normalize")
				}
				cmdAndArgs = append(cmdAndArgs, jfn)
				res, err = execCommand(
					gDebug,
					gOutput,
					cmdAndArgs,
					nil,
				)
			}
		}
		if err != nil {
			if res != "" {
				printf("%s:\n%s\n", fn, res)
			}
			return
		}
		if gNoJpeg {
			return
		}
		// Q="${JQUAL}" jpeg.sh "${jf}"
		res, err = execCommand(
			gDebug,
			gOutput,
			[]string{"jpeg", jfn},
			jpegEnvMap,
		)
		if err != nil && res != "" {
			printf("%s:\n%s\n", fn, res)
		}
	}
	if gWBSrc != "" && !gNoConvert {
		err = getWBColor()
		if err != nil {
			return
		}
		if wbColor == "" {
			err = fmt.Errorf("cannot detect WBSRC color")
			return
		}
		printf("WBSRC color %s\n", wbColor)
	}
	// drop AAC and co_*.jpeg
	var exists bool
	if !gKeep {
		defer func() {
			_ = os.Remove(root + ".aac")
			dFrame := 1
			for {
				ffn := fmt.Sprintf("co_%s_%06d.jpeg", root, dFrame)
				exists, e := fileExists(ffn)
				if e != nil || !exists {
					break
				}
				_ = os.Remove(ffn)
				dFrame++
			}
		}()
	}
	frame := 1
	ch = make(chan error)
	nThreads := 0
	for {
		ffn := fmt.Sprintf("%s_%06d.png", root, frame)
		exists, err = fileExists(ffn)
		if err != nil {
			return
		}
		if !exists {
			break
		}
		if frame == 1 {
			err = getFrameSize(ffn)
			if err != nil {
				return
			}
			if frameSize == "" {
				err = fmt.Errorf("cannot detect frame size")
				return
			}
			printf("frame size %s\n", frameSize)
		}
		go processFrame(ch, ffn)
		nThreads++
		if nThreads == gThrN {
			err = <-ch
			nThreads--
			if err != nil {
				return
			}
		}
		frame++
	}
	for nThreads > 0 {
		err = <-ch
		nThreads--
		if err != nil {
			return
		}
	}
	frame--
	printf("%s: %d frames\n", fn, frame)
	//ffmpeg -framerate "${FPS}" -i "co_${root}_%06d.jpeg" -r "${FPS}" -i "${root}.aac" -s "${size}" -vcodec h264 -mbd 2 -preset slower -crf "${VQ}" -shortest -y "${root}.mp4"
	var res string
	res, err = execCommand(
		gDebug,
		gOutput,
		[]string{
			"ffmpeg", "-framerate", gFPS, "-i", "co_" + root + "_%06d.jpeg", "-r", gFPS, "-i", root + ".aac",
			"-s", frameSize, "-vcodec", "h264", "-mbd", "2", "-preset", "slower", "-crf", gVQ, "-shortest", "-y", root + ".mp4",
		},
		nil,
	)
	if err != nil {
		if res != "" {
			printf("%s:\n%s\n", fn, res)
		}
		return
	}
	printf("%s.mp4 saved\n", root)
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
	if os.Getenv("FPS") != "" {
		fFPS, _ := strconv.ParseFloat(os.Getenv("FPS"), 64)
		if fFPS > 0.0 {
			gFPS = fmt.Sprintf("%.3f", fFPS)
		}
	}
	if os.Getenv("JQUAL") != "" {
		iJqual, _ := strconv.Atoi(os.Getenv("JQUAL"))
		if iJqual > 0 {
			gJqual = strconv.Itoa(iJqual)
		}
	}
	if os.Getenv("IJQUAL") != "" {
		iIJqual, _ := strconv.Atoi(os.Getenv("IJQUAL"))
		if iIJqual > 0 {
			gIJqual = strconv.Itoa(iIJqual)
		}
	}
	gOutput = os.Getenv("OUTPUT") != ""
	gKeep = os.Getenv("KEEP") != ""
	gJpegNoDefault = os.Getenv("JPEG_NO_DEFAULT") != ""
	gNoJpeg = os.Getenv("NO_JPEG") != ""
	gNoNorm = os.Getenv("NO_NORMALIZE") != ""
	gNoConvert = os.Getenv("NO_CONVERT") != ""
	gWBSrc = os.Getenv("WBSRC")
	if gNoJpeg && gNoConvert {
		fmt.Printf("error: you cannot skip convert & jpeg tools at the same time\n")
		return
	}
	if gNoConvert && gWBSrc != "" {
		fmt.Printf("error: you can only use WBSRC with convert tool enabled\n")
		return
	}
	gThrN = getThreadsNum()
	for _, arg := range os.Args[1:] {
		dtStart := time.Now()
		err := awbmov(arg)
		dtEnd := time.Now()
		elapsed := dtEnd.Sub(dtStart)
		if err != nil {
			printf("'%s': error: %+v (took %v)\n", arg, err, elapsed)
			continue
		}
		printf("%s: ok, took %v\n", arg, elapsed)
	}
}
