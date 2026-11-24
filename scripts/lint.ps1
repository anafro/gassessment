<#
.SYNOPSIS
Runs repository lint, tests, build and container configuration validation.
.PARAMETER Fix
Formats Go and Prettier-managed files before validation. Does not install tools.
.NOTES
Requires the pinned PSScriptAnalyzer module, installed npm dependencies and Docker.
Postgres integration tests run only when TEST_DATABASE_URL is set.
#>
param([switch]$Fix)
$ErrorActionPreference = 'Stop'
Set-Location (Split-Path $PSScriptRoot -Parent)
Import-Module ./.cache/psmodules/PSScriptAnalyzer/1.24.0/PSScriptAnalyzer.psd1 -ErrorAction Stop
$issues = @(Invoke-ScriptAnalyzer -Path scripts -Recurse -Settings ./PSScriptAnalyzerSettings.psd1)
if ($issues.Count -gt 0) {
    $issues | Format-Table
    throw 'PowerShell lint failed'
}
# Native program failures do not automatically obey ErrorActionPreference.
function Assert-NativeSuccess {
    if ($LASTEXITCODE -ne 0) { throw "Check failed (exit $LASTEXITCODE)" }
}
$env:GOCACHE = Join-Path (Get-Location) '.cache/go'
$env:GOMODCACHE = Join-Path (Get-Location) '.cache/gomod'
foreach ($file in Get-ChildItem backend/*.go) {
    $formatted = (& gofmt $file.FullName) -join "`n"
    Assert-NativeSuccess
    $formatted = $formatted.Replace("`t", '    ') + "`n"
    if ($Fix) {
        [IO.File]::WriteAllText($file.FullName, $formatted)
    } elseif ([IO.File]::ReadAllText($file.FullName).Replace("`r`n", "`n") -cne $formatted) {
        throw "Go formatting differs: $($file.Name). Run scripts/lint.ps1 -Fix"
    }
}
if ($Fix) {
    & npm run format
    Assert-NativeSuccess
}
& npm run lint
Assert-NativeSuccess
Push-Location backend
try {
    & go vet ./...
    Assert-NativeSuccess
    & go test ./...
    Assert-NativeSuccess
} finally { Pop-Location }
Push-Location frontend
try {
    & npm run build
    Assert-NativeSuccess
} finally { Pop-Location }
& docker compose config --quiet
Assert-NativeSuccess
foreach ($directory in @('backend', 'frontend')) {
    Get-Content "$directory/Dockerfile" -Raw | docker run --rm -i hadolint/hadolint:v2.12.0 hadolint --ignore DL3018 --ignore DL3016 --ignore DL3008 -
    Assert-NativeSuccess
}
& docker run --rm --add-host api:127.0.0.1 --mount "type=bind,source=$(Join-Path (Get-Location) 'frontend/nginx.conf'),target=/etc/nginx/conf.d/default.conf,readonly" nginx:alpine nginx -t
Assert-NativeSuccess


