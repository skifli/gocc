package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/integrii/flaggy"
	"github.com/op/go-logging"
	"golang.org/x/exp/slices"
)

var VERSION = "1.0.0"

var log = logging.MustGetLogger("gocc")
var notAllowed = make(map[string][]string)
var successfulNum = 0

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func parseConfig() (string, string) {
	cwd, err := os.Getwd()
	check(err)

	target := ""
	dump := ""
	config := ""

	flaggy.AddPositionalValue(&target, "target", 1, true, "The path to the file to cross-compile.")
	flaggy.String(&dump, "d", "dump", "The path to the folder to dump the cross-compiled binaries in. Defaults to `build` in the cwd. The specified folder will be created if it does not exist.")
	flaggy.String(&config, "c", "config", "The path to the config file.")

	flaggy.DefaultParser.Description = "gocc: Go Cross-Compiling made easy. Get more information at https://github.com/skifli/gocc"
	flaggy.SetVersion(VERSION)

	flaggy.Parse()

	if target == "" {
		flaggy.ShowHelpAndExit("No target file to cross-compile specified.\n\nUsage:\n")
	}

	if _, err := os.Stat(target); errors.Is(err, os.ErrNotExist) {
		flaggy.ShowHelpAndExit(fmt.Sprintf("Target file '%s' to cross-compile does not exist.\n\nUsage:\n", target))
	}

	if config != "" {
		file, err := os.Open(config)

		if errors.Is(err, os.ErrNotExist) {
			flaggy.ShowHelpAndExit(fmt.Sprintf("Config file '%s' does not exist.\n\nUsage:\n", config))
		}

		check(err)
		defer func() {
			check(file.Close())
		}()

		scanner := bufio.NewScanner(file)
		lineNum := 1

		for scanner.Scan() {
			line := scanner.Text()

			if len(strings.TrimSpace(line)) == 0 || strings.HasPrefix(line, "#") {
				continue
			}

			build := strings.Split(line, "/")
			build[0] = strings.TrimSpace(build[0])
			build[1] = strings.TrimSpace(build[1])

			if len(build) != 2 {
				flaggy.ShowHelpAndExit(fmt.Sprintf("Error on config file:%d - Expected OS and architecture separated by a '/', found '%s'.", lineNum, line))
			}

			notAllowed[build[0]] = append(notAllowed[build[0]], build[1])

			lineNum++
		}

		check(scanner.Err())
		log.Debug("Parsed configuration file.")
	}

	if dump == "" {
		dump = filepath.Join(cwd, "build")
		log.Debugf("Dump directory set to '%s' because it wasn't specified.", dump)
	} else {
		log.Debugf("Dump directory set to '%s'.", dump)
	}

	err = os.MkdirAll(dump, 0700)
	check(err)

	return dump, target
}

func checkNotAllowed(build []string) bool {
	if slices.Contains(notAllowed["*"], build[1]) {
		return true
	}

	if slices.Contains(notAllowed[build[0]], "*") {
		return true
	}

	if slices.Contains(notAllowed[build[0]], build[1]) {
		return true
	}

	return false
}

func main() {
	logging.SetBackend(logging.NewBackendFormatter(logging.NewLogBackend(os.Stderr, "", 0), logging.MustStringFormatter(`%{color}[%{time:15:04:05.000}] %{level} (%{id})%{color:reset} - %{message}`)))

	dump, target := parseConfig()
	targetName := target[:len(target)-len(filepath.Ext(target))]

	cmd := exec.Command("go", "tool", "dist", "list")
	buildsBytes, err := cmd.CombinedOutput()
	check(err)

	builds := strings.FieldsFunc(string(buildsBytes), func(r rune) bool { return r == '\n' })

	log.Debug("Beginning compilation of targets.")

	for _, buildStr := range builds {
		build := strings.Split(buildStr, "/")

		if checkNotAllowed(build) {
			log.Debugf("Skipping '%s' because the config disallows it.", buildStr)
			continue
		}

		log.Debugf("Compiling for '%s'.", buildStr)

		path := ""

		if build[0] == "windows" {
			path = filepath.Join(dump, fmt.Sprintf("%s-%s-%s.exe", targetName, build[0], build[1]))
		} else {
			path = filepath.Join(dump, fmt.Sprintf("%s-%s-%s", targetName, build[0], build[1]))
		}

		cmd = exec.Command("go", "build", "-o", path, target)
		cmd.Env = append(append(os.Environ(), "GOOS="+build[0]), "GOARCH="+build[1])

		outputBytes, err := cmd.CombinedOutput()

		if err != nil {
			log.Warningf("Failed to compile for '%s'. %s: %s", buildStr, err, strings.Join(strings.FieldsFunc(string(outputBytes), func(r rune) bool { return r == '\n' }), "; "))
		} else {
			log.Debugf("Successfully compiled for '%s'.", buildStr)
			successfulNum++
		}
	}

	log.Debugf("Compilation of all targets completed. %d / %d targets successfuly compiled (%f%%)", successfulNum, len(builds), float64(successfulNum)/float64(len(builds))*100)
}
