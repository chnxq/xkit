# Phase 3 Feature Gap Review

Use this reference after Phase 2 when the target project already builds, but you still need to understand what functionally differs from the chosen reference project.

This phase is not a raw diff pass. It is a feature review.

## Goal

Turn repository differences into a prioritized feature-gap summary.

The main question is:

- "What can the reference project do that the target project still cannot do, or does differently?"

Not:

- "Which files are still different?"

## Input

Compare:

- target project root
- `<ReferenceProjectRoot>`

Assume:

- Phase 1 generation is complete
- Phase 2 repair is complete enough for the target to build and test

## Review method

### 1. Start from functional surfaces

Review by feature surface first, for example:

- login, logout, refresh token, captcha
- current user, profile, portal, menu
- role, permission, org unit, dictionary, tenant
- audit logs, permission logs, analytics dashboard
- internal messages, announcements, SSE
- default data, SQL bootstrap, operational scripts

Do not start by traversing directories mechanically unless you already know which feature area the directory serves.

### 2. For each feature area, answer four questions

For every major area:

1. Does the target already have equivalent functionality?
2. If not, is the gap caused by missing generated inputs, missing preserved hand-written code, missing config/data, or intentional design divergence?
3. Does the gap matter now?
4. Which files or modules likely own the gap?

### 3. Classify each gap

Use these categories:

- blocking: prevents the target from being used as intended
- important: does not block startup, but leaves a clearly incomplete business capability
- optional: useful parity work, but not required for the current milestone
- intentional: should stay different because it belongs to template/generator boundaries or target-specific design

### 4. Look for common false positives

Not every difference is a bug or missing feature.

Common acceptable differences:

- module path differences
- generated file layout differences caused by newer generator output
- template-owned startup scaffolding that is intentionally generic
- reference-only demo data or environment-specific operational scripts

## Required output file

Write the result into the target repo at:

- `<target>/docs/xkit-helper-feature-gap-review.md`

If the `docs` directory does not exist, create it.

You may add a short pointer from `<target>/readme.md`, but the main content should live in the dedicated review file.

## Recommended output format

For each functional area, record:

- feature area
- target status
- reference status
- gap classification
- likely owner files
- recommended next action

Also include a short header section with:

- review date
- target project path
- resolved `ReferenceSource`
- reference revision or commit when available
- current verification baseline such as `go test ./...`

## Suggested feature checklist

- authentication and token flow
- captcha and login protection
- viewer auth and permission middleware
- menu, portal, and profile APIs
- user, role, permission, org-unit, tenant management
- audit log families and DB logging behavior
- analytics metrics and dashboard data
- internal messaging, recipient state, revoke rules, SSE push
- default data initialization
- SQL bootstrap/demo data

## Final success condition

Phase 3 is complete when:

- the remaining differences are understandable by feature, not just by file
- the next iteration can be planned from the summary
- the target repo contains `docs/xkit-helper-feature-gap-review.md` in Chinese
