analyze-admin: Candidate generated-file list

Purpose
- Catalog generated outputs in admin for extracting diffs and identifying template candidates.

Summary
- Focus areas: admin/internal/service (*.gen.go + *_ext.go), admin/internal/data/repo (*.gen.go), admin/internal/server/bootstrap registrars, admin/api/gen (protobuf .pb.go / _grpc.pb.go).

1) admin/internal/service (generated + extension pairs)
- api_audit_log_service.gen.go / api_audit_log_service_ext.go
- api_service.gen.go / api_service_ext.go
- data_access_audit_log_service.gen.go / data_access_audit_log_service_ext.go
- dict_entry_service.gen.go / dict_entry_service_ext.go
- dict_type_service.gen.go / dict_type_service_ext.go
- file_service.gen.go / file_service_ext.go
- internal_message_category_service.gen.go / internal_message_category_service_ext.go
- internal_message_recipient_service.gen.go / internal_message_recipient_service_ext.go
- internal_message_service.gen.go / internal_message_service_ext.go
- language_service.gen.go / language_service_ext.go
- login_audit_log_service.gen.go / login_audit_log_service_ext.go
- login_policy_service.gen.go / login_policy_service_ext.go
- menu_service.gen.go / menu_service_ext.go
- operation_audit_log_service.gen.go / operation_audit_log_service_ext.go
- org_unit_service.gen.go / org_unit_service_ext.go
- permission_audit_log_service.gen.go / permission_audit_log_service_ext.go
- permission_group_service.gen.go / permission_group_service_ext.go
- permission_service.gen.go / permission_service_ext.go
- policy_evaluation_log_service.gen.go / policy_evaluation_log_service_ext.go
- position_service.gen.go / position_service_ext.go
- role_service.gen.go / role_service_ext.go
- task_service.gen.go / task_service_ext.go
- tenant_service.gen.go / tenant_service_ext.go
- user_service.gen.go / user_service_ext.go

2) admin/internal/data/repo (examples)
- admin/internal/data/repo/user_repo.gen.go
- admin/internal/data/repo/user_credential_repo.gen.go
- admin/internal/data/repo/tenant_repo.gen.go
- admin/internal/data/repo/task_repo.gen.go
- admin/internal/data/repo/role_repo.gen.go
- admin/internal/data/repo/position_repo.gen.go
- admin/internal/data/repo/policy_evaluation_log_repo.gen.go
- admin/internal/data/repo/permission_repo.gen.go
- admin/internal/data/repo/permission_group_repo.gen.go
- admin/internal/data/repo/permission_audit_log_repo.gen.go
- admin/internal/data/repo/org_unit_repo.gen.go
- admin/internal/data/repo/operation_audit_log_repo.gen.go
- admin/internal/data/repo/menu_repo.gen.go
- admin/internal/data/repo/login_policy_repo.gen.go
- admin/internal/data/repo/login_audit_log_repo.gen.go
- admin/internal/data/repo/language_repo.gen.go
- admin/internal/data/repo/internal_message_category_repo.gen.go
- admin/internal/data/repo/file_repo.gen.go
- admin/internal/data/repo/dict_type_repo.gen.go
- admin/internal/data/repo/dict_entry_repo.gen.go
- admin/internal/data/repo/data_access_audit_log_repo.gen.go
- admin/internal/data/repo/api_repo.gen.go

3) server / bootstrap generated files
- admin/internal/bootstrap/generated_servers.gen.go
- admin/internal/server/grpc_register.gen.go
- admin/internal/server/rest_register.gen.go
- admin/internal/data/bootstrap/ent_client.gen.go

4) admin/api/gen (protobuf outputs — large set)
- Many .pb.go and _grpc.pb.go under admin/api/gen/** (authentication, permission, identity, admin, audit, storage, resource, dict, etc.)
- Examples: admin/api/gen/authentication/v1/*.pb.go, admin/api/gen/permission/v1/*.pb.go, admin/api/gen/identity/v1/*.pb.go, admin/api/gen/admin/v1/*.pb.go

Notes & next steps
- This file is a starting point for extracting diffs. Recommended next actions:
  1) If repository has git history: run `git log -p -- admin/internal/**.gen.go admin/api/gen/**` to extract recent diffs (3-6 weeks).
  2) If comparing against xkit-template: generate code from template into a temp dir and run `git diff --no-index` vs current files to obtain candidate patches.
  3) Prioritize files where companion *_ext.go exist — these often contain manual overrides and are key template-candidates.

If you want, extract diffs now (requires git history or template output). Which extraction method to use?
