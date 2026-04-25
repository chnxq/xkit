package proto

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadServices_ClassifiesMethodsAndHandlesNestedBraces(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	protoPath := filepath.Join(root, "api", "protos", "admin", "v1", "i_user.proto")
	if err := os.MkdirAll(filepath.Dir(protoPath), 0o755); err != nil {
		t.Fatalf("mkdir proto dir: %v", err)
	}

	content := `syntax = "proto3";

package admin.service.v1;

service UserService {
  rpc List (ListRequest) returns (ListResponse) {
    option (google.api.http) = {
      get: "/admin/v1/users"
      additional_bindings {
        get: "/admin/v1/users:alias"
      }
    };
  }

  rpc EditUserPassword (EditUserPasswordRequest) returns (EditUserPasswordResponse) {}
  rpc Login (LoginRequest) returns (LoginResponse) {}
}`
	if err := os.WriteFile(protoPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write proto file: %v", err)
	}

	services, err := LoadServices(root)
	if err != nil {
		t.Fatalf("load services: %v", err)
	}

	service, ok := services["admin.service.v1.UserService"]
	if !ok {
		t.Fatalf("expected service index entry")
	}

	if got, want := len(service.Methods), 3; got != want {
		t.Fatalf("method count mismatch: got %d want %d", got, want)
	}

	tests := map[string]string{
		"List":             "standard",
		"EditUserPassword": "semi-standard",
		"Login":            "special",
	}
	for _, method := range service.Methods {
		want := tests[method.Name]
		if method.Classification != want {
			t.Fatalf("classification mismatch for %s: got %s want %s", method.Name, method.Classification, want)
		}
	}
}
