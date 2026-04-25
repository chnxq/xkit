package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/chnxq/xkit/internal/codegen"
	"github.com/chnxq/xkit/internal/config"
	"github.com/chnxq/xkit/internal/project"
)

const usageText = `Usage:
  xkit gen service <service> [--project <path>] [--config <path>] [--domain <name>] [--dry-run]
  xkit gen repo <service> [--project <path>] [--config <path>] [--domain <name>] [--dry-run]
  xkit gen register <service> [--project <path>] [--config <path>] [--domain <name>] [--dry-run]
  xkit gen wire <service> [--project <path>] [--config <path>] [--domain <name>] [--dry-run]
  xkit gen bootstrap <service> [--project <path>] [--config <path>] [--domain <name>] [--dry-run]
  xkit gen all <service> [--project <path>] [--config <path>] [--domain <name>] [--dry-run]
`

func Run(args []string, version string) error {
	if len(args) == 0 {
		printUsage(os.Stdout)
		return nil
	}

	switch args[0] {
	case "help", "-h", "--help":
		printUsage(os.Stdout)
		return nil
	case "gen":
		return runGen(args[1:], version)
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

type genOptions struct {
	projectRoot string
	configPath  string
	domain      string
	dryRun      bool
}

func runGen(args []string, version string) error {
	if len(args) < 2 {
		return errors.New("gen requires a target and service name")
	}

	target := strings.TrimSpace(args[0])
	serviceName := strings.TrimSpace(args[1])
	if target == "" || serviceName == "" {
		return errors.New("gen requires a target and service name")
	}

	var options genOptions
	flagSet := flag.NewFlagSet("gen "+target, flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&options.projectRoot, "project", "", "target project root")
	flagSet.StringVar(&options.configPath, "config", "", "path to generation config")
	flagSet.StringVar(&options.domain, "domain", "", "domain name used to resolve the default config path")
	flagSet.BoolVar(&options.dryRun, "dry-run", false, "plan file writes without modifying the target project")
	if err := flagSet.Parse(args[2:]); err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	projectInfo, err := project.Discover(options.projectRoot, cwd)
	if err != nil {
		return err
	}

	configPath := options.configPath
	if configPath == "" {
		domain := options.domain
		if domain == "" {
			domain = serviceName
		}
		configPath = filepath.Join(projectInfo.Root, "app", serviceName, "service", "gen", domain+".yaml")
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}
	if cfg.Service != serviceName {
		return fmt.Errorf("config service %q does not match argument %q", cfg.Service, serviceName)
	}

	runner, err := codegen.New(projectInfo, cfg, codegen.Options{DryRun: options.dryRun, Version: version})
	if err != nil {
		return err
	}

	result, err := runner.Generate(target)
	if err != nil {
		return err
	}

	printResult(target, options.dryRun, result)
	return nil
}

func printResult(target string, dryRun bool, result codegen.Result) {
	mode := "wrote"
	if dryRun {
		mode = "planned"
	}

	for _, path := range result.Written {
		fmt.Printf("%s %s (%s)\n", mode, path, target)
	}
	for _, path := range result.Skipped {
		fmt.Printf("skipped %s (exists)\n", path)
	}
}

func printUsage(w io.Writer) {
	_, _ = io.WriteString(w, usageText)
}
