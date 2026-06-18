param(
    [string]$WorkspaceRoot = "D:\GoProjects\XAdmin",
    [string]$ProjectName = "",
    [string]$Module = "",
    [string]$AppName = "",
    [string]$ServiceName = "admin",
    [string]$TypeScriptRoot = "",
    [string]$CanonicalConfigPath = "",
    [string]$ConfigPath = "",
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

function Show-ExecutionSummary {
    param(
        [string]$WorkspaceRoot,
        [string]$ProjectName,
        [string]$Module,
        [string]$AppName,
        [string]$ServiceName,
        [string]$ProjectRoot,
        [string]$TypeScriptRoot,
        [string]$ConfigPath,
        [string]$CanonicalConfigPath,
        [bool]$InteractiveMode,
        [bool]$SkipDryRun,
        [bool]$SkipGoGetUpdateAll,
        [bool]$SmokeTest
    )

    Write-Host ""
    Write-Host "================ xkit generateAll summary ================"
    Write-Host "WorkspaceRoot:      $WorkspaceRoot"
    Write-Host "ProjectName:        $ProjectName"
    Write-Host "Module:             $Module"
    Write-Host "AppName:            $AppName"
    Write-Host "ServiceName:        $ServiceName"
    Write-Host "ProjectRoot:        $ProjectRoot"
    Write-Host "TypeScriptRoot:     $TypeScriptRoot"
    Write-Host "TargetConfig:       $ConfigPath"
    Write-Host "CanonicalConfig:    $CanonicalConfigPath"
    Write-Host "InteractiveMode:    $InteractiveMode"
    Write-Host "SkipDryRun:         $SkipDryRun"
    Write-Host "SkipGoGetUpdateAll: $SkipGoGetUpdateAll"
    Write-Host "SmokeTest:          $SmokeTest"
    Write-Host ""
    Write-Host "Planned stages:"
    Write-Host "  1. init template"
    Write-Host "  2. init source"
    Write-Host "  3. apply canonical admin config"
    Write-Host "  4. generate Go/OpenAPI/TypeScript API code"
    Write-Host "  5. generate Ent code"
    Write-Host "  6. run xkit gen all (including frontend meta generation)"
    Write-Host "  7. tidy dependencies and run tests"
    if ($SmokeTest) {
        Write-Host "  8. run server smoke test"
    }
    Write-Host "=========================================================="
}

function Confirm-OrAbort {
    param(
        [bool]$InteractiveMode
    )

    if (-not $InteractiveMode) {
        return
    }

    $answer = Read-Host "Continue with this plan? [Y/n]"
    if ([string]::IsNullOrWhiteSpace($answer)) {
        return
    }

    $normalized = $answer.Trim().ToLowerInvariant()
    if ($normalized -eq "y" -or $normalized -eq "yes") {
        return
    }

    throw "Aborted by user before execution."
}

$InteractiveMode = Test-InteractiveSession

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
$ProjectRoot = Join-Path $WorkspaceRoot $ProjectName
$TypeScriptRoot = Resolve-InputValue `
    -Name "TypeScriptRoot" `
    -Value $TypeScriptRoot `
    -DefaultValue (Join-Path $ProjectRoot ".generated-ui") `
    -Hint "generated TypeScript output directory"
$ConfigDirName = "$ProjectName-config"
if ((Join-Path $SourceRoot $ConfigDirName) -eq (Join-Path $SourceRoot "admin-config")) {
    $ConfigDirName = "$ProjectName-target-config"
}
$DefaultConfigPath = Join-Path $SourceRoot "$ConfigDirName\$ServiceName.yaml"
$CanonicalConfigPath = Resolve-InputValue `
    -Name "CanonicalConfigPath" `
    -Value $CanonicalConfigPath `
    -DefaultValue $DefaultCanonicalConfigPath `
    -Hint "canonical config source copied onto ConfigPath after init source"
$DefaultCanonicalConfigPath = Join-Path $SourceRoot "admin-config\$ServiceName.yaml"
$ConfigPath = Resolve-InputValue `
    -Name "ConfigPath" `
    -Value $ConfigPath `
    -DefaultValue $DefaultConfigPath `
    -Hint "generation config file path used by init source and xkit gen all"

Show-ExecutionSummary `
    -WorkspaceRoot $WorkspaceRoot `
    -ProjectName $ProjectName `
    -Module $Module `
    -AppName $AppName `
    -ServiceName $ServiceName `
    -ProjectRoot $ProjectRoot `
    -TypeScriptRoot $TypeScriptRoot `
    -CanonicalConfigPath $CanonicalConfigPath `
    -ConfigPath $ConfigPath `
    -InteractiveMode $InteractiveMode `
    -SkipDryRun ([bool]$SkipDryRun) `
    -SkipGoGetUpdateAll ([bool]$SkipGoGetUpdateAll) `
    -SmokeTest ([bool]$SmokeTest)

Confirm-OrAbort -InteractiveMode $InteractiveMode

function Invoke-Step {
    param(
        [string]$Name,
        [scriptblock]$Action
    )

    Write-Host ""
    Write-Host "==> $Name"
    $global:LASTEXITCODE = 0
    & $Action
    if ($global:LASTEXITCODE -ne 0) {
        throw "$Name failed with exit code $global:LASTEXITCODE"
    }
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
                --config $ConfigPath `
                --typescript-project $TypeScriptRoot `
                --dry-run
        }
    }

    Invoke-Step "Import source and generate config" {
        go run ./cmd/xkit init source $SourceRoot `
            --project $ProjectRoot `
            --service $ServiceName `
            --config $ConfigPath `
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
            go run -mod=mod entgo.io/ent/cmd/ent generate ./internal/data/ent/schema --feature privacy,sql/upsert,sql/versioned-migration,sql/modifier
        } finally {
            Pop-Location
        }
    }

    if (-not $SkipDryRun) {
        Invoke-Step "Preview xkit dynamic code generation" {
            go run ./cmd/xkit gen all $ServiceName `
                --project $ProjectRoot `
                --config $ConfigPath `
                --typescript-project $TypeScriptRoot `
                --dry-run
        }
    }

    Invoke-Step "Generate xkit dynamic code and frontend meta" {
        go run ./cmd/xkit gen all $ServiceName `
            --project $ProjectRoot `
            --config $ConfigPath `
            --typescript-project $TypeScriptRoot
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
