# Repo List Pre-Filter Hook

## Goal

Provide a stable handwritten hook before generated repo list filters run, so
relation-based filtering can live in `*_repo_ext.go` instead of requiring manual
edits to `*_repo.gen.go`.

This is meant for cases like:

- filtering `user` by `org_unit_id`
- filtering `user` by `position_id`
- filtering `user` by `role_id`
- other list-time relation or aggregate predicates that are not direct entity fields

## Generated Contract

For generated repo `List(...)`, xkit now emits an optional hook with this shape:

```go
<repoStructName>CustomList(context.Context, *ent.<Entity>Query, *paginationv1.PagingRequest) (*paginationv1.PagingRequest, error)
```

Example:

```go
userRepoCustomList(ctx context.Context, builder *ent.UserQuery, req *paginationv1.PagingRequest) (*paginationv1.PagingRequest, error)
```

If implemented in `*_repo_ext.go`, the generated `List(...)` method will:

1. build the base `ent.Query`
2. apply tenant scope guards first
3. call the handwritten list pre-filter hook
4. run generated direct-field filters via `applyGeneratedFilters(...)`
5. continue with sorting, paging, count, DTO hydration

## Design Constraints

The hook is intentionally narrow.

It may:

- add extra `builder.Where(...)` predicates
- translate relation filters into `IDIn(...)`
- clone and trim `PagingRequest.filter_expr.conditions`
- preserve non-relation conditions for generated filtering

It should not:

- reimplement generated direct-field filtering
- reimplement sorting or paging
- mutate unrelated request semantics
- bypass tenant guards

## Safety Rule

The hook must preserve all non-custom conditions.

Typical pattern:

1. inspect `req.GetFilterExpr().GetConditions()`
2. consume only the relation conditions you own
3. keep the remaining conditions
4. return a cloned `PagingRequest` whose `filter_expr.conditions` only removes the consumed relation conditions

This guarantees existing generated filters for normal fields like `username`,
`status`, `created_at`, etc. continue to work unchanged.

## Why This Exists

Without this hook, projects drift into one of two bad states:

1. manual edits inside `*_repo.gen.go`
2. pushing relation-specific filtering into frontend or service layers

This hook keeps the filtering responsibility in repo layer, while preserving
generator overwrite safety.
