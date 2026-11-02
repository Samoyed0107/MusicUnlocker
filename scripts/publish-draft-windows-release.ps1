[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$Tag,
    [Parameter(Mandatory)][string]$ArtifactDirectory,
    [Parameter(Mandatory)][string]$ReleaseNote
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$repoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot ".."))
$artifactRoot = [System.IO.Path]::GetFullPath((Join-Path $repoRoot $ArtifactDirectory))
$notePath = [System.IO.Path]::GetFullPath((Join-Path $repoRoot $ReleaseNote))

Push-Location $repoRoot
try {
    & "$PSScriptRoot/verify-windows-release.ps1" -ArtifactDirectory $ArtifactDirectory -Tag $Tag | Out-Host
    if (-not (Get-Command "gh" -ErrorAction SilentlyContinue)) {
        throw "GitHub CLI is required to create or update a Draft Release"
    }

    $assets = @(Get-ChildItem -LiteralPath $artifactRoot -File | Sort-Object Name | ForEach-Object FullName)
    $isRC = $Tag -cmatch '-rc\.([1-9][0-9]*)$'
    $releaseTitle = if ($isRC) { "MusicUnlocker $($Tag.Substring(1).Split('-')[0]) RC $($Matches[1])" } else { "MusicUnlocker $($Tag.Substring(1))" }
    $releaseFlags = @("--draft")
    if ($isRC) { $releaseFlags += "--prerelease" }
    $existingText = & gh release view $Tag --json isDraft,isPrerelease,tagName,assets 2>$null
    $releaseExists = $LASTEXITCODE -eq 0

    if ($releaseExists) {
        $existing = $existingText | ConvertFrom-Json
        if (-not $existing.isDraft -or $existing.tagName -cne $Tag) {
            throw "Release $Tag is not the exact matching draft; refusing to modify it"
        }
        if ([bool]$existing.isPrerelease -ne [bool]$isRC) {
            throw "Release $Tag prerelease flag differs from tag policy; refusing to modify it"
        }
        $expectedNames = @($assets | ForEach-Object { Split-Path -Leaf $_ })
        $unexpectedAssets = @($existing.assets | Where-Object { $_.name -notin $expectedNames })
        if ($unexpectedAssets.Count -gt 0) {
            throw "Draft Release $Tag contains unexpected assets; refusing to delete or replace them: $((@($unexpectedAssets.name) | Sort-Object) -join ', ')"
        }
        & gh release edit $Tag @releaseFlags --title $releaseTitle --notes-file $notePath
        if ($LASTEXITCODE -ne 0) { throw "Unable to update Draft Release $Tag notes" }
        & gh release upload $Tag @assets --clobber
        if ($LASTEXITCODE -ne 0) { throw "Unable to upload assets to Draft Release $Tag" }
    } else {
        & gh release create $Tag @releaseFlags --verify-tag --title $releaseTitle --notes-file $notePath @assets
        if ($LASTEXITCODE -ne 0) { throw "Unable to create Draft Release $Tag" }
    }
    $confirmedText = & gh release view $Tag --json isDraft,isPrerelease,tagName
    if ($LASTEXITCODE -ne 0) { throw "Unable to confirm Draft Release $Tag state" }
    $confirmed = $confirmedText | ConvertFrom-Json
    if (-not $confirmed.isDraft -or $confirmed.tagName -cne $Tag -or [bool]$confirmed.isPrerelease -ne [bool]$isRC) {
        throw "Draft Release $Tag failed final draft/prerelease identity verification"
    }
} finally {
    Pop-Location
}
