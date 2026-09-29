package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/NickLinney/praetorium-cli/internal/config"
	"github.com/NickLinney/praetorium-cli/internal/coordinator"
	"github.com/NickLinney/praetorium-cli/internal/render"
	"github.com/NickLinney/praetorium-cli/internal/request"
	"github.com/NickLinney/praetorium-cli/internal/systemalias"
)

var (
	Version   = "0.0.0-pre-alpha.1"
	Commit    = "unknown"
	BuildDate = "unknown"
)

const (
	exitOK          = 0
	exitUsage       = 2
	exitConfig      = 3
	exitCoordinator = 4
	exitRejected    = 5
	exitExecution   = 6
)

type globals struct {
	profile    string
	output     string
	endpoint   string
	timeout    int
	hasTimeout bool
}

func main() { os.Exit(run(os.Args[1:])) }

func run(args []string) int {
	global, remaining, err := extractGlobals(args)
	if err != nil {
		return fail(exitUsage, err)
	}
	if len(remaining) == 0 || remaining[0] == "help" || remaining[0] == "--help" || remaining[0] == "-h" {
		fmt.Print(helpText())
		return exitOK
	}
	if remaining[0] == "version" || remaining[0] == "--version" {
		fmt.Printf("praetorium %s\ncommit: %s\nbuilt: %s\ntarget: %s/%s\n", Version, Commit, BuildDate, runtimeGOOS(), runtimeGOARCH())
		return exitOK
	}
	command := remaining[0]
	commandArgs := remaining[1:]
	if command == "exec" {
		command = "execute"
	}
	if command == "prompt" {
		if len(commandArgs) == 0 {
			return fail(exitUsage, errors.New("prompt alias requires prompt text"))
		}
		commandArgs = []string{"--prompt", strings.Join(commandArgs, " ")}
		command = "execute"
	}
	switch command {
	case "execute":
		return executeCommand(global, commandArgs)
	case "inspect":
		return inspectCommand(global, commandArgs)
	case "config":
		return configCommand(global, commandArgs)
	default:
		return fail(exitUsage, fmt.Errorf("unknown command %q", command))
	}
}

func extractGlobals(args []string) (globals, []string, error) {
	result := globals{profile: "default"}
	remaining := make([]string, 0, len(args))
	for index := 0; index < len(args); index++ {
		arg := args[index]
		name, value, hasValue := splitFlag(arg)
		switch name {
		case "--profile":
			if !hasValue {
				index++
				if index >= len(args) {
					return result, nil, errors.New("--profile requires a value")
				}
				value = args[index]
			}
			result.profile = value
		case "--output":
			if !hasValue {
				index++
				if index >= len(args) {
					return result, nil, errors.New("--output requires a value")
				}
				value = args[index]
			}
			result.output = value
		case "--endpoint":
			if !hasValue {
				index++
				if index >= len(args) {
					return result, nil, errors.New("--endpoint requires a value")
				}
				value = args[index]
			}
			result.endpoint = value
		case "--timeout":
			if !hasValue {
				index++
				if index >= len(args) {
					return result, nil, errors.New("--timeout requires a value")
				}
				value = args[index]
			}
			parsed, scanErr := strconv.Atoi(value)
			if scanErr != nil {
				return result, nil, fmt.Errorf("--timeout must be an integer")
			}
			result.timeout, result.hasTimeout = parsed, true
		default:
			remaining = append(remaining, arg)
		}
	}
	return result, remaining, nil
}

func splitFlag(arg string) (string, string, bool) {
	if index := strings.IndexByte(arg, '='); index > 0 && strings.HasPrefix(arg, "--") {
		return arg[:index], arg[index+1:], true
	}
	return arg, "", false
}

func executeCommand(global globals, args []string) int {
	flags := flag.NewFlagSet("execute", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	prompt := flags.String("prompt", "", "inline prompt")
	promptFile := flags.String("prompt-file", "", "file containing prompt text")
	requestFile := flags.String("request", "", "JSON HarnessExecutionRequest file")
	if err := flags.Parse(args); err != nil {
		return fail(exitUsage, err)
	}
	if flags.NArg() > 0 {
		return fail(exitUsage, fmt.Errorf("unexpected execute argument %q", flags.Arg(0)))
	}
	sources := 0
	if *prompt != "" {
		sources++
	}
	if *promptFile != "" {
		sources++
	}
	if *requestFile != "" {
		sources++
	}
	if sources > 1 {
		return fail(exitUsage, errors.New("choose only one of --prompt, --prompt-file, or --request"))
	}
	cfg, err := loadConfig(global)
	if err != nil {
		return fail(exitConfig, err)
	}
	var body []byte
	switch {
	case *requestFile != "":
		body, err = request.LoadFile(*requestFile)
	case *promptFile != "":
		content, readErr := os.ReadFile(*promptFile)
		if readErr != nil {
			err = fmt.Errorf("read prompt %s: %w", *promptFile, readErr)
		} else {
			body, err = request.BuildPrompt(string(content), cfg.Values.TimeoutSeconds)
		}
	case *prompt != "":
		body, err = request.BuildPrompt(*prompt, cfg.Values.TimeoutSeconds)
	default:
		if isTerminalStdin() {
			return fail(exitUsage, errors.New("provide --prompt, --prompt-file, --request, or pipe a prompt on stdin"))
		}
		content, readErr := io.ReadAll(io.LimitReader(os.Stdin, 16<<20))
		if readErr != nil {
			err = fmt.Errorf("read stdin: %w", readErr)
		} else {
			body, err = request.BuildPrompt(string(content), cfg.Values.TimeoutSeconds)
		}
	}
	if err != nil {
		return fail(exitUsage, err)
	}
	responseBody, status, err := (coordinator.Client{Endpoint: cfg.Values.Endpoint, Timeout: time.Duration(cfg.Values.TimeoutSeconds) * time.Second}).Execute(context.Background(), body)
	if err != nil {
		return fail(exitCoordinator, err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return fail(exitCoordinator, fmt.Errorf("coordinator returned invalid JSON (HTTP %d): %w", status, err))
	}
	if cfg.Values.OutputFormat == "json" {
		if err := render.JSON(os.Stdout, responseBody); err != nil {
			return fail(exitCoordinator, err)
		}
	} else if err := render.HumanResult(os.Stdout, result); err != nil {
		return fail(exitCoordinator, err)
	}
	if status >= 200 && status < 300 && result["status"] == "succeeded" {
		return exitOK
	}
	if status >= 400 && status < 500 {
		return exitRejected
	}
	return exitExecution
}

func inspectCommand(global globals, args []string) int {
	flags := flag.NewFlagSet("inspect", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	requestFile := flags.String("request", "", "JSON HarnessExecutionRequest file")
	if err := flags.Parse(args); err != nil {
		return fail(exitUsage, err)
	}
	if *requestFile == "" {
		return fail(exitUsage, errors.New("inspect requires --request PATH"))
	}
	if flags.NArg() > 0 {
		return fail(exitUsage, fmt.Errorf("unexpected inspect argument %q", flags.Arg(0)))
	}
	body, err := request.LoadFile(*requestFile)
	if err != nil {
		return fail(exitUsage, err)
	}
	if global.output == "json" || global.output == "" {
		err = render.JSON(os.Stdout, body)
	} else {
		err = render.JSON(os.Stdout, body)
	}
	if err != nil {
		return fail(exitCoordinator, err)
	}
	return exitOK
}

func configCommand(global globals, args []string) int {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		fmt.Print(configHelpText())
		return exitOK
	}
	subcommand := args[0]
	if len(args) > 1 {
		return fail(exitUsage, fmt.Errorf("unexpected config argument %q", args[1]))
	}
	cfg, err := loadConfig(global)
	if err != nil {
		return fail(exitConfig, err)
	}
	switch subcommand {
	case "show":
		return printConfig(cfg, global.output)
	case "validate":
		fmt.Printf("Configuration valid for profile %q.\n", cfg.Profile)
		return exitOK
	case "apply":
		if !cfg.Values.SystemAliasEnabled {
			fmt.Println("System alias is disabled; no external state changed.")
			return exitOK
		}
		state, applyErr := systemalias.Apply(cfg.ConfigDir, cfg.Values.SystemAliasName)
		if applyErr != nil {
			return fail(exitConfig, applyErr)
		}
		fmt.Printf("System alias %q prepared at %s. Source that file to enable it in the current shell.\n", state.Name, state.Path)
		return exitOK
	case "status":
		if !cfg.Values.SystemAliasEnabled {
			fmt.Println("System alias is disabled.")
			return exitOK
		}
		state, inspectErr := systemalias.Inspect(cfg.ConfigDir, cfg.Values.SystemAliasName)
		if inspectErr != nil {
			return fail(exitConfig, inspectErr)
		}
		if global.output == "json" {
			encoded, _ := json.MarshalIndent(state, "", "  ")
			fmt.Println(string(encoded))
			return exitOK
		}
		fmt.Printf("System alias: %s\nPath: %s\nApplied: %t\n", state.Name, state.Path, state.Applied)
		if len(state.Collisions) > 0 {
			fmt.Printf("Collisions: %s\n", strings.Join(state.Collisions, "; "))
		}
		return exitOK
	default:
		return fail(exitUsage, fmt.Errorf("unknown config command %q", subcommand))
	}
}

func loadConfig(global globals) (config.Effective, error) {
	return config.Load(global.profile, config.Overrides{OutputFormat: global.output, Endpoint: global.endpoint, TimeoutSeconds: global.timeout, HasTimeout: global.hasTimeout})
}

func printConfig(cfg config.Effective, output string) int {
	if output == "json" {
		encoded, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return fail(exitConfig, err)
		}
		fmt.Println(string(encoded))
		return exitOK
	}
	fmt.Printf("Profile: %s\nConfig directory: %s\nEndpoint: %s\nTimeout: %ds\nOutput: %s\nSystem alias: %t (%s)\nCollision policy: %s\n", cfg.Profile, cfg.ConfigDir, cfg.Values.Endpoint, cfg.Values.TimeoutSeconds, cfg.Values.OutputFormat, cfg.Values.SystemAliasEnabled, cfg.Values.SystemAliasName, cfg.Values.CollisionPolicy)
	fmt.Printf("Default config: %s (%s)\n", cfg.DefaultPath, foundLabel(cfg.DefaultFound))
	if cfg.ProfilePath != "" {
		fmt.Printf("Profile config: %s (%s)\n", cfg.ProfilePath, foundLabel(cfg.ProfileFound))
	}
	return exitOK
}

func foundLabel(found bool) string {
	if found {
		return "found"
	}
	return "built-in/default"
}

func fail(code int, err error) int { fmt.Fprintf(os.Stderr, "praetorium: %s\n", err); return code }

func isTerminalStdin() bool {
	info, err := os.Stdin.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func runtimeGOOS() string   { return runtime.GOOS }
func runtimeGOARCH() string { return runtime.GOARCH }

func helpText() string {
	return `Praetorium CLI — Coordinator client

Usage:
  praetorium execute [--prompt TEXT | --prompt-file PATH | --request PATH]
  praetorium inspect --request PATH
  praetorium config {show|validate|apply|status}
  praetorium exec ...                 application alias for execute
  praetorium prompt TEXT              application alias for execute --prompt
  praetorium version

Global options:
  --profile NAME       select a named configuration profile
  --endpoint URL       override the Coordinator endpoint
  --timeout SECONDS    override the request timeout
  --output human|json  select output format

The CLI submits only to the Praetorium Coordinator; it does not call Pi directly.
`
}

func configHelpText() string {
	return `Usage: praetorium config {show|validate|apply|status}

  show       display effective configuration
  validate   validate configuration without mutation
  apply      safely prepare the configured system alias
  status     inspect observed system-alias state
`
}
