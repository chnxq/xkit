package template

import _ "embed"

//go:embed service_file.tmpl
var ServiceFile string

//go:embed service_ext.tmpl
var ServiceExt string

//go:embed repo_file.tmpl
var RepoFile string

//go:embed repo_ext.tmpl
var RepoExt string

//go:embed register_file.tmpl
var RegisterFile string

//go:embed wire_file.tmpl
var WireFile string

//go:embed bootstrap_generated_servers.tmpl
var BootstrapGeneratedServers string

//go:embed bootstrap_ent_client.tmpl
var BootstrapEntClient string
