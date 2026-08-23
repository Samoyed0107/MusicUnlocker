[CmdletBinding()]
param(
    [switch]$RequireNSIS,
    [switch]$VerifyWailsOutputOnly,
    [string[]]$WailsOutput
)

$ErrorActionPreference = "Stop"
$PSNativeCommandUseErrorActionPreference = $true
Set-StrictMode -Version Latest

function Get-NormalizedLines {
    param([AllowNull()][string[]]$Output)

    @($Output | ForEach-Object { "$_".Trim() } | Where-Object { $_.Length -gt 0 })
}

function Assert-SingleExactLine {
    param(
        [Parameter(Mandatory)][string]$Tool,
        [Parameter(Mandatory)][string[]]$Output,
        [Parameter(Mandatory)][string]$Expected
    )

    $lines = @(Get-NormalizedLines $Output)
    if ($lines.Count -ne 1 -or $lines[0] -cne $Expected) {
        throw "Unexpected $Tool version. Expected exactly '$Expected'; received: [$($lines -join ' | ')]"
    }
}

function Assert-WailsVersionOutput {
    param([Parameter(Mandatory)][string[]]$Output)

    $versionLines = @(Get-NormalizedLines $Output | Where-Object { $_ -match '^v\d+\.\d+\.\d+$' })
    if ($versionLines.Count -ne 1 -or $versionLines[0] -cne "v2.15.0") {
        throw "Unexpected Wails version. Expected exactly one standalone 'v2.15.0' line; received version lines: [$($versionLines -join ' | ')]"
    }
}

if ($VerifyWailsOutputOnly) {
    if (-not $PSBoundParameters.ContainsKey("WailsOutput")) {
        throw "-VerifyWailsOutputOnly requires -WailsOutput"
    }
    Assert-WailsVersionOutput $WailsOutput
    Write-Output "WAILS_VERSION_OUTPUT_OK"
    exit 0
}

$goVersion = @(& go env GOVERSION)
$nodeVersion = @(& node --version)
$npmVersion = @(& npm --version)
$wailsVersion = @(& wails version)

$goVersion | ForEach-Object { Write-Output $_ }
$nodeVersion | ForEach-Object { Write-Output $_ }
$npmVersion | ForEach-Object { Write-Output $_ }
$wailsVersion | ForEach-Object { Write-Output $_ }

Assert-SingleExactLine -Tool "Go" -Output $goVersion -Expected "go1.26.5"
Assert-SingleExactLine -Tool "Node.js" -Output $nodeVersion -Expected "v24.18.0"
Assert-SingleExactLine -Tool "npm" -Output $npmVersion -Expected "11.16.0"
Assert-WailsVersionOutput $wailsVersion

if ($RequireNSIS) {
    $nsisVersion = @(& makensis /VERSION)
    $nsisVersion | ForEach-Object { Write-Output $_ }
    Assert-SingleExactLine -Tool "NSIS" -Output $nsisVersion -Expected "v3.12"
}

Write-Output "WINDOWS_TOOLCHAIN_OK"
