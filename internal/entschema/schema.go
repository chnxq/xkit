package entschema

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Schema struct {
	Name   string
	Path   string
	Fields []Field
}

type Field struct {
	Name      string
	Kind      string
	Optional  bool
	Nillable  bool
	Immutable bool
}

var (
	typePattern          = regexp.MustCompile(`type\s+([A-Z][A-Za-z0-9_]*)\s+struct\s*\{`)
	fieldPattern         = regexp.MustCompile(`field\.([A-Za-z0-9_]+)\("([^"]+)"\)`)
	mixinPattern         = regexp.MustCompile(`mixin\.([A-Za-z0-9_]+)(?:\[[^\]]+\])?\s*\{\s*\}`)
	optionalCallPattern  = regexp.MustCompile(`\.\s*Optional\s*\(\s*\)`)
	nillableCallPattern  = regexp.MustCompile(`\.\s*Nillable\s*\(\s*\)`)
	immutableCallPattern = regexp.MustCompile(`\.\s*Immutable\s*\(\s*\)`)
)

func Load(projectRoot string) (map[string]Schema, error) {
	return LoadDir(filepath.Join(projectRoot, "internal", "data", "ent", "schema"))
}

func LoadDir(schemaRoot string) (map[string]Schema, error) {
	items := make(map[string]Schema)

	err := filepath.WalkDir(schemaRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		schema, err := parseFile(path)
		if err != nil {
			return err
		}
		if schema.Name != "" {
			items[schema.Name] = schema
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan ent schema: %w", err)
	}

	return items, nil
}

func parseFile(path string) (Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Schema{}, fmt.Errorf("read ent schema %s: %w", path, err)
	}

	content := stripLineComments(string(data))
	matches := typePattern.FindStringSubmatch(content)
	if len(matches) != 2 {
		return Schema{}, nil
	}

	schema := Schema{Name: matches[1], Path: path}
	fieldMatches := fieldPattern.FindAllStringSubmatchIndex(content, -1)
	for index, match := range fieldMatches {
		kind := content[match[2]:match[3]]
		name := content[match[4]:match[5]]
		end := len(content)
		if index+1 < len(fieldMatches) {
			end = fieldMatches[index+1][0]
		}
		chain := content[match[1]:end]
		schema.Fields = append(schema.Fields, Field{
			Name:      name,
			Kind:      kind,
			Optional:  optionalCallPattern.MatchString(chain),
			Nillable:  nillableCallPattern.MatchString(chain),
			Immutable: immutableCallPattern.MatchString(chain),
		})
	}
	schema.Fields = appendMissingFields(schema.Fields, fieldsFromMixins(content)...)

	return schema, nil
}

func fieldsFromMixins(content string) []Field {
	matches := mixinPattern.FindAllStringSubmatch(content, -1)
	fields := make([]Field, 0, len(matches))
	for _, match := range matches {
		if len(match) != 2 {
			continue
		}
		fields = append(fields, knownMixinFields(match[1])...)
	}
	return fields
}

func knownMixinFields(name string) []Field {
	switch name {
	case "AutoIncrementId":
		return []Field{{Name: "id", Kind: "Uint32", Nillable: true, Immutable: true}}
	case "AutoIncrementId64":
		return []Field{{Name: "id", Kind: "Uint64", Nillable: true, Immutable: true}}
	case "OperatorID":
		return []Field{
			{Name: "created_by", Kind: "Uint32", Optional: true, Nillable: true},
			{Name: "updated_by", Kind: "Uint32", Optional: true, Nillable: true},
			{Name: "deleted_by", Kind: "Uint32", Optional: true, Nillable: true},
		}
	case "OperatorID64":
		return []Field{
			{Name: "created_by", Kind: "Uint64", Optional: true, Nillable: true},
			{Name: "updated_by", Kind: "Uint64", Optional: true, Nillable: true},
			{Name: "deleted_by", Kind: "Uint64", Optional: true, Nillable: true},
		}
	case "TimeAt":
		return []Field{
			{Name: "created_at", Kind: "Time", Optional: true, Nillable: true, Immutable: true},
			{Name: "updated_at", Kind: "Time", Optional: true, Nillable: true},
			{Name: "deleted_at", Kind: "Time", Optional: true, Nillable: true},
		}
	case "Remark":
		return []Field{{Name: "remark", Kind: "String", Optional: true, Nillable: true}}
	case "TenantID":
		return []Field{{Name: "tenant_id", Kind: "Uint32", Optional: true, Nillable: true, Immutable: true}}
	case "SortOrder":
		return []Field{{Name: "sort_order", Kind: "Uint32", Optional: true, Nillable: true}}
	case "SwitchStatus":
		return []Field{{Name: "status", Kind: "Enum", Nillable: true}}
	default:
		return nil
	}
}

func appendMissingFields(fields []Field, additions ...Field) []Field {
	seen := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		seen[field.Name] = struct{}{}
	}
	for _, field := range additions {
		if _, ok := seen[field.Name]; ok {
			continue
		}
		fields = append(fields, field)
		seen[field.Name] = struct{}{}
	}
	return fields
}

func stripLineComments(content string) string {
	lines := strings.Split(content, "\n")
	for index, line := range lines {
		if comment := strings.Index(line, "//"); comment >= 0 {
			lines[index] = line[:comment]
		}
	}
	return strings.Join(lines, "\n")
}
