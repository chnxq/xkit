# xkit

Service code generation lives here.

Current commands:

```bash
xkit gen service <service> [--project <path>] [--config <path>] [--domain <name>] [--dry-run]
xkit gen repo <service> [--project <path>] [--config <path>] [--domain <name>] [--dry-run]
xkit gen register <service> [--project <path>] [--config <path>] [--domain <name>] [--dry-run]
xkit gen wire <service> [--project <path>] [--config <path>] [--domain <name>] [--dry-run]
xkit gen all <service> [--project <path>] [--config <path>] [--domain <name>] [--dry-run]
```

Example against the current XAdmin workspace:

```bash
go run ./cmd/xkit gen all admin --project ../xadmin-web --config examples/xadmin-web/admin.yaml --dry-run
```

The generator currently supports service stubs, repository CRUD scaffolds,
transport registration files, and Wire provider sets. Generated files include an
`xkit` version and generation timestamp in the header.

