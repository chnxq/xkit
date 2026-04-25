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
	optionalCallPattern  = regexp.MustCompile(`\.\s*Optional\s*\(\s*\)`)
	nillableCallPattern  = regexp.MustCompile(`\.\s*Nillable\s*\(\s*\)`)
	immutableCallPattern = regexp.MustCompile(`\.\s*Immutable\s*\(\s*\)`)
)

func Load(projectRoot string) (map[string]Schema, error) {
	schemaRoot := filepath.Join(projectRoot, "internal", "data", "ent", "schema")
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

	return schema, nil
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
