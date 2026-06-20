param(
    [string]$WorkspaceRoot = "D:\GoProjects\XAdmin",
    [string]$ModuleName = "",
    [string]$ServiceName = "xdev",
    [string]$HostProject = "",
    [string]$TypeScriptRoot = "",
    [string]$SourceRoot = "",
    [string]$ModuleRoot = "",
    [string]$CanonicalConfigPath = "",
    [string]$ConfigPath = "",
    [switch]$SkipDryRun,
    [switch]$SkipTypeScript
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
        [string]$XkitRoot,
        [string]$SourceRoot,
        [string]$HostProject,
        [string]$ModuleName,
        [string]$ServiceName,
        [string]$ModuleRoot,
        [string]$ConfigPath,
        [string]$CanonicalConfigPath,
        [string]$TypeScriptRoot,
        [string]$FrontendApiRoot,
        [bool]$InteractiveMode,
        [bool]$SkipDryRun,
        [bool]$SkipTypeScript
    )

    Write-Host ""
    Write-Host "================ xkit generateModule summary ================"
    Write-Host "WorkspaceRoot:      $WorkspaceRoot"
    Write-Host "XkitRoot:           $XkitRoot"
    Write-Host "SourceRoot:         $SourceRoot"
    Write-Host "HostProject:        $HostProject"
    Write-Host "ModuleName:         $ModuleName"
    Write-Host "ServiceName:        $ServiceName"
    Write-Host "ModuleRoot:         $ModuleRoot"
    Write-Host "TargetConfig:       $ConfigPath"
    Write-Host "CanonicalConfig:    $CanonicalConfigPath"
    Write-Host "TypeScriptRoot:     $TypeScriptRoot"
    Write-Host "FrontendApiRoot:    $FrontendApiRoot"
    Write-Host "InteractiveMode:    $InteractiveMode"
    Write-Host "SkipDryRun:         $SkipDryRun"
    Write-Host "SkipTypeScript:     $SkipTypeScript"
    Write-Host ""
    Write-Host "Planned stages:"
    Write-Host "  1. init module"
    Write-Host "  2. apply canonical module config"
    Write-Host "  3. generate Go/OpenAPI/TypeScript API code"
    Write-Host "  4. generate Ent code"
    Write-Host "  5. run xkit gen module (including frontend meta generation)"
    Write-Host "============================================================="
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
        throw "Canonical config not found: $CanonicalPath"
    }

    $content = Get-Content -LiteralPath $CanonicalPath -Raw -Encoding utf8
    $content = [System.Text.RegularExpressions.Regex]::Replace(
        $content,
        '(?m)^module:\s*.+$',
        "module: $TargetModule"
    )
    $content = $content.Replace("$ModuleName/api/gen/", "$TargetModule/api/gen/")

    $targetDir = Split-Path -Parent $TargetPath
    if (-not (Test-Path -LiteralPath $targetDir)) {
        New-Item -ItemType Directory -Path $targetDir -Force | Out-Null
    }

    [System.IO.File]::WriteAllText($TargetPath, $content, [System.Text.UTF8Encoding]::new($false))
}

$InteractiveMode = Test-InteractiveSession

$ModuleName = Resolve-InputValue `
    -Name "ModuleName" `
    -Value $ModuleName `
    -DefaultValue "xdev" `
    -Hint "module directory name under admin/modules, for example xdev" `
    -Required

$XkitRoot = Join-Path $WorkspaceRoot "xkit"
$SourceRoot = Resolve-InputValue `
    -Name "SourceRoot" `
    -Value $SourceRoot `
    -DefaultValue (Join-Path $XkitRoot "examples\$ModuleName") `
    -Hint "module example source root"

$HostProject = Resolve-InputValue `
    -Name "HostProject" `
    -Value $HostProject `
    -DefaultValue (Join-Path $WorkspaceRoot "admin") `
    -Hint "target host admin project root"

$ModuleRoot = Resolve-InputValue `
    -Name "ModuleRoot" `
    -Value $ModuleRoot `
    -DefaultValue (Join-Path $HostProject "modules\$ModuleName") `
    -Hint "target module source root under host project"

$TypeScriptRoot = Resolve-InputValue `
    -Name "TypeScriptRoot" `
    -Value $TypeScriptRoot `
    -DefaultValue (Join-Path $WorkspaceRoot "admin-ui") `
    -Hint "target frontend project root"

$CanonicalConfigPath = Resolve-InputValue `
    -Name "CanonicalConfigPath" `
    -Value $CanonicalConfigPath `
    -DefaultValue (Join-Path $SourceRoot "$ModuleName-config\$ModuleName.yaml") `
    -Hint "canonical config source copied onto TargetConfig after init module"

$ConfigPath = Resolve-InputValue `
    -Name "ConfigPath" `
    -Value $ConfigPath `
    -DefaultValue (Join-Path $SourceRoot "$ModuleName-target-config\$ModuleName.yaml") `
    -Hint "effective module generation config; future generation should always use this file"

$FrontendApiRoot = Join-Path $TypeScriptRoot "apps\web-antd\src\api\generated\$ServiceName"

Show-ExecutionSummary `
    -WorkspaceRoot $WorkspaceRoot `
    -XkitRoot $XkitRoot `
    -SourceRoot $SourceRoot `
    -HostProject $HostProject `
    -ModuleName $ModuleName `
    -ServiceName $ServiceName `
    -ModuleRoot $ModuleRoot `
    -ConfigPath $ConfigPath `
    -CanonicalConfigPath $CanonicalConfigPath `
    -TypeScriptRoot $TypeScriptRoot `
    -FrontendApiRoot $FrontendApiRoot `
    -InteractiveMode $InteractiveMode `
    -SkipDryRun ([bool]$SkipDryRun) `
    -SkipTypeScript ([bool]$SkipTypeScript)

Confirm-OrAbort -InteractiveMode $InteractiveMode

Push-Location $XkitRoot
try {
    if (-not $SkipDryRun) {
        Invoke-Step "Preview module source import" {
            go run .\cmd\xkit init module $SourceRoot `
                --project $HostProject `
                --module-name $ModuleName `
                --module-root $ModuleRoot `
                --service $ServiceName `
                --config $ConfigPath `
                --typescript-project $TypeScriptRoot `
                --force `
                --dry-run
        }
    }

    Invoke-Step "Import module source" {
        go run .\cmd\xkit init module $SourceRoot `
            --project $HostProject `
            --module-name $ModuleName `
            --module-root $ModuleRoot `
            --service $ServiceName `
            --config $ConfigPath `
            --typescript-project $TypeScriptRoot `
            --force
    }

    Invoke-Step "Apply canonical module config" {
        Sync-CanonicalConfig `
            -CanonicalPath $CanonicalConfigPath `
            -TargetPath $ConfigPath `
            -TargetModule ("admin/modules/" + $ModuleName)
    }

    Invoke-Step "Generate API Go code" {
        Push-Location (Join-Path $ModuleRoot "api")
        try {
            buf generate --template buf.gen.yaml
        } finally {
            Pop-Location
        }
    }

    Invoke-Step "Generate OpenAPI document" {
        Push-Location (Join-Path $ModuleRoot "api")
        try {
            buf generate --template "buf.$ModuleName.openapi.gen.yaml"
        } finally {
            Pop-Location
        }
    }

    if (-not $SkipTypeScript) {
        Invoke-Step "Generate TypeScript API code" {
            Push-Location (Join-Path $ModuleRoot "api")
            try {
                buf generate --template "buf.vue.$ModuleName.typescript.gen.yaml"
            } finally {
                Pop-Location
            }
        }
    }

    Invoke-Step "Prepare dependencies before Ent generation" {
        Push-Location $HostProject
        try {
            go mod tidy
        } finally {
            Pop-Location
        }
    }

    Invoke-Step "Generate Ent code" {
        Push-Location $ModuleRoot
        try {
            go run -mod=mod entgo.io/ent/cmd/ent generate .\data\schema --feature privacy,sql/upsert,sql/versioned-migration,sql/modifier
        } finally {
            Pop-Location
        }
    }

    if (-not $SkipDryRun) {
        Invoke-Step "Preview xkit gen module" {
            go run .\cmd\xkit gen module $ModuleName $ServiceName `
                --project $HostProject `
                --module-root $ModuleRoot `
                --config $ConfigPath `
                --typescript-project $TypeScriptRoot `
                --dry-run
        }
    }

    Invoke-Step "Generate module backend code and frontend meta" {
        go run .\cmd\xkit gen module $ModuleName $ServiceName `
            --project $HostProject `
            --module-root $ModuleRoot `
            --config $ConfigPath `
            --typescript-project $TypeScriptRoot
    }
}
finally {
    Pop-Location
}
