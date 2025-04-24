package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/fatih/color"
	"golang.design/x/clipboard"
	"golang.org/x/term"
)

var (
	version string
	build   string
	logFile *os.File
	dateRegex *regexp.Regexp = regexp.MustCompile(`\${DATETIME(_[A-z]{3}|[+-][0-9]{1,2}|)}`)
	VarMap  map[string]strReturn = make(map[string]strReturn)
)

type strReturn interface {
	String() string
}

type simpleString struct {
	strReturn
	Value string
}

func (s simpleString) String() string {
	return s.Value
}

type dateVar struct {
	strReturn
	Timezone *time.Location
}

func (d dateVar) String() string {
	return time.Now().In(d.Timezone).Format("2006-01-02T03:04:05PM")
}

func init() {
	var err error
	logFile, err = os.Create("log.txt")
	if err != nil {
		fmt.Println("ERROR CREATING LOGFILE: ", err.Error())
		os.Exit(1)
	}
	if err := clipboard.Init(); err != nil {
		fmt.Println("ERROR INITIALIZING CLIPBOARD: ", err.Error())
		os.Exit(1)
	}
}

func parseFile(path string) {
	// Open File
	fh, err := os.Open(path)
	if err != nil {
		fmt.Println("ERROR OPENING FILE: ", err.Error())
		return
	}
	defer fh.Close()

	// Read File
	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		ln := strings.TrimSpace(scanner.Text())
		if len(ln) == 0 {
			continue
		}

		fancyLn, plainLn := ReplaceVars(ln)

		switch plainLn[0] {
		case '#':
			fmt.Println(color.GreenString(strings.TrimSpace(fancyLn[1:])))
		case '!':
			RunCommand(strings.TrimSpace(plainLn[1:]))
		case '$':
			part := strings.SplitN(plainLn[1:], "=", 2)
			if len(part) == 2 {
				ReadVar(strings.TrimSpace(part[0]), strings.TrimSpace(part[1]))
				break
			}
			fallthrough
		default:
			WriteCommand(strings.TrimSpace(plainLn), strings.TrimSpace(fancyLn))
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("ERROR READING FILE: ", err.Error())
		return
	}

	// Close File
	if err := fh.Close(); err != nil {
		fmt.Println("ERROR CLOSING FILE: ", err.Error())
		return
	}
}

// Collect Input

// Process Commands
func WriteCommand(clip, prompt string) {
	fmt.Fprintln(logFile, clip)
	clipboard.Write(clipboard.FmtText, []byte(clip))
	fmt.Println(prompt)
	fmt.Printf("\n%s", color.BlueString("Press Enter to Continue..."))
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func RunCommand(prompt string) {
	part := strings.Split(prompt[1:], " ")
	fmt.Println(color.YellowString(fmt.Sprintf("Executing command: %s", prompt)))
	cmd := exec.Command(part[0], part[1:]...)
	if err := cmd.Start(); err != nil {
		fmt.Println("ERROR RUNNING COMMAND: ", err.Error())
	}
	if err := cmd.Process.Release(); err != nil {
		fmt.Println("ERROR RELEASING COMMAND: ", err.Error())
	}
}

func ReadVar(name, prompt string) {
	fmt.Printf("%s: ", color.MagentaString(prompt))
	response, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	// Sanitize - trim whitespace and quotes
	response = strings.TrimFunc(response, func(r rune) bool {
		if r == '\'' || r == '"' {
			return true
		}
		return !unicode.IsLetter(r) && !unicode.IsNumber(r) && !unicode.IsPunct(r)
	})
	VarMap[name] = simpleString{Value: response}
	fmt.Fprintf(logFile, "%s = %s\n", prompt, response)
}

func ReplaceVars(prompt string) (string, string) {
	color_prompt := prompt
	plain_prompt := prompt
	if loc := dateRegex.FindStringSubmatchIndex(prompt); loc != nil {
		if len(loc) == 2 { // No timezone specified - use UTC
			color_prompt = color_prompt[:loc[0]] + color.MagentaString(time.Now().In(time.UTC).Format("2006-01-02T03:04:05PM")) + color_prompt[loc[1]:]
			plain_prompt = plain_prompt[:loc[0]] + time.Now().In(time.UTC).Format("2006-01-02T03:04:05PM") + plain_prompt[loc[1]:]
		} else if len(loc) == 4 { // Timezone specified
			zone := color_prompt[loc[2]:loc[3]]
			if len(zone) == 4 { // Timezone in format _EST
				zone = strings.ToUpper(zone)[1:] // Format timezone
				if timeLoc, err := time.LoadLocation(zone); err != nil {
					fmt.Fprintf(logFile, "ERROR: unable to load timezone %s\n", zone)
				} else {
					color_prompt = color_prompt[:loc[0]] + color.MagentaString(time.Now().In(timeLoc).Format("2006-01-02T03:04:05PM")) + color_prompt[loc[1]:]
					plain_prompt = plain_prompt[:loc[0]] + time.Now().In(timeLoc).Format("2006-01-02T03:04:05PM") + plain_prompt[loc[1]:]
				}
			} else { // Timezone in format +5
				if offset, err := strconv.Atoi(zone); err != nil {
					fmt.Fprintf(logFile, "ERROR: unable to parse timezone offset %s\n", zone)
				} else {
					color_prompt = color_prompt[:loc[0]] + color.MagentaString(time.Now().UTC().Add(time.Duration(offset)*time.Hour).Format("2006-01-02T03:04:05PM")) + color_prompt[loc[1]:]
					plain_prompt = plain_prompt[:loc[0]] + time.Now().UTC().Add(time.Duration(offset)*time.Hour).Format("2006-01-02T03:04:05PM") + plain_prompt[loc[1]:]
				}
			}
		}
	}
	for key, value := range VarMap {
		color_prompt = strings.ReplaceAll(color_prompt, "$"+key, color.MagentaString(value.String()))
		plain_prompt = strings.ReplaceAll(plain_prompt, "$"+key, value.String())
	}
	return color_prompt, plain_prompt
}

func SpawnShell() {
	fullPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(logFile, "ERROR: failed to get path to self\nPlease run in an interactive shell.\n\t%s", err.Error())
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		if path, err := exec.LookPath("cmd"); err != nil {
			fmt.Fprintf(logFile, "ERROR: unable to find cmd shell\nPlease run in an interactive shell.\n")
			os.Exit(1)
		} else {
			cmd = exec.Command(path, "/c", fullPath)
		}
	case "linux":
		if path, err := exec.LookPath("xterm"); err != nil {
			fmt.Fprintf(logFile, "ERROR: unable to find xterm\nPlease run in an interactive shell.\n")
			os.Exit(1)
		} else {
			cmd = exec.Command(path, "-e", fullPath)
		}
	case "darwin":
		if _, err := os.Stat("/System/Applications/Utilities/Terminal.app/Contents/MacOS/Terminal"); os.IsNotExist(err) {
			fmt.Fprintf(logFile, "ERROR: unable to find Terminal.app\nPlease run in an interactive shell.\n")
			os.Exit(1)
		} else {
			cmd = exec.Command("/System/Applications/Utilities/Terminal.app/Contents/MacOS/Terminal", fullPath)
		}
	default:
		fmt.Fprintf(logFile, "ERROR: unable to spawn shell on %s\nPlease run in an interactive shell.\n", runtime.GOOS)
		os.Exit(1)
	}

	// Spawn shell
	if err := cmd.Start(); err != nil {
		fmt.Fprintln(logFile, err.Error())
		os.Exit(1)
	}
	if err := cmd.Process.Release(); err != nil {
		fmt.Fprintln(logFile, err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func verifyScript() bool {
	path := VarMap["{SCRIPT_PATH}"].String()
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("ERROR: SCRIPT NOT FOUND: %s\n", path)
		return false
	}
	// Generate file checksum
	scriptFile, err := os.Open(path)
	if err != nil {
		fmt.Printf("ERROR OPENING SCRIPT: %s\n", err.Error())
		return false
	}
	scriptData, err := io.ReadAll(scriptFile)
	if err != nil {
		fmt.Printf("ERROR READING SCRIPT: %s\n", err.Error())
		return false
	}
	VarMap["{SCRIPT_CHECKSUM}"] = simpleString{Value: fmt.Sprintf("%x", sha256.Sum256(scriptData))}
	return true
}

func main() {
	defer func() {
		fmt.Println("Script complete")
		fmt.Println("Press Ctrl-c to end program and close window..")
		os.Stdin.WriteTo(io.Discard)
	}()
	defer logFile.Close()

	// Respawn in terminal if not found
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		SpawnShell()
	}

	// Check if script was passed in
	if len(os.Args) > 1 {
		if os.Args[1] == "--version" {
			fmt.Println("Version: ", version)
			fmt.Println("Build: ", build)
			os.Exit(0)
		}
		filepath := strings.Trim(os.Args[1], "\"'")
		VarMap["{SCRIPT_PATH}"] = simpleString{Value: filepath}
	} else {
		VarMap["{SCRIPT_PATH}"] = simpleString{}
	}
	// Verify script file exists
	for !verifyScript() {
		ReadVar("{SCRIPT_PATH}", "Enter path to script")
	}
	fmt.Fprintf(logFile, "Using script: %s\nChecksum: %s\n", VarMap["{SCRIPT_PATH}"].String(), VarMap["{SCRIPT_CHECKSUM}"].String())

	parseFile(VarMap["{SCRIPT_PATH}"].String())

	fmt.Fprintf(logFile, "Completed: %s\n", time.Now().String())

	if err := logFile.Close(); err != nil {
		fmt.Println("ERROR CLOSING LOGFILE: ", err.Error())
		return
	}
}
