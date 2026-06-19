package template

import _ "embed"

//go:embed service_file.tmpl
var ServiceFile string

//go:embed service_ext.tmpl
var ServiceExt string

//go:embed service_shared_ext.tmpl
var ServiceSharedExt string

//go:embed repo_file.tmpl
var RepoFile string

//go:embed repo_ext.tmpl
var RepoExt string

//go:embed repo_shared_ext.tmpl
var RepoSharedExt string

//go:embed register_file.tmpl
var RegisterFile string

//go:embed wire_file.tmpl
var WireFile string

//go:embed bootstrap_generated_servers.tmpl
var BootstrapGeneratedServers string

//go:embed bootstrap_generated_data_providers.tmpl
var BootstrapGeneratedDataProviders string

//go:embed bootstrap_hooks_ext.tmpl
var BootstrapHooksExt string

//go:embed bootstrap_ent_client.tmpl
var BootstrapEntClient string

//go:embed bootstrap_ent_client_ext.tmpl
var BootstrapEntClientExt string

//go:embed module_entry.tmpl
var ModuleEntry string

//go:embed module_shared_ext.tmpl
var ModuleSharedExt string

//go:embed frontend_view_meta.tmpl
var FrontendViewMeta string

//go:embed frontend_view_config.tmpl
var FrontendViewConfig string

//go:embed frontend_page_i18n_zh.tmpl
var FrontendPageI18nZH string

//go:embed frontend_page_i18n_en.tmpl
var FrontendPageI18nEN string
