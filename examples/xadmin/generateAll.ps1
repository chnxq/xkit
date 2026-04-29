param(
    [string]$WorkspaceRoot = "D:\GoProjects\XAdmin",
    [string]$ProjectName = "xadmin",
    [string]$Module = "",
    [string]$AppName = "XAdmin",
    [string]$ServiceName = "admin",
    [string]$TypeScriptRoot = "",
    [switch]$SkipDryRun,
    [switch]$SmokeTest
)

$ErrorActionPreference = "Stop"

$XkitRoot = Join-Path $WorkspaceRoot "xkit"
$TemplateRoot = Join-Path $WorkspaceRoot "xkit-template"
$SourceRoot = Join-Path $XkitRoot "examples\xadmin"
if ([string]::IsNullOrWhiteSpace($Module)) {
    $Module = $ProjectName
}
$ProjectRoot = Join-Path $WorkspaceRoot $ProjectName
if ([string]::IsNullOrWhiteSpace($TypeScriptRoot)) {
    $TypeScriptRoot = Join-Path $WorkspaceRoot "$ProjectName-ui"
}
$ConfigPath = Join-Path $SourceRoot "$ProjectName-config\$ServiceName.yaml"

Write-Host "WorkspaceRoot: $WorkspaceRoot"
Write-Host "ProjectName:   $ProjectName"
Write-Host "Module:        $Module"
Write-Host "ProjectRoot:   $ProjectRoot"
Write-Host "TypeScriptRoot:$TypeScriptRoot"
Write-Host "ConfigPath:    $ConfigPath"

function Invoke-Step {
    param(
        [string]$Name,
        [scriptblock]$Action
    )

    Write-Host ""
    Write-Host "==> $Name"
    & $Action
}

Push-Location $XkitRoot
try {
    if (-not $SkipDryRun) {
        Invoke-Step "Preview bootstrap template copy" {
            go run ./cmd/xkit init template $TemplateRoot `
                --project $ProjectRoot `
                --module $Module `
                --app-name $AppName `
                --command-name $ProjectName `
                --service-name $ServiceName `
                --dry-run
        }
    }

    Invoke-Step "Copy bootstrap template" {
        go run ./cmd/xkit init template $TemplateRoot `
            --project $ProjectRoot `
            --module $Module `
            --app-name $AppName `
            --command-name $ProjectName `
            --service-name $ServiceName `
            --skip-go-get-update-all
    }

    if (-not $SkipDryRun) {
        Invoke-Step "Preview source import and config generation" {
            go run ./cmd/xkit init source $SourceRoot `
                --project $ProjectRoot `
                --service $ServiceName `
                --typescript-project $TypeScriptRoot `
                --dry-run
        }
    }

    Invoke-Step "Import source and generate config" {
        go run ./cmd/xkit init source $SourceRoot `
            --project $ProjectRoot `
            --service $ServiceName `
            --typescript-project $TypeScriptRoot
    }

    Invoke-Step "Generate API Go code" {
        Push-Location (Join-Path $ProjectRoot "api")
        try {
            buf generate --template buf.gen.yaml
        } finally {
            Pop-Location
        }
    }

    Invoke-Step "Generate OpenAPI document" {
        Push-Location (Join-Path $ProjectRoot "api")
        try {
            buf generate --template buf.admin.openapi.gen.yaml
        } finally {
            Pop-Location
        }
    }

    Invoke-Step "Generate TypeScript API code" {
        Push-Location (Join-Path $ProjectRoot "api")
        try {
            buf generate --template buf.vue.admin.typescript.gen.yaml
        } finally {
            Pop-Location
        }
    }

    Invoke-Step "Prepare dependencies before Ent generation" {
        Push-Location $ProjectRoot
        try {
            go mod tidy
        } finally {
            Pop-Location
        }
    }

    Invoke-Step "Generate Ent code" {
        Push-Location $ProjectRoot
        try {
            go run -mod=mod entgo.io/ent/cmd/ent generate ./internal/data/ent/schema --feature privacy,sql/upsert,sql/versioned-migration
        } finally {
            Pop-Location
        }
    }

    if (-not $SkipDryRun) {
        Invoke-Step "Preview xkit dynamic code generation" {
            go run ./cmd/xkit gen all $ServiceName `
                --project $ProjectRoot `
                --config $ConfigPath `
                --dry-run
        }
    }

    Invoke-Step "Generate xkit dynamic code" {
        go run ./cmd/xkit gen all $ServiceName `
            --project $ProjectRoot `
            --config $ConfigPath
    }

    Invoke-Step "Final dependency update" {
        Push-Location $ProjectRoot
        try {
            go get -u all
            go mod tidy
        } finally {
            Pop-Location
        }
    }

    Invoke-Step "Run tests" {
        Push-Location $ProjectRoot
        try {
            go test ./...
        } finally {
            Pop-Location
        }
    }

    if ($SmokeTest) {
        Invoke-Step "Run server smoke test" {
            Push-Location $ProjectRoot
            try {
                go run ./cmd/server server -config_path ./configs
            } finally {
                Pop-Location
            }
        }
    }
} finally {
    Pop-Location
}
