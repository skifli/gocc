package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/akamensky/argparse"
	"github.com/goccy/go-json"
	"github.com/op/go-logging"
	"golang.org/x/exp/slices"
)

var VERSION = "v1.3.0"

var log = logging.MustGetLogger("gocc")
var mode = ""
var processedNum = 0
var targets []string
var successfulNum = 0

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func argsError(errorMsg string) {
	fmt.Printf("%s\nRun '%s -h' for help.", errorMsg, os.Args[0])
	os.Exit(1)
}

func parseConfig() (string, string) {
	cwd, err := os.Getwd()
	check(err)

	parser := argparse.NewParser("gocc", fmt.Sprintf("Go Cross-Compiling made easy (%s). Get more information at https://github.com/skifli/gocc", VERSION))

	target := parser.StringPositional(&argparse.Options{Required: true, Help: "The path to the file to cross-compile."})
	dump := parser.String("d", "dump", &argparse.Options{Required: false, Help: "The path to the folder to dump the cross-compiled binaries in. Defaults to `build` in the cwd. The specified folder will be created if it does not exist."})
	config := parser.String("c", "config", &argparse.Options{Required: false, Help: "The path to the config file."})

	err = parser.Parse(os.Args)

	if err != nil {
		panic(err)
	}
	if _, err := os.Stat(*target); errors.Is(err, os.ErrNotExist) {
		argsError(fmt.Sprintf("Target file '%s' to cross-compile does not exist.", *target))
	}

	if *config != "" {
		file, err := os.Open(*config)

		if errors.Is(err, os.ErrNotExist) {
			argsError(fmt.Sprintf("Config file '%s' does not exist.", *config))
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
				argsError("Expected config key 'mode' to have type of 'string'.")
			}

			if mode != "allow" && mode != "disallow" {
				argsError(fmt.Sprintf("Expected config key 'mode' to have value of either 'allowed' or 'disallowed', got '%s'.", mode))
			}
		} else {
			argsError("Config file does not contain required key 'mode'.")
		}

		if targetsAny, ok := configJSON["targets"]; ok {
			for _, target := range targetsAny.([]any) {
				targets = append(targets, target.(string))
			}
		} else {
			argsError("Config file does not contain required key 'targets'.")
		}

		log.Debug("Parsed configuration file.")
	}

	if *dump == "" {
		*dump = filepath.Join(cwd, "build")
		log.Debugf("Dump directory set to '%s' because it wasn't specified.", dump)
	} else {
		log.Debugf("Dump directory set to '%s'.", dump)
	}

	err = os.MkdirAll(*dump, 0700)
	check(err)

	return *dump, *target
}

func checkForUpdate() {
	resp, err := http.Get("https://api.github.com/repos/skifli/gocc/releases/latest")
	check(err)

	defer func() {
		resp.Body.Close()
	}()

	bodyBytes, err := io.ReadAll(resp.Body)
	check(err)

	bodyJson := make(map[string]any)
	err = json.Unmarshal(bodyBytes, &bodyJson)
	check(err)

	tag := bodyJson["tag_name"].(string)

	if tag != VERSION {
		log.Warningf("Update available (%s -> %s).", VERSION, tag)
	}
}

func checkNotAllowed(buildStr string, build []string) bool {
	if mode == "disallow" {
		if slices.Contains(targets, buildStr) || slices.Contains(targets, build[0]+"/*") || slices.Contains(targets, "*/"+build[1]) {
			return true
		}
	} else {
		if !slices.Contains(targets, buildStr) && !slices.Contains(targets, build[0]+"/*") && !slices.Contains(targets, "*/"+build[1]) {
			return true
		}
	}

	return false
}

func main() {
	logging.SetBackend(logging.NewBackendFormatter(logging.NewLogBackend(os.Stderr, "", 0), logging.MustStringFormatter(`%{color}[%{time:15:04:05.000}] %{level}%{color:reset} - %{message}`)))

	checkForUpdate()

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

		if checkNotAllowed(buildStr, build) {
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
