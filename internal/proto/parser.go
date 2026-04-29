package proto

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Service struct {
	FullName  string
	Package   string
	Name      string
	ProtoPath string
	Methods   []Method
}

type Method struct {
	Name           string
	InputType      string
	OutputType     string
	Classification string
}

var (
	packagePattern = regexp.MustCompile(`^\s*package\s+([A-Za-z0-9_.]+)\s*;`)
	servicePattern = regexp.MustCompile(`^\s*service\s+([A-Za-z0-9_]+)\s*\{`)
	rpcPattern     = regexp.MustCompile(`^\s*rpc\s+([A-Za-z0-9_]+)\s*\(\s*([^)]+)\s*\)\s*returns\s*\(\s*([^)]+)\s*\)`)
)

func LoadServices(projectRoot string) (map[string]Service, error) {
	return LoadServicesDir(filepath.Join(projectRoot, "api", "protos"))
}

func LoadServicesDir(protoRoot string) (map[string]Service, error) {
	services := make(map[string]Service)

	err := filepath.WalkDir(protoRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".proto" {
			return nil
		}

		fileServices, err := parseFile(path)
		if err != nil {
			return err
		}
		for _, service := range fileServices {
			services[service.FullName] = service
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan proto files: %w", err)
	}

	return services, nil
}

func parseFile(path string) ([]Service, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open proto file %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var pkg string
	var current *Service
	var braceDepth int
	var services []Service

	for scanner.Scan() {
		line := scanner.Text()

		if pkg == "" {
			if matches := packagePattern.FindStringSubmatch(line); len(matches) == 2 {
				pkg = matches[1]
			}
		}

		if current == nil {
			if matches := servicePattern.FindStringSubmatch(line); len(matches) == 2 {
				current = &Service{
					Package:   pkg,
					Name:      matches[1],
					ProtoPath: path,
				}
				current.FullName = pkg + "." + current.Name
				braceDepth = 1
			}
			continue
		}

		if matches := rpcPattern.FindStringSubmatch(line); len(matches) == 4 {
			current.Methods = append(current.Methods, Method{
				Name:           matches[1],
				InputType:      strings.TrimSpace(matches[2]),
				OutputType:     strings.TrimSpace(matches[3]),
				Classification: classifyMethod(matches[1]),
			})
		}

		braceDepth += strings.Count(line, "{")
		braceDepth -= strings.Count(line, "}")
		if braceDepth == 0 {
			services = append(services, *current)
			current = nil
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan proto file %s: %w", path, err)
	}

	return services, nil
}

func classifyMethod(name string) string {
	switch name {
	case "List", "Get", "Create", "Update", "Delete", "Count", "Exists":
		return "standard"
	case "BatchCreate", "BatchDelete", "Import", "Export", "AssignRole", "EditUserPassword":
		return "semi-standard"
	}

	if strings.HasSuffix(name, "Exists") {
		return "standard"
	}
	if strings.HasPrefix(name, "Count") {
		return "standard"
	}

	return "special"
}
