package scaffold

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type TemplateOptions struct {
	TemplateRoot string
	ProjectRoot  string
	Module       string
	AppName      string
	CommandName  string
	ServiceName  string
	Force        bool
	DryRun       bool
}

type TemplateResult struct {
	Written []string
	Skipped []string
	Removed []string
}

type Manifest struct {
	Name      string            `yaml:"name"`
	Kind      string            `yaml:"kind"`
	Version   string            `yaml:"version"`
	Variables map[string]string `yaml:"variables"`
	Ignore    []string          `yaml:"ignore"`
	Preserve  []string          `yaml:"preserve"`
	Obsolete  []string          `yaml:"obsolete"`
}

func ApplyTemplate(options TemplateOptions) (TemplateResult, error) {
	if strings.TrimSpace(options.TemplateRoot) == "" {
		return TemplateResult{}, errors.New("template root is required")
	}
	if strings.TrimSpace(options.ProjectRoot) == "" {
		return TemplateResult{}, errors.New("project root is required")
	}

	templateRoot, err := filepath.Abs(options.TemplateRoot)
	if err != nil {
		return TemplateResult{}, fmt.Errorf("resolve template root: %w", err)
	}
	projectRoot, err := filepath.Abs(options.ProjectRoot)
	if err != nil {
		return TemplateResult{}, fmt.Errorf("resolve project root: %w", err)
	}

	manifest, err := loadManifest(templateRoot)
	if err != nil {
		return TemplateResult{}, err
	}
	options = fillTemplateDefaults(options, manifest, projectRoot)

	files, err := collectTemplateFiles(templateRoot, manifest)
	if err != nil {
		return TemplateResult{}, err
	}

	var result TemplateResult
	for _, rel := range files {
		sourcePath := filepath.Join(templateRoot, filepath.FromSlash(rel))
		targetPath := filepath.Join(projectRoot, filepath.FromSlash(rel))

		if exists(targetPath) {
			preserved := matchesAny(rel, manifest.Preserve)
			if !options.Force || (preserved && !isGeneratedByXkit(targetPath)) {
				result.Skipped = append(result.Skipped, targetPath)
				continue
			}
		}

		content, err := os.ReadFile(sourcePath)
		if err != nil {
			return result, fmt.Errorf("read template file %s: %w", sourcePath, err)
		}
		rendered := renderTemplateContent(content, options, manifest)

		if options.DryRun {
			result.Written = append(result.Written, targetPath)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return result, fmt.Errorf("create directory for %s: %w", targetPath, err)
		}
		if err := os.WriteFile(targetPath, rendered, 0o644); err != nil {
			return result, fmt.Errorf("write %s: %w", targetPath, err)
		}
		result.Written = append(result.Written, targetPath)
	}

	if options.Force {
		removed, err := removeObsoleteFiles(projectRoot, manifest.Obsolete, options.DryRun)
		if err != nil {
			return result, err
		}
		result.Removed = append(result.Removed, removed...)
	}

	return result, nil
}

func GoGetUpdateAll(projectRoot string) (string, error) {
	if strings.TrimSpace(projectRoot) == "" {
		return "", errors.New("project root is required")
	}
	root, err := filepath.Abs(projectRoot)
	if err != nil {
		return "", fmt.Errorf("resolve project root: %w", err)
	}
	if !exists(filepath.Join(root, "go.mod")) {
		return "", fmt.Errorf("go.mod not found under %s", root)
	}

	cmd := exec.Command("go", "get", "-u", "all")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("go get -u all: %w", err)
	}
	return string(output), nil
}

func removeObsoleteFiles(projectRoot string, obsolete []string, dryRun bool) ([]string, error) {
	var removed []string
	for _, rel := range obsolete {
		rel = strings.TrimSpace(filepath.ToSlash(rel))
		if rel == "" {
			continue
		}
		targetPath, err := resolveProjectRelativePath(projectRoot, rel)
		if err != nil {
			return removed, err
		}
		if !exists(targetPath) {
			continue
		}
		removed = append(removed, targetPath)
		if dryRun {
			continue
		}
		if err := os.RemoveAll(targetPath); err != nil {
			return removed, fmt.Errorf("remove obsolete template file %s: %w", targetPath, err)
		}
	}
	return removed, nil
}

func resolveProjectRelativePath(projectRoot, rel string) (string, error) {
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("obsolete path must be relative: %s", rel)
	}
	targetPath := filepath.Clean(filepath.Join(projectRoot, filepath.FromSlash(rel)))
	relative, err := filepath.Rel(projectRoot, targetPath)
	if err != nil {
		return "", fmt.Errorf("resolve obsolete path %s: %w", rel, err)
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("obsolete path escapes project root: %s", rel)
	}
	return targetPath, nil
}

func loadManifest(templateRoot string) (Manifest, error) {
	path := filepath.Join(templateRoot, "template.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("read template manifest %s: %w", path, err)
	}

	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("parse template manifest %s: %w", path, err)
	}
	if manifest.Variables == nil {
		manifest.Variables = make(map[string]string)
	}
	return manifest, nil
}

func fillTemplateDefaults(options TemplateOptions, manifest Manifest, projectRoot string) TemplateOptions {
	if strings.TrimSpace(options.Module) == "" {
		if module, err := readModule(projectRoot); err == nil && module != "" {
			options.Module = module
		} else if module := strings.TrimSpace(manifest.Variables["module"]); module != "" {
			options.Module = module
		} else {
			options.Module = filepath.Base(projectRoot)
		}
	}
	if strings.TrimSpace(options.AppName) == "" {
		if appName := strings.TrimSpace(manifest.Variables["app_name"]); appName != "" {
			options.AppName = appName
		} else {
			options.AppName = filepath.Base(projectRoot)
		}
	}
	if strings.TrimSpace(options.CommandName) == "" {
		if commandName := strings.TrimSpace(manifest.Variables["command_name"]); commandName != "" {
			options.CommandName = commandName
		} else {
			options.CommandName = moduleBase(options.Module)
		}
	}
	if strings.TrimSpace(options.ServiceName) == "" {
		if serviceName := strings.TrimSpace(manifest.Variables["service_name"]); serviceName != "" {
			options.ServiceName = serviceName
		} else {
			options.ServiceName = options.CommandName
		}
	}
	return options
}

func collectTemplateFiles(templateRoot string, manifest Manifest) ([]string, error) {
	var files []string
	if err := filepath.WalkDir(templateRoot, func(filePath string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(templateRoot, filePath)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		if rel == "template.yaml" {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			if matchesAny(rel+"/", manifest.Ignore) {
				return filepath.SkipDir
			}
			return nil
		}
		if matchesAny(rel, manifest.Ignore) {
			return nil
		}

		files = append(files, rel)
		return nil
	}); err != nil {
		return nil, err
	}

	sort.Strings(files)
	return files, nil
}

func renderTemplateContent(content []byte, options TemplateOptions, manifest Manifest) []byte {
	replacements := map[string]string{
		"module":       options.Module,
		"app_name":     options.AppName,
		"command_name": options.CommandName,
		"service_name": options.ServiceName,
	}

	rendered := string(content)
	for name, value := range replacements {
		rendered = strings.ReplaceAll(rendered, "{{"+name+"}}", value)
		rendered = strings.ReplaceAll(rendered, "{{ "+name+" }}", value)
		rendered = strings.ReplaceAll(rendered, "__"+strings.ToUpper(name)+"__", value)
	}

	type replacement struct {
		oldValue string
		newValue string
	}
	var literalReplacements []replacement
	for name, oldValue := range manifest.Variables {
		newValue, ok := replacements[name]
		if !ok || strings.TrimSpace(oldValue) == "" {
			continue
		}
		literalReplacements = append(literalReplacements, replacement{
			oldValue: oldValue,
			newValue: newValue,
		})
	}
	sort.Slice(literalReplacements, func(i, j int) bool {
		return len(literalReplacements[i].oldValue) > len(literalReplacements[j].oldValue)
	})
	for _, replacement := range literalReplacements {
		rendered = strings.ReplaceAll(rendered, replacement.oldValue, replacement.newValue)
	}

	return []byte(rendered)
}

func matchesAny(rel string, patterns []string) bool {
	rel = filepath.ToSlash(rel)
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(filepath.ToSlash(pattern))
		if pattern == "" {
			continue
		}
		if strings.HasSuffix(pattern, "/**") {
			prefix := strings.TrimSuffix(pattern, "/**")
			if rel == prefix || strings.HasPrefix(rel, prefix+"/") {
				return true
			}
			continue
		}
		if strings.HasSuffix(pattern, "/*") {
			prefix := strings.TrimSuffix(pattern, "/*")
			if strings.HasPrefix(rel, prefix+"/") && !strings.Contains(strings.TrimPrefix(rel, prefix+"/"), "/") {
				return true
			}
			continue
		}
		if ok, _ := path.Match(pattern, rel); ok {
			return true
		}
		if ok, _ := path.Match(pattern, path.Base(rel)); ok {
			return true
		}
		if rel == pattern {
			return true
		}
	}
	return false
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func isGeneratedByXkit(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return bytes.Contains(data, []byte("Code generated by xkit. DO NOT EDIT."))
}

func readModule(projectRoot string) (string, error) {
	data, err := os.ReadFile(filepath.Join(projectRoot, "go.mod"))
	if err != nil {
		return "", err
	}
	for _, line := range bytes.Split(data, []byte{'\n'}) {
		line = bytes.TrimSpace(line)
		if bytes.HasPrefix(line, []byte("module ")) {
			return strings.TrimSpace(strings.TrimPrefix(string(line), "module ")), nil
		}
	}
	return "", errors.New("module directive not found")
}

func moduleBase(module string) string {
	module = strings.Trim(module, "/")
	if module == "" {
		return ""
	}
	parts := strings.Split(module, "/")
	return parts[len(parts)-1]
}
