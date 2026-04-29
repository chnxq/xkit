package sourceimport

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"github.com/chnxq/xkit/internal/config"
	"github.com/chnxq/xkit/internal/entschema"
	"github.com/chnxq/xkit/internal/project"
	xproto "github.com/chnxq/xkit/internal/proto"
	"gopkg.in/yaml.v3"
)

type Options struct {
	SourceRoot  string
	ProjectRoot string
	Module      string
	Service     string
	ConfigPath  string
	Force       bool
	DryRun      bool
}

type Result struct {
	Written          []string
	Skipped          []string
	SkippedResources []string
	ConfigPath       string
}

type sourceRoots struct {
	Schema string
	Proto  string
}

type messageDef struct {
	Fields      []string
	OneofFields map[string][]string
}

var (
	messagePattern = regexp.MustCompile(`^\s*message\s+([A-Za-z0-9_]+)\s*\{`)
	oneofPattern   = regexp.MustCompile(`^\s*oneof\s+([A-Za-z0-9_]+)\s*\{`)
	fieldPattern   = regexp.MustCompile(`^\s*(?:optional|required|repeated)?\s*([A-Za-z0-9_.<>]+)\s+([A-Za-z0-9_]+)\s*=\s*\d+`)
	packagePattern = regexp.MustCompile(`^\s*package\s+([A-Za-z0-9_.]+)\s*;`)
)

func Import(options Options) (Result, error) {
	if strings.TrimSpace(options.SourceRoot) == "" {
		return Result{}, fmt.Errorf("source root is required")
	}

	sourceRoot, err := filepath.Abs(options.SourceRoot)
	if err != nil {
		return Result{}, fmt.Errorf("resolve source root: %w", err)
	}
	sourceRoot = filepath.Clean(sourceRoot)

	projectRoot := strings.TrimSpace(options.ProjectRoot)
	if projectRoot == "" {
		projectRoot = "."
	}
	projectInfo, err := project.ModuleRoot(projectRoot)
	if err != nil {
		return Result{}, err
	}
	if strings.TrimSpace(options.Module) != "" {
		projectInfo.Module = strings.TrimSpace(options.Module)
	}

	service := strings.TrimSpace(options.Service)
	if service == "" {
		service = "admin"
	}

	roots, err := findSourceRoots(sourceRoot)
	if err != nil {
		return Result{}, err
	}

	schemas, err := entschema.LoadDir(roots.Schema)
	if err != nil {
		return Result{}, err
	}
	services, err := xproto.LoadServicesDir(roots.Proto)
	if err != nil {
		return Result{}, err
	}
	messages, err := loadMessages(roots.Proto)
	if err != nil {
		return Result{}, err
	}

	cfg, skipped := buildConfig(projectInfo.Module, service, schemas, services, messages)
	if len(cfg.Resources) == 0 {
		return Result{}, fmt.Errorf("no matching proto services found for schemas under %s", roots.Schema)
	}

	configPath := strings.TrimSpace(options.ConfigPath)
	if configPath == "" {
		configPath = filepath.Join(sourceRoot, filepath.Base(projectInfo.Root)+"-config", service+".yaml")
	}
	configPath, err = filepath.Abs(configPath)
	if err != nil {
		return Result{}, fmt.Errorf("resolve config path: %w", err)
	}
	configPath = filepath.Clean(configPath)

	var result Result
	result.ConfigPath = configPath
	result.SkippedResources = skipped

	if err := copyTree(roots.Proto, filepath.Join(projectInfo.Root, "api", "protos"), options, &result); err != nil {
		return result, err
	}
	if err := copyTree(roots.Schema, filepath.Join(projectInfo.Root, "internal", "data", "ent", "schema"), options, &result); err != nil {
		return result, err
	}

	configData, err := yaml.Marshal(cfg)
	if err != nil {
		return result, fmt.Errorf("marshal config: %w", err)
	}
	if err := writeFile(configPath, configData, options, &result); err != nil {
		return result, err
	}

	return result, nil
}

func findSourceRoots(sourceRoot string) (sourceRoots, error) {
	schemaCandidates := []string{
		filepath.Join(sourceRoot, "schema"),
		filepath.Join(sourceRoot, "data", "schema"),
		filepath.Join(sourceRoot, "internal", "data", "ent", "schema"),
	}
	protoCandidates := []string{
		filepath.Join(sourceRoot, "api", "protos"),
		filepath.Join(sourceRoot, "protos"),
	}

	roots := sourceRoots{}
	for _, candidate := range schemaCandidates {
		if isDir(candidate) {
			roots.Schema = candidate
			break
		}
	}
	for _, candidate := range protoCandidates {
		if isDir(candidate) {
			roots.Proto = candidate
			break
		}
	}

	if roots.Schema == "" {
		return sourceRoots{}, fmt.Errorf("source root %s is missing schema, data/schema, or internal/data/ent/schema", sourceRoot)
	}
	if roots.Proto == "" {
		return sourceRoots{}, fmt.Errorf("source root %s is missing api/protos or protos", sourceRoot)
	}
	return roots, nil
}

func copyTree(sourceRoot, targetRoot string, options Options, result *Result) error {
	return filepath.WalkDir(sourceRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(sourceRoot, path)
		if err != nil {
			return fmt.Errorf("compute relative path for %s: %w", path, err)
		}
		targetPath := filepath.Join(targetRoot, rel)
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read source file %s: %w", path, err)
		}
		return writeFile(targetPath, content, options, result)
	})
}

func writeFile(path string, content []byte, options Options, result *Result) error {
	if existing, err := os.ReadFile(path); err == nil {
		if bytes.Equal(existing, content) || !options.Force {
			result.Skipped = append(result.Skipped, path)
			return nil
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("read target file %s: %w", path, err)
	}

	if options.DryRun {
		result.Written = append(result.Written, path)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create directory for %s: %w", path, err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	result.Written = append(result.Written, path)
	return nil
}

func buildConfig(module, service string, schemas map[string]entschema.Schema, services map[string]xproto.Service, messages map[string]messageDef) (config.Config, []string) {
	schemaNames := sortedSchemaNames(schemas)
	serviceCandidates := servicesByEntity(services)
	resourceByEntity := make(map[string]config.Resource)
	var skipped []string

	for _, schemaName := range schemaNames {
		schema := schemas[schemaName]
		protoService, ok := selectProtoService(service, schemaName, serviceCandidates[schemaName])
		if !ok {
			skipped = append(skipped, schemaName)
			continue
		}

		operations, existsFields := operationsFromService(protoService, schema, messages)
		isPublicService := protoService.Package == service+".service.v1"
		hasRepo := len(operations) > 0
		if !isPublicService && !hasRepo {
			skipped = append(skipped, schemaName)
			continue
		}

		resource := config.Resource{
			Name:          toSnake(schemaName),
			ProtoService:  protoService.FullName,
			Entity:        schemaName,
			DTOImport:     packageImportPath(module, inferDTOPackage(protoService, schemaName, service)),
			DTOType:       schemaName,
			RepoInterface: schemaName + "Repo",
			ExistsFields:  existsFields,
			Filters: config.FilterConfig{
				Allow: allowedFilters(schema.Fields),
			},
			Operations: operations,
			Generate: config.GenerateFlags{
				ServiceStub:  isPublicService,
				RepoCRUD:     hasRepo,
				RestRegister: isPublicService,
				GRPCRegister: isPublicService,
			},
		}
		resourceByEntity[schemaName] = resource
	}

	if user, ok := resourceByEntity["User"]; ok {
		if _, hasCredential := resourceByEntity["UserCredential"]; hasCredential && serviceHasMethod(services[user.ProtoService], "EditUserPassword") {
			user.ServiceMethods = map[string]config.ServiceMethodConfig{
				"EditUserPassword": userEditPasswordMethod(),
			}
			resourceByEntity["User"] = user
		}
	}

	resources := make([]config.Resource, 0, len(resourceByEntity))
	for _, schemaName := range schemaNames {
		resource, ok := resourceByEntity[schemaName]
		if ok {
			resources = append(resources, resource)
		}
	}
	slices.SortFunc(resources, func(a, b config.Resource) int {
		return strings.Compare(a.Name, b.Name)
	})

	return config.Config{
		Service:   service,
		Module:    module,
		Resources: resources,
	}, skipped
}

func sortedSchemaNames(schemas map[string]entschema.Schema) []string {
	names := make([]string, 0, len(schemas))
	for name := range schemas {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func servicesByEntity(services map[string]xproto.Service) map[string][]xproto.Service {
	out := make(map[string][]xproto.Service)
	for _, service := range services {
		if !strings.HasSuffix(service.Name, "Service") {
			continue
		}
		entity := strings.TrimSuffix(service.Name, "Service")
		out[entity] = append(out[entity], service)
	}
	return out
}

func selectProtoService(targetService, entity string, candidates []xproto.Service) (xproto.Service, bool) {
	if len(candidates) == 0 {
		return xproto.Service{}, false
	}
	preferredPackage := targetService + ".service.v1"
	slices.SortFunc(candidates, func(a, b xproto.Service) int {
		if a.Package == preferredPackage && b.Package != preferredPackage {
			return -1
		}
		if a.Package != preferredPackage && b.Package == preferredPackage {
			return 1
		}
		return strings.Compare(a.FullName, b.FullName)
	})
	return candidates[0], strings.TrimSuffix(candidates[0].Name, "Service") == entity
}

func operationsFromService(service xproto.Service, schema entschema.Schema, messages map[string]messageDef) (config.OperationFlags, []string) {
	operations := make(config.OperationFlags)
	var existsFields []string
	for _, method := range service.Methods {
		switch method.Name {
		case "List":
			operations["list"] = true
		case "Get":
			operations["get"] = true
		case "Count":
			operations["count"] = true
		case "Create":
			operations["create"] = true
		case "Update":
			operations["update"] = true
		case "Delete":
			operations["delete"] = true
		case "ResetCredential":
			if schema.Name == "UserCredential" {
				operations["resetcredential"] = true
			}
		default:
			if strings.HasSuffix(method.Name, "Exists") {
				fields := existsFieldsFromRequest(method.InputType, service.Package, schema, messages)
				if len(fields) > 0 {
					operations["exists"] = true
					existsFields = fields
				}
			}
		}
	}
	return operations, existsFields
}

func existsFieldsFromRequest(typeName, servicePackage string, schema entschema.Schema, messages map[string]messageDef) []string {
	fullName := resolveProtoType(typeName, servicePackage)
	message, ok := messages[fullName]
	if !ok {
		return nil
	}
	fields := message.OneofFields["query_by"]
	if len(fields) == 0 {
		return nil
	}

	schemaFields := map[string]struct{}{"id": {}}
	for _, field := range schema.Fields {
		schemaFields[field.Name] = struct{}{}
	}
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		if _, ok := schemaFields[field]; ok {
			out = append(out, field)
		}
	}
	return out
}

func allowedFilters(fields []entschema.Field) []string {
	out := make([]string, 0, len(fields)+1)
	seen := make(map[string]struct{}, len(fields)+1)
	addFilter := func(name string, kind string) {
		if name == "" || !isFilterKind(kind) {
			return
		}
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	addFilter("id", "Uint32")
	for _, field := range fields {
		addFilter(field.Name, field.Kind)
	}
	return out
}

func isFilterKind(kind string) bool {
	switch kind {
	case "String", "Enum", "Uint", "Uint8", "Uint16", "Uint32", "Uint64", "Int", "Int8", "Int16", "Int32", "Int64":
		return true
	default:
		return false
	}
}

func inferDTOPackage(service xproto.Service, entity, targetService string) string {
	for _, method := range service.Methods {
		for _, typeName := range []string{method.OutputType, method.InputType} {
			fullName := resolveProtoType(typeName, service.Package)
			pkg, base := splitProtoType(fullName)
			if pkg == "" || isInfraProtoPackage(pkg) {
				continue
			}
			if pkg == targetService+".service.v1" {
				continue
			}
			if strings.Contains(base, entity) {
				return pkg
			}
		}
	}
	if service.Package != targetService+".service.v1" {
		return service.Package
	}
	return service.Package
}

func packageImportPath(module, pkg string) string {
	parts := strings.Split(pkg, ".")
	for index := 0; index+1 < len(parts); index++ {
		if parts[index] == "service" && strings.HasPrefix(parts[index+1], "v") {
			domain := strings.Join(parts[:index], "/")
			if domain == "" {
				domain = parts[0]
			}
			return filepath.ToSlash(filepath.Join(module, "api", "gen", domain, parts[index+1]))
		}
	}
	return filepath.ToSlash(filepath.Join(module, "api", "gen", strings.ReplaceAll(pkg, ".", "/")))
}

func resolveProtoType(typeName, currentPackage string) string {
	typeName = strings.TrimSpace(strings.TrimPrefix(typeName, "stream "))
	if typeName == "" {
		return ""
	}
	if strings.Contains(typeName, ".") {
		return typeName
	}
	return currentPackage + "." + typeName
}

func splitProtoType(fullName string) (string, string) {
	index := strings.LastIndex(fullName, ".")
	if index < 0 {
		return "", fullName
	}
	return fullName[:index], fullName[index+1:]
}

func isInfraProtoPackage(pkg string) bool {
	return pkg == "google.protobuf" || strings.HasPrefix(pkg, "pagination.")
}

func serviceHasMethod(service xproto.Service, methodName string) bool {
	for _, method := range service.Methods {
		if method.Name == methodName {
			return true
		}
	}
	return false
}

func userEditPasswordMethod() config.ServiceMethodConfig {
	return config.ServiceMethodConfig{
		Imports: []config.ImportConfig{
			{
				Alias: "authenticationv1",
				Path:  "{{module}}/api/gen/authentication/v1",
			},
		},
		Repos: []config.RepoConfig{
			{
				Field:     "userCredentialRepo",
				Interface: "UserCredentialRepo",
			},
		},
		Body: `user, err := s.{{repoField}}.Get({{ctx}}, &v11.GetUserRequest{
  QueryBy: &v11.GetUserRequest_Id{
    Id: {{param.req}}.GetUserId(),
  },
})
if err != nil {
  return nil, err
}

if _, err = s.userCredentialRepo.ResetCredential({{ctx}}, &authenticationv1.ResetCredentialRequest{
  IdentityType:  authenticationv1.UserCredential_USERNAME,
  Identifier:    user.GetUsername(),
  NewCredential: {{param.req}}.GetNewPassword(),
  NeedDecrypt:   false,
}); err != nil {
  s.log.Errorf("reset user password err: %v", err)
  return nil, err
}

return {{successReturn}}, nil`,
	}
}

func loadMessages(protoRoot string) (map[string]messageDef, error) {
	messages := make(map[string]messageDef)
	err := filepath.WalkDir(protoRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".proto" {
			return nil
		}
		fileMessages, err := parseMessagesFile(path)
		if err != nil {
			return err
		}
		for name, message := range fileMessages {
			messages[name] = message
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan proto messages: %w", err)
	}
	return messages, nil
}

func parseMessagesFile(path string) (map[string]messageDef, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read proto file %s: %w", path, err)
	}
	lines := strings.Split(string(data), "\n")
	pkg := ""
	messages := make(map[string]messageDef)

	var currentName string
	var current messageDef
	var messageDepth int
	var oneofName string
	var oneofDepth int

	for _, line := range lines {
		if comment := strings.Index(line, "//"); comment >= 0 {
			line = line[:comment]
		}
		if pkg == "" {
			if matches := packagePattern.FindStringSubmatch(line); len(matches) == 2 {
				pkg = matches[1]
			}
		}

		if currentName == "" {
			if matches := messagePattern.FindStringSubmatch(line); len(matches) == 2 {
				currentName = matches[1]
				current = messageDef{OneofFields: make(map[string][]string)}
				messageDepth = strings.Count(line, "{") - strings.Count(line, "}")
				oneofName = ""
				oneofDepth = 0
				if messageDepth <= 0 {
					fullName := currentName
					if pkg != "" {
						fullName = pkg + "." + currentName
					}
					messages[fullName] = current
					currentName = ""
				}
			}
			continue
		}

		if oneofName == "" {
			if matches := oneofPattern.FindStringSubmatch(line); len(matches) == 2 {
				oneofName = matches[1]
				oneofDepth = 0
			}
		} else if matches := fieldPattern.FindStringSubmatch(line); len(matches) == 3 {
			current.OneofFields[oneofName] = append(current.OneofFields[oneofName], matches[2])
		}

		if matches := fieldPattern.FindStringSubmatch(line); len(matches) == 3 {
			current.Fields = append(current.Fields, matches[2])
		}

		openCount := strings.Count(line, "{")
		closeCount := strings.Count(line, "}")
		if oneofName != "" {
			oneofDepth += openCount
			oneofDepth -= closeCount
			if oneofDepth <= 0 {
				oneofName = ""
				oneofDepth = 0
			}
		}
		messageDepth += openCount
		messageDepth -= closeCount
		if messageDepth <= 0 {
			fullName := currentName
			if pkg != "" {
				fullName = pkg + "." + currentName
			}
			messages[fullName] = current
			currentName = ""
		}
	}
	return messages, nil
}

func toSnake(value string) string {
	var out []rune
	var prev rune
	for index, current := range value {
		if index > 0 && unicode.IsUpper(current) && (unicode.IsLower(prev) || unicode.IsDigit(prev)) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(current))
		prev = current
	}
	return string(out)
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
