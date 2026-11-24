<#
.SYNOPSIS
Runs repository lint, tests, build and container configuration validation.
.PARAMETER Fix
Formats Go and Prettier-managed files before validation. Does not install tools.
.NOTES
Requires the pinned PSScriptAnalyzer module, installed npm dependencies and Docker.
Starts isolated Postgres for all Go tests; no external test database is required.
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
    # Decode gofmt explicitly: Windows PowerShell otherwise uses the console code page.
    $formatter = New-Object System.Diagnostics.Process
    $formatter.StartInfo.FileName = (Get-Command gofmt -ErrorAction Stop).Source
    $formatter.StartInfo.Arguments = '"' + $file.FullName + '"'
    $formatter.StartInfo.UseShellExecute = $false
    $formatter.StartInfo.CreateNoWindow = $true
    $formatter.StartInfo.RedirectStandardOutput = $true
    $formatter.StartInfo.StandardOutputEncoding = [Text.Encoding]::UTF8
    try {
        [void]$formatter.Start()
        $formatted = $formatter.StandardOutput.ReadToEnd()
        $formatter.WaitForExit()
        if ($formatter.ExitCode -ne 0) { throw "gofmt failed for $($file.Name)" }
    } finally { $formatter.Dispose() }
    $formatted = $formatted.Replace("`r`n", "`n").Replace("`t", '    ')
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
# Use a unique disposable database, never the application's database or volume.
$testContainerName = 'sportsbook-check-' + [guid]::NewGuid().ToString('N')
$previousTestDatabaseUrl = $env:TEST_DATABASE_URL
$testContainerId = $null
try {
    $testContainerId = & docker run --detach --rm --name $testContainerName --env POSTGRES_PASSWORD=integration-test --publish 127.0.0.1::5432 postgres:17-alpine
    Assert-NativeSuccess
    $ready = $false
    for ($attempt = 0; $attempt -lt 60; $attempt++) {
        & docker exec $testContainerId pg_isready --host 127.0.0.1 --username postgres *> $null
        if ($LASTEXITCODE -eq 0) {
            $ready = $true
            break
        }
        Start-Sleep -Milliseconds 500
    }
    if (-not $ready) {
        & docker logs $testContainerId
        Assert-NativeSuccess
        throw 'Test Postgres did not become ready within 30 seconds'
    }
    $containerJson = & docker inspect $testContainerId
    Assert-NativeSuccess
    $containerInfo = $containerJson | ConvertFrom-Json
    $testPort = $containerInfo[0].NetworkSettings.Ports.'5432/tcp'[0].HostPort
    if (-not $testPort) { throw 'Test Postgres port mapping is missing' }
    $env:TEST_DATABASE_URL = "host=127.0.0.1 port=$testPort user=postgres password=integration-test dbname=postgres sslmode=disable"
    Push-Location backend
    try {
        & go vet ./...
        Assert-NativeSuccess
        # Disable caching and show the integration test actually running.
        & go test -count=1 -v ./...
        Assert-NativeSuccess
    } finally { Pop-Location }
} finally {
    $env:TEST_DATABASE_URL = $previousTestDatabaseUrl
    if ($testContainerId) {
        & docker rm --force --volumes $testContainerId
        Assert-NativeSuccess
    }
}
Push-Location frontend
try {
    & npm test
    Assert-NativeSuccess
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


