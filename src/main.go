package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/integrii/flaggy"
	jsoniter "github.com/json-iterator/go"
	"github.com/op/go-logging"
	"golang.org/x/exp/slices"
)

var VERSION = "1.1.0"

var json = jsoniter.ConfigFastest
var log = logging.MustGetLogger("gocc")
var mode = ""
var processedNum = 0
var targets = make(map[string][]string)
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

		bytes, err := io.ReadAll(file)
		check(err)

		configJSON := make(map[string]any)
		err = json.Unmarshal(bytes, &configJSON)
		check(err)

		if modeAny, ok := configJSON["mode"]; ok {
			switch value := modeAny.(type) {
			case string:
				mode = value
			default:
				flaggy.ShowHelpAndExit("Expected config key 'mode' to have type of 'string'.\n\nUsage:\n")
			}

			if mode != "allow" && mode != "disallow" {
				flaggy.ShowHelpAndExit(fmt.Sprintf("Expected config key 'mode' to have value of either 'allowed' or 'disallowed', got '%s'.\n\nUsage:\n", mode))
			}
		} else {
			flaggy.ShowHelpAndExit("Config file does not contain required key 'mode'.\n\nUsage:\n")
		}

		if targetsAny, ok := configJSON["targets"]; ok {
			targetsSlice := targetsAny.([]any)

			for _, lineAny := range targetsSlice {
				line := lineAny.(string)

				build := strings.Split(line, "/")
				build[0] = strings.TrimSpace(build[0])
				build[1] = strings.TrimSpace(build[1])

				if len(build) != 2 {
					flaggy.ShowHelpAndExit(fmt.Sprintf("Error in configuration file - Expected OS and architecture separated by a '/', found '%s'.\n\nUsage:\n", line))
				}

				targets[build[0]] = append(targets[build[0]], build[1])
			}
		} else {
			flaggy.ShowHelpAndExit("Config file does not contain required key 'targets'.\n\nUsage:\n")
		}

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
	if mode == "disallow" {
		if slices.Contains(targets["*"], build[1]) {
			return true
		}

		if slices.Contains(targets[build[0]], "*") {
			return true
		}

		if slices.Contains(targets[build[0]], build[1]) {
			return true
		}
	} else {
		found := false

		if slices.Contains(targets["*"], build[1]) {
			found = true
		}

		if slices.Contains(targets[build[0]], "*") {
			found = true
		}

		if slices.Contains(targets[build[0]], build[1]) {
			found = true
		}

		return !found
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

	cmd = exec.Command("gcc")

	if err := cmd.Run(); err == nil {
		os.Setenv("CGO_ENABLED", "1")
	}

	start := time.Now()

	for _, buildStr := range builds {
		build := strings.Split(buildStr, "/")

		if checkNotAllowed(build) {
			log.Debugf("Skipping '%s' because the config disallows it.", buildStr)
			continue
		}

		log.Debugf("Compiling for '%s'.", buildStr)
		processedNum++

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

	log.Debugf("Compilation of targets completed in %.2f seconds. %d / %d targets successfully compiled (%.2f%%).", time.Since(start).Seconds(), successfulNum, processedNum, float64(successfulNum)/float64(processedNum)*100)
}
