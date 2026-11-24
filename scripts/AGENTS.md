# Scripts instructions

Inherit ../AGENTS.md. This module contains PowerShell 7 development tooling.

## Code style

- Use four spaces and no tabs. Follow PSScriptAnalyzer and the root PSScriptAnalyzerSettings.psd1.
- Use approved Verb-Noun function names, descriptive variables, explicit parameters and structured objects rather than parsing presentation text.
- Prefer current PowerShell 7 syntax when it simplifies code; use native cmdlets for filesystem work and -LiteralPath for paths that may contain special characters.
- Fail on errors. Check LASTEXITCODE immediately after native commands; ErrorActionPreference alone does not handle their failures.
- Resolve repository paths from PSScriptRoot. Restore location changes with try/finally. Keep caches inside .cache and avoid reserved/system variable names.
- Keep validation deterministic and report failures. Do not silently skip missing tools or failed checks. Do not add destructive cleanup or dependency installation to the regular lint command.

## Commands

From the repository root:

```powershell
Save-Module -Name PSScriptAnalyzer -RequiredVersion 1.24.0 -Path .cache/psmodules -Force
Import-Module ./.cache/psmodules/PSScriptAnalyzer/1.24.0/PSScriptAnalyzer.psd1
Invoke-ScriptAnalyzer -Path scripts -Recurse -Settings ./PSScriptAnalyzerSettings.psd1
pwsh -File scripts/lint.ps1 -Fix
pwsh -File scripts/lint.ps1
```

Run the full validation after every script or instruction change. The nginx check uses an isolated container with an api host mapping, so it does not require the running application.
