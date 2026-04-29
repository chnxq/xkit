package project

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const EnvProjectRoot = "XKIT_PROJECT_ROOT"

type Info struct {
	Root   string
	Module string
}

func Discover(explicitRoot, cwd string) (Info, error) {
	return discover(explicitRoot, cwd, validate)
}

func DiscoverModule(explicitRoot, cwd string) (Info, error) {
	return discover(explicitRoot, cwd, validateModule)
}

func ModuleRoot(root string) (Info, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return Info{}, fmt.Errorf("resolve project root: %w", err)
	}
	return validateModule(filepath.Clean(abs))
}

func discover(explicitRoot, cwd string, validator func(string) (Info, error)) (Info, error) {
	var candidates []string
	if explicitRoot != "" {
		candidates = append(candidates, explicitRoot)
	}

	if envRoot := strings.TrimSpace(os.Getenv(EnvProjectRoot)); envRoot != "" {
		candidates = append(candidates, envRoot)
	}

	for _, ancestor := range walkUp(cwd) {
		candidates = append(candidates, ancestor)
		candidates = append(candidates, filepath.Join(ancestor, "..", "xadmin-web"))
	}

	seen := make(map[string]struct{})
	var attempted []string
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		abs, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		abs = filepath.Clean(abs)
		if _, ok := seen[abs]; ok {
			continue
		}
		seen[abs] = struct{}{}
		attempted = append(attempted, abs)

		info, err := validator(abs)
		if err == nil {
			return info, nil
		}
	}

	if len(attempted) == 0 {
		return Info{}, errors.New("no project roots were checked")
	}

	return Info{}, fmt.Errorf("unable to discover target project root; checked: %s", strings.Join(attempted, ", "))
}

func validate(root string) (Info, error) {
	goModPath := filepath.Join(root, "go.mod")
	apiProtosPath := filepath.Join(root, "api", "protos")

	if !isFile(goModPath) {
		return Info{}, fmt.Errorf("%s is missing go.mod", root)
	}
	if !isDir(apiProtosPath) {
		return Info{}, fmt.Errorf("%s is missing api/protos", root)
	}

	moduleName, err := readModule(goModPath)
	if err != nil {
		return Info{}, err
	}

	return Info{
		Root:   root,
		Module: moduleName,
	}, nil
}

func validateModule(root string) (Info, error) {
	goModPath := filepath.Join(root, "go.mod")
	if !isFile(goModPath) {
		return Info{}, fmt.Errorf("%s is missing go.mod", root)
	}

	moduleName, err := readModule(goModPath)
	if err != nil {
		return Info{}, err
	}

	return Info{
		Root:   root,
		Module: moduleName,
	}, nil
}

func readModule(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open go.mod: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "module ") {
			continue
		}
		moduleName := strings.TrimSpace(strings.TrimPrefix(line, "module "))
		if moduleName == "" {
			break
		}
		return moduleName, nil
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scan go.mod: %w", err)
	}

	return "", fmt.Errorf("module declaration not found in %s", path)
}

func walkUp(start string) []string {
	if start == "" {
		return nil
	}

	abs, err := filepath.Abs(start)
	if err != nil {
		return nil
	}

	var roots []string
	current := filepath.Clean(abs)
	for {
		roots = append(roots, current)
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return roots
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
