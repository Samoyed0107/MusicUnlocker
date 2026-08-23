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
    $existingText = & gh release view $Tag --json isDraft,tagName,assets 2>$null
    $releaseExists = $LASTEXITCODE -eq 0

    if ($releaseExists) {
        $existing = $existingText | ConvertFrom-Json
        if (-not $existing.isDraft) {
            throw "Release $Tag already exists and is not a draft; refusing to modify it"
        }
        $expectedNames = @($assets | ForEach-Object { Split-Path -Leaf $_ })
        $unexpectedAssets = @($existing.assets | Where-Object { $_.name -notin $expectedNames })
        if ($unexpectedAssets.Count -gt 0) {
            throw "Draft Release $Tag contains unexpected assets; refusing to delete or replace them: $((@($unexpectedAssets.name) | Sort-Object) -join ', ')"
        }
        & gh release edit $Tag --draft --notes-file $notePath
        if ($LASTEXITCODE -ne 0) { throw "Unable to update Draft Release $Tag notes" }
        & gh release upload $Tag @assets --clobber
        if ($LASTEXITCODE -ne 0) { throw "Unable to upload assets to Draft Release $Tag" }
    } else {
        & gh release create $Tag --draft --verify-tag --notes-file $notePath @assets
        if ($LASTEXITCODE -ne 0) { throw "Unable to create Draft Release $Tag" }
    }
} finally {
    Pop-Location
}
