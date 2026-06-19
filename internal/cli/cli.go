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
	"github.com/chnxq/xkit/internal/scaffold"
	"github.com/chnxq/xkit/internal/sourceimport"
)

const defaultTemplateSource = "https://github.com/chnxq/xkit-template.git"

const usageText = `Usage:
  xkit init template [template-source] [--project <path>] [--module <module>] [--app-name <name>] [--command-name <name>] [--service-name <name>] [--force] [--dry-run] [--skip-go-get-update-all]
  xkit init source <source-path> [--project <path>] [--service <name>] [--config <path>] [--typescript-project <path>] [--force] [--dry-run]
  xkit init module <source-path> [--project <path>] [--module-name <name>] [--module-root <path>] [--service <name>] [--config <path>] [--typescript-project <path>] [--template-root <path>] [--force] [--dry-run]
  xkit gen module <module-name> <service> [--project <path>] [--module-root <path>] [--config <path>] [--domain <name>] [--typescript-project <path>] [--dry-run]
  xkit gen service <service> [--project <path>] [--config <path>] [--domain <name>] [--dry-run]
  xkit gen repo <service> [--project <path>] [--config <path>] [--domain <name>] [--dry-run]
  xkit gen register <service> [--project <path>] [--config <path>] [--domain <name>] [--dry-run]
  xkit gen wire <service> [--project <path>] [--config <path>] [--domain <name>] [--dry-run]
  xkit gen bootstrap <service> [--project <path>] [--config <path>] [--domain <name>] [--dry-run]
  xkit gen frontend-meta <service> [--project <path>] [--config <path>] [--domain <name>] [--typescript-project <path>] [--dry-run]
  xkit gen all <service> [--project <path>] [--config <path>] [--domain <name>] [--typescript-project <path>] [--dry-run]

Default template source:
  https://github.com/chnxq/xkit-template.git
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
	case "init":
		return runInit(args[1:])
	case "gen":
		switch args[1] {
		case "module":
			return runGenModule(args[2:], version)
		default:
			return runGenProject(args[1:], version)
		}
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

type initOptions struct {
	projectRoot string
	module      string
	appName     string
	commandName string
	serviceName string
	force       bool
	dryRun      bool
	skipGoGet   bool
}

func runInit(args []string) error {
	if len(args) < 1 {
		return errors.New("init requires a kind")
	}

	kind := strings.TrimSpace(args[0])
	switch kind {
	case "template":
		return runInitTemplate(args[1:])
	case "source":
		return runInitSource(args[1:])
	case "module":
		return runInitModule(args[1:])
	default:
		return fmt.Errorf("unknown init kind %q", kind)
	}
}

func runInitTemplate(args []string) error {
	templateSource := defaultTemplateSource
	flagArgs := args
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		templateSource = strings.TrimSpace(args[0])
		flagArgs = args[1:]
	}
	if templateSource == "" {
		return errors.New("init template requires a template source")
	}

	var options initOptions
	flagSet := flag.NewFlagSet("init template", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&options.projectRoot, "project", "", "target project root")
	flagSet.StringVar(&options.module, "module", "", "target module path")
	flagSet.StringVar(&options.appName, "app-name", "", "target application name")
	flagSet.StringVar(&options.commandName, "command-name", "", "target command name")
	flagSet.StringVar(&options.serviceName, "service-name", "", "target service name")
	flagSet.BoolVar(&options.force, "force", false, "overwrite existing non-preserved files")
	flagSet.BoolVar(&options.dryRun, "dry-run", false, "plan file writes without modifying the target project")
	flagSet.BoolVar(&options.skipGoGet, "skip-go-get-update-all", false, "skip running go get -u all after copying the template")
	if err := flagSet.Parse(flagArgs); err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	projectRoot := options.projectRoot
	if strings.TrimSpace(projectRoot) == "" {
		projectRoot = cwd
	}

	result, err := scaffold.ApplyTemplate(scaffold.TemplateOptions{
		TemplateRoot: templateSource,
		ProjectRoot:  projectRoot,
		Module:       options.module,
		AppName:      options.appName,
		CommandName:  options.commandName,
		ServiceName:  options.serviceName,
		Force:        options.force,
		DryRun:       options.dryRun,
	})
	if err != nil {
		return err
	}

	printScaffoldResult("template", options.dryRun, result)
	if !options.dryRun && !options.skipGoGet {
		output, err := scaffold.GoGetUpdateAll(projectRoot)
		if output != "" {
			fmt.Print(output)
			if !strings.HasSuffix(output, "\n") {
				fmt.Println()
			}
		}
		if err != nil {
			return err
		}
		fmt.Printf("ran go get -u all (%s)\n", projectRoot)
	}
	return nil
}

type sourceOptions struct {
	projectRoot    string
	serviceName    string
	configPath     string
	typeScriptRoot string
	force          bool
	dryRun         bool
}

func runInitSource(args []string) error {
	if len(args) < 1 {
		return errors.New("init source requires a source path")
	}

	sourceRoot := strings.TrimSpace(args[0])
	if sourceRoot == "" {
		return errors.New("init source requires a source path")
	}

	var options sourceOptions
	flagSet := flag.NewFlagSet("init source", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&options.projectRoot, "project", "", "target project root")
	flagSet.StringVar(&options.serviceName, "service", "admin", "service name used in generated xkit config")
	flagSet.StringVar(&options.configPath, "config", "", "path to write generated xkit config")
	flagSet.StringVar(&options.typeScriptRoot, "typescript-project", "", "target TypeScript project root; relative paths are resolved beside the Go project")
	flagSet.BoolVar(&options.force, "force", false, "overwrite existing target files")
	flagSet.BoolVar(&options.dryRun, "dry-run", false, "plan file writes without modifying the target project")
	if err := flagSet.Parse(args[1:]); err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	projectInfo, err := project.DiscoverModule(options.projectRoot, cwd)
	if err != nil {
		return err
	}

	result, err := sourceimport.Import(sourceimport.Options{
		SourceRoot:     sourceRoot,
		ProjectRoot:    projectInfo.Root,
		Module:         projectInfo.Module,
		Service:        options.serviceName,
		ConfigPath:     options.configPath,
		TypeScriptRoot: options.typeScriptRoot,
		Force:          options.force,
		DryRun:         options.dryRun,
	})
	if err != nil {
		return err
	}

	printSourceImportResult(options.dryRun, result)
	return nil
}

type moduleInitOptions struct {
	projectRoot    string
	moduleName     string
	moduleRoot     string
	serviceName    string
	configPath     string
	typeScriptRoot string
	templateRoot   string
	force          bool
	dryRun         bool
}

func runInitModule(args []string) error {
	if len(args) < 1 {
		return errors.New("init module requires a source path")
	}

	sourceRoot := strings.TrimSpace(args[0])
	if sourceRoot == "" {
		return errors.New("init module requires a source path")
	}

	var options moduleInitOptions
	flagSet := flag.NewFlagSet("init module", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&options.projectRoot, "project", "", "target host project root")
	flagSet.StringVar(&options.moduleName, "module-name", "", "target module name")
	flagSet.StringVar(&options.moduleRoot, "module-root", "", "target module root; defaults to <project>/modules/<module-name>")
	flagSet.StringVar(&options.serviceName, "service", "admin", "service name used in generated xkit config")
	flagSet.StringVar(&options.configPath, "config", "", "path to write generated module config")
	flagSet.StringVar(&options.typeScriptRoot, "typescript-project", "", "target TypeScript project root; relative paths are resolved beside the Go project")
	flagSet.StringVar(&options.templateRoot, "template-root", "", "template root used to copy module assets; defaults to sibling xkit-template")
	flagSet.BoolVar(&options.force, "force", false, "overwrite existing target files")
	flagSet.BoolVar(&options.dryRun, "dry-run", false, "plan file writes without modifying the target project")
	if err := flagSet.Parse(args[1:]); err != nil {
		return err
	}

	if strings.TrimSpace(options.moduleName) == "" {
		return errors.New("init module requires --module-name")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	projectInfo, err := project.DiscoverModule(options.projectRoot, cwd)
	if err != nil {
		return err
	}

	templateRoot := strings.TrimSpace(options.templateRoot)
	if templateRoot == "" {
		templateRoot = filepath.Join(filepath.Dir(cwd), "xkit-template")
	}

	result, err := sourceimport.ImportModule(sourceimport.ModuleOptions{
		SourceRoot:     sourceRoot,
		ProjectRoot:    projectInfo.Root,
		Module:         projectInfo.Module,
		ModuleName:     options.moduleName,
		ModuleRoot:     options.moduleRoot,
		Service:        options.serviceName,
		ConfigPath:     options.configPath,
		TypeScriptRoot: options.typeScriptRoot,
		TemplateRoot:   templateRoot,
		Force:          options.force,
		DryRun:         options.dryRun,
	})
	if err != nil {
		return err
	}

	printSourceImportResult(options.dryRun, result)
	return nil
}

type genOptions struct {
	projectRoot    string
	moduleRoot     string
	configPath     string
	domain         string
	typeScriptRoot string
	dryRun         bool
}

func runGenProject(args []string, version string) error {
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
	flagSet.StringVar(&options.typeScriptRoot, "typescript-project", "", "target TypeScript project root; relative paths are resolved beside the Go project")
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

	runner, err := codegen.NewProjectRunner(projectInfo, cfg, codegen.Options{
		DryRun:         options.dryRun,
		Version:        version,
		TypeScriptRoot: options.typeScriptRoot,
	})
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

func runGenModule(args []string, version string) error {
	if len(args) < 2 {
		return errors.New("gen module requires a module name and service name")
	}

	moduleName := strings.TrimSpace(args[0])
	serviceName := strings.TrimSpace(args[1])
	if moduleName == "" || serviceName == "" {
		return errors.New("gen module requires a module name and service name")
	}

	var options genOptions
	flagSet := flag.NewFlagSet("gen module", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&options.projectRoot, "project", "", "target project root")
	flagSet.StringVar(&options.moduleRoot, "module-root", "", "target module root; defaults to <project>/modules/<module-name>")
	flagSet.StringVar(&options.configPath, "config", "", "path to generation config")
	flagSet.StringVar(&options.domain, "domain", "", "domain name used to resolve the default config path")
	flagSet.StringVar(&options.typeScriptRoot, "typescript-project", "", "target TypeScript project root; relative paths are resolved beside the Go project")
	flagSet.BoolVar(&options.dryRun, "dry-run", false, "plan file writes without modifying the target project")
	if err := flagSet.Parse(args[2:]); err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	projectInfo, err := project.DiscoverModule(options.projectRoot, cwd)
	if err != nil {
		return err
	}

	moduleRoot := strings.TrimSpace(options.moduleRoot)
	if moduleRoot == "" {
		moduleRoot = filepath.Join(projectInfo.Root, "modules", moduleName)
	}
	if !filepath.IsAbs(moduleRoot) {
		moduleRoot = filepath.Join(projectInfo.Root, moduleRoot)
	}

	configPath := options.configPath
	if configPath == "" {
		configPath = filepath.Join(moduleRoot, moduleName+".yaml")
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}
	if cfg.Service != serviceName {
		return fmt.Errorf("config service %q does not match argument %q", cfg.Service, serviceName)
	}

	runner, err := codegen.NewModuleRunner(projectInfo, cfg, codegen.Options{
		DryRun:         options.dryRun,
		Version:        version,
		TypeScriptRoot: options.typeScriptRoot,
		ModuleName:     moduleName,
		ModuleRoot:     moduleRoot,
	})
	if err != nil {
		return err
	}

	result, err := runner.Generate("module")
	if err != nil {
		return err
	}

	printResult("module", options.dryRun, result)
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

func printScaffoldResult(target string, dryRun bool, result scaffold.TemplateResult) {
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
	removeMode := "removed"
	if dryRun {
		removeMode = "planned remove"
	}
	for _, path := range result.Removed {
		fmt.Printf("%s %s (%s)\n", removeMode, path, target)
	}
}

func printSourceImportResult(dryRun bool, result sourceimport.Result) {
	mode := "wrote"
	if dryRun {
		mode = "planned"
	}

	for _, path := range result.Written {
		fmt.Printf("%s %s (source)\n", mode, path)
	}
	for _, path := range result.Skipped {
		fmt.Printf("skipped %s (exists)\n", path)
	}
	if len(result.SkippedResources) > 0 {
		fmt.Printf("skipped resources without matching proto service: %s\n", strings.Join(result.SkippedResources, ", "))
	}
	fmt.Printf("config %s\n", result.ConfigPath)
	fmt.Printf("typescript %s\n", result.TypeScriptRoot)
}

func printUsage(w io.Writer) {
	_, _ = io.WriteString(w, usageText)
}
