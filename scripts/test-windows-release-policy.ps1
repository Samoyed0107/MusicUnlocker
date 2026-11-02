[CmdletBinding()]
param()

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$repoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot ".."))
$verify = Join-Path $PSScriptRoot "verify-windows-release.ps1"
$createdNotes = [System.Collections.Generic.List[string]]::new()
$passCount = 0

function Assert-ReleaseContext {
    param(
        [AllowEmptyString()][string]$Tag,
        [bool]$ShouldPass,
        [string]$Note,
        [bool]$Prerelease = $false,
        [string]$ErrorContains = ""
    )
    $outputFile = Join-Path ([System.IO.Path]::GetTempPath()) ("musicunlocker-rc-policy-" + [guid]::NewGuid().ToString("N") + ".txt")
    $caught = $null
    try {
        & $verify -Tag $Tag -GitHubOutput $outputFile | Out-Null
    } catch {
        $caught = $_.Exception.Message
    }
    try {
        if ($ShouldPass) {
            if ($null -ne $caught) { throw "Expected tag '$Tag' accepted; got: $caught" }
            $lines = @(Get-Content -LiteralPath $outputFile)
            if ("release_note=$Note" -notin $lines) { throw "Wrong release-note mapping for '$Tag': $($lines -join '; ')" }
            if ("prerelease=$($Prerelease.ToString().ToLowerInvariant())" -notin $lines) { throw "Wrong prerelease flag for '$Tag'" }
            if ("tag=$Tag" -notin $lines) { throw "Wrong exact tag output for '$Tag'" }
            $commitLine = @($lines | Where-Object { $_ -match '^commit=[0-9a-f]{40}$' })
            if ($commitLine.Count -ne 1) { throw "Missing candidate commit output for '$Tag'" }
        } else {
            if ($null -eq $caught) { throw "Expected tag '$Tag' rejected" }
            if (-not $caught.Contains($ErrorContains)) {
                throw "Wrong rejection for tag '$Tag'. Expected '$ErrorContains'; received '$caught'"
            }
        }
        $script:passCount++
        Write-Output "PASS $($script:passCount): [$Tag]"
    } finally {
        Remove-Item -LiteralPath $outputFile -Force -ErrorAction SilentlyContinue
    }
}

Push-Location $repoRoot
try {
    # Test-only notes are always cleaned and are never committed or published.
    foreach ($filename in @("0.7.0.md", "0.7.0-rc.2.md", "0.7.0-rc.42.md")) {
        $path = Join-Path $repoRoot ("docs/releases/" + $filename)
        if (-not (Test-Path -LiteralPath $path)) {
            [System.IO.File]::WriteAllText($path, "# Temporary release policy test note`n", [System.Text.UTF8Encoding]::new($false))
            $createdNotes.Add($path)
        }
    }
    Assert-ReleaseContext -Tag "" -ShouldPass $true -Note "docs/releases/0.7.0-rc.md"
    Assert-ReleaseContext -Tag "v0.7.0" -ShouldPass $true -Note "docs/releases/0.7.0.md"
    Assert-ReleaseContext -Tag "v0.7.0-rc.1" -ShouldPass $true -Note "docs/releases/0.7.0-rc.1.md" -Prerelease $true
    Assert-ReleaseContext -Tag "v0.7.0-rc.2" -ShouldPass $true -Note "docs/releases/0.7.0-rc.2.md" -Prerelease $true
    Assert-ReleaseContext -Tag "v0.7.0-rc.42" -ShouldPass $true -Note "docs/releases/0.7.0-rc.42.md" -Prerelease $true

    foreach ($tag in @(
        "v0.7.1-rc.1", "v0.8.0-rc.1", "v0.7.0-rc.0",
        "v0.7.0-rc.", "v0.7.0-rc.a", "v0.7.0-beta.1",
        "0.7.0-rc.1", "vv0.7.0-rc.1", "v0.7.0-rc.1-wrong",
        "v0.7.0-rc.01", " v0.7.0-rc.1", "v0.7.0-rc.1 ",
        "v0.7.0-RC.1"
    )) {
        Assert-ReleaseContext -Tag $tag -ShouldPass $false -ErrorContains "does not match authoritative version"
    }
    Assert-ReleaseContext -Tag "v0.7.0-rc.3" -ShouldPass $false -ErrorContains "Release note does not exist"
} finally {
    foreach ($path in $createdNotes) {
        Remove-Item -LiteralPath $path -Force -ErrorAction SilentlyContinue
    }
    Pop-Location
}
Write-Output "RELEASE_TAG_POLICY_TESTS_PASS=$passCount"
