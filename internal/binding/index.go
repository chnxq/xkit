package binding

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type ServiceBinding struct {
	FullName      string
	ServiceName   string
	InterfaceName string
	ImportPath    string
	PackageName   string
	Imports       map[string]string
	Methods       []Method
}

type Method struct {
	Name    string
	Params  []string
	Results []string
}

var serviceNamePattern = regexp.MustCompile(`ServiceName:\s*"([^"]+)"`)

func Load(projectRoot, moduleName string) (map[string]ServiceBinding, error) {
	genRoot := filepath.Join(projectRoot, "api", "gen")
	bindings := make(map[string]ServiceBinding)

	err := filepath.WalkDir(genRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, "_grpc.pb.go") {
			return nil
		}

		fileBindings, err := parseFile(projectRoot, moduleName, path)
		if err != nil {
			return err
		}
		for _, binding := range fileBindings {
			bindings[binding.FullName] = binding
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan generated grpc bindings: %w", err)
	}

	return bindings, nil
}

func parseFile(projectRoot, moduleName, path string) ([]ServiceBinding, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read binding file %s: %w", path, err)
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, content, parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("parse binding file %s: %w", path, err)
	}

	relDir, err := filepath.Rel(projectRoot, filepath.Dir(path))
	if err != nil {
		return nil, fmt.Errorf("compute binding import path for %s: %w", path, err)
	}

	serviceNames := serviceNamesFromContent(string(content))
	if len(serviceNames) == 0 {
		return nil, nil
	}

	imports := make(map[string]string)
	for _, imp := range file.Imports {
		importPath, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			return nil, fmt.Errorf("parse import path in %s: %w", path, err)
		}
		alias := ""
		if imp.Name != nil {
			alias = imp.Name.Name
		} else {
			parts := strings.Split(importPath, "/")
			alias = parts[len(parts)-1]
		}
		imports[alias] = importPath
	}

	importPath := moduleName + "/" + filepath.ToSlash(relDir)
	var bindings []ServiceBinding
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			fullName, ok := serviceNames[typeSpec.Name.Name]
			if !ok {
				continue
			}

			iface, ok := typeSpec.Type.(*ast.InterfaceType)
			if !ok {
				continue
			}

			binding := ServiceBinding{
				FullName:      fullName,
				ServiceName:   shortServiceName(fullName),
				InterfaceName: typeSpec.Name.Name,
				ImportPath:    importPath,
				PackageName:   file.Name.Name,
				Imports:       imports,
			}

			for _, field := range iface.Methods.List {
				if len(field.Names) != 1 {
					continue
				}

				methodName := field.Names[0].Name
				if strings.HasPrefix(methodName, "mustEmbedUnimplemented") {
					continue
				}

				funcType, ok := field.Type.(*ast.FuncType)
				if !ok {
					continue
				}

				binding.Methods = append(binding.Methods, Method{
					Name:    methodName,
					Params:  fieldListTypes(fset, funcType.Params),
					Results: fieldListTypes(fset, funcType.Results),
				})
			}

			bindings = append(bindings, binding)
		}
	}

	return bindings, nil
}

func serviceNamesFromContent(content string) map[string]string {
	matches := serviceNamePattern.FindAllStringSubmatch(content, -1)
	names := make(map[string]string, len(matches))
	for _, match := range matches {
		if len(match) != 2 {
			continue
		}
		fullName := match[1]
		serviceName := shortServiceName(fullName)
		if serviceName == "" {
			continue
		}
		names[serviceName+"Server"] = fullName
	}
	return names
}

func fieldListTypes(fset *token.FileSet, fields *ast.FieldList) []string {
	if fields == nil {
		return nil
	}

	var types []string
	for _, field := range fields.List {
		typeText := exprString(fset, field.Type)
		repeat := 1
		if len(field.Names) > 0 {
			repeat = len(field.Names)
		}
		for range repeat {
			types = append(types, typeText)
		}
	}

	return types
}

func exprString(fset *token.FileSet, expr ast.Expr) string {
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, expr); err != nil {
		return ""
	}
	return buf.String()
}

func shortServiceName(fullName string) string {
	if fullName == "" {
		return ""
	}
	parts := strings.Split(fullName, ".")
	return parts[len(parts)-1]
}
