param(
    [string]$WorkspaceRoot = "D:\GoProjects\XAdmin",
    [string]$ProjectName = "",
    [string]$Module = "",
    [string]$AppName = "",
    [string]$ServiceName = "admin",
    [string]$TypeScriptRoot = "",
    [switch]$SkipDryRun,
    [switch]$SkipGoGetUpdateAll,
    [switch]$SmokeTest
)

$ErrorActionPreference = "Stop"

function Test-InteractiveSession {
    try {
        return -not [Console]::IsInputRedirected
    } catch {
        return $true
    }
}

function Resolve-InputValue {
    param(
        [string]$Name,
        [string]$Value,
        [string]$DefaultValue = "",
        [string]$Hint = "",
        [switch]$Required
    )

    if (-not [string]::IsNullOrWhiteSpace($Value)) {
        return $Value.Trim()
    }

    if (Test-InteractiveSession) {
        if (-not [string]::IsNullOrWhiteSpace($Hint)) {
            Write-Host "${Name}: $Hint"
        }

        $prompt = if ([string]::IsNullOrWhiteSpace($DefaultValue)) {
            $Name
        } else {
            "$Name [$DefaultValue]"
        }

        $inputValue = Read-Host $prompt
        if ([string]::IsNullOrWhiteSpace($inputValue)) {
            $inputValue = $DefaultValue
        }

        if (-not [string]::IsNullOrWhiteSpace($inputValue)) {
            return $inputValue.Trim()
        }
    } elseif (-not [string]::IsNullOrWhiteSpace($DefaultValue)) {
        return $DefaultValue.Trim()
    }

    if ($Required) {
        throw "$Name is required. Pass -$Name explicitly or run the script in an interactive terminal."
    }

    return ""
}

$ProjectName = Resolve-InputValue `
    -Name "ProjectName" `
    -Value $ProjectName `
    -Hint "target repo directory name, for example qadmin" `
    -Required

$Module = Resolve-InputValue `
    -Name "Module" `
    -Value $Module `
    -DefaultValue $ProjectName `
    -Hint "Go module name, usually the same as ProjectName"

$AppName = Resolve-InputValue `
    -Name "AppName" `
    -Value $AppName `
    -DefaultValue $ProjectName `
    -Hint "human-facing app name used by bootstrap metadata, for example QAdmin"

$XkitRoot = Join-Path $WorkspaceRoot "xkit"
$TemplateRoot = Join-Path $WorkspaceRoot "xkit-template"
$ExamplesRoot = Join-Path $XkitRoot "examples"
$SourceRoot = Join-Path $ExamplesRoot "admin"
$CanonicalConfigPath = Join-Path $SourceRoot "admin-config\$ServiceName.yaml"
$ProjectRoot = Join-Path $WorkspaceRoot $ProjectName
$TypeScriptRoot = Resolve-InputValue `
    -Name "TypeScriptRoot" `
    -Value $TypeScriptRoot `
    -DefaultValue (Join-Path $ProjectRoot ".generated-ui") `
    -Hint "generated TypeScript output directory"
$ConfigPath = Join-Path $SourceRoot "$ProjectName-config\$ServiceName.yaml"

Write-Host "WorkspaceRoot:  $WorkspaceRoot"
Write-Host "ExamplesRoot:   $ExamplesRoot"
Write-Host "SourceRoot:     $SourceRoot"
Write-Host "ProjectName:    $ProjectName"
Write-Host "Module:         $Module"
Write-Host "ProjectRoot:    $ProjectRoot"
Write-Host "TypeScriptRoot: $TypeScriptRoot"
Write-Host "ConfigPath:     $ConfigPath"
Write-Host "CanonicalConfig:$CanonicalConfigPath"

function Invoke-Step {
    param(
        [string]$Name,
        [scriptblock]$Action
    )

    Write-Host ""
    Write-Host "==> $Name"
    & $Action
}

function Sync-CanonicalConfig {
    param(
        [string]$CanonicalPath,
        [string]$TargetPath,
        [string]$TargetModule
    )

    if (-not (Test-Path -LiteralPath $CanonicalPath)) {
        return
    }

    $content = Get-Content -LiteralPath $CanonicalPath -Raw -Encoding utf8
    $content = [System.Text.RegularExpressions.Regex]::Replace(
        $content,
        '(?m)^module:\s*.+$',
        "module: $TargetModule"
    )
    $content = $content.Replace("admin/api/gen/", "$TargetModule/api/gen/")

    $targetDir = Split-Path -Parent $TargetPath
    if (-not (Test-Path -LiteralPath $targetDir)) {
        New-Item -ItemType Directory -Path $targetDir -Force | Out-Null
    }
    [System.IO.File]::WriteAllText($TargetPath, $content, [System.Text.UTF8Encoding]::new($false))
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
        $args = @(
            "run", "./cmd/xkit", "init", "template", $TemplateRoot,
            "--project", $ProjectRoot,
            "--module", $Module,
            "--app-name", $AppName,
            "--command-name", $ProjectName,
            "--service-name", $ServiceName
        )
        if ($SkipGoGetUpdateAll) {
            $args += "--skip-go-get-update-all"
        }
        & go @args
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
            --typescript-project $TypeScriptRoot `
            --force
    }

    Invoke-Step "Apply canonical example config" {
        Sync-CanonicalConfig `
            -CanonicalPath $CanonicalConfigPath `
            -TargetPath $ConfigPath `
            -TargetModule $Module
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
