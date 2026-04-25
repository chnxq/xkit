package template

import _ "embed"

//go:embed service_file.tmpl
var ServiceFile string

//go:embed service_ext.tmpl
var ServiceExt string

//go:embed repo_file.tmpl
var RepoFile string

//go:embed register_file.tmpl
var RegisterFile string

//go:embed wire_file.tmpl
var WireFile string

//go:embed bootstrap_main.tmpl
var BootstrapMain string

//go:embed bootstrap_server_cmd.tmpl
var BootstrapServerCmd string

//go:embed bootstrap_app.tmpl
var BootstrapApp string

//go:embed bootstrap_data.tmpl
var BootstrapData string

//go:embed bootstrap_ent_client.tmpl
var BootstrapEntClient string

//go:embed bootstrap_server.tmpl
var BootstrapServer string

//go:embed bootstrap_http_server.tmpl
var BootstrapHTTPServer string

//go:embed bootstrap_grpc_server.tmpl
var BootstrapGRPCServer string
