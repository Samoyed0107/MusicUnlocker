[CmdletBinding()]
param(
    [string]$ArtifactDirectory,
    [string]$Tag,
    [string]$GitHubOutput
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$repoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot ".."))

function Invoke-Checked {
    param([Parameter(Mandatory)][string]$Command, [Parameter(Mandatory)][string[]]$Arguments)
    $output = & $Command @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "$Command failed with exit code $LASTEXITCODE"
    }
    $output
}

Push-Location $repoRoot
try {
    $release = (Invoke-Checked "go" @("run", "./cmd/releaseinfo") | ConvertFrom-Json)
    if (-not $release.name -or $release.name -notmatch '^[A-Za-z0-9._-]+$') {
        throw "Release name is missing or unsafe for artifact filenames"
    }
    if (-not $release.version -or $release.version -notmatch '^\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?$') {
        throw "Release version is missing or is not a supported semantic version"
    }

    # Keep the application version independent of the exact prerelease tag.
    $expectedStableTag = "v$($release.version)"
    $releaseNoteRelative = "docs/releases/$($release.version)-rc.md"
    $isPrerelease = $false
    if ($Tag) {
        if ($Tag -ceq $expectedStableTag) {
            $releaseNoteRelative = "docs/releases/$($release.version).md"
        } elseif ($Tag -cmatch ('^' + [regex]::Escape($expectedStableTag) + '-rc\.([1-9][0-9]*)$')) {
            $releaseNoteRelative = "docs/releases/$($release.version)-rc.$($Matches[1]).md"
            $isPrerelease = $true
        } else {
            throw "Release tag '$Tag' does not match authoritative version '$($release.version)' (expected '$expectedStableTag' or -rc.N starting at 1)"
        }
    }
    $releaseNote = Join-Path $repoRoot $releaseNoteRelative
    if (-not (Test-Path -LiteralPath $releaseNote -PathType Leaf)) {
        throw "Release note does not exist for tag '$Tag': $releaseNoteRelative"
    }

    $applicationNames = @(
        "$($release.name)-Setup-$($release.version)-win64.exe",
        "$($release.name)-$($release.version)-win64.exe",
        "$($release.name)-$($release.version)-win64-portable.zip"
    )
    $expectedNames = $applicationNames + @("LICENSE.txt", "THIRD_PARTY_NOTICES.txt", "SHA256SUMS")

    if ($ArtifactDirectory) {
        $artifactRoot = [System.IO.Path]::GetFullPath((Join-Path $repoRoot $ArtifactDirectory))
        $repoPrefix = $repoRoot.TrimEnd([System.IO.Path]::DirectorySeparatorChar) + [System.IO.Path]::DirectorySeparatorChar
        if (-not $artifactRoot.StartsWith($repoPrefix, [System.StringComparison]::OrdinalIgnoreCase)) {
            throw "Artifact directory must resolve inside the repository: $artifactRoot"
        }
        if (-not (Test-Path -LiteralPath $artifactRoot -PathType Container)) {
            throw "Release artifact directory does not exist: $artifactRoot"
        }

        $entries = @(Get-ChildItem -LiteralPath $artifactRoot -Force)
        $unexpected = @($entries | Where-Object { $_.PSIsContainer -or $_.Name -notin $expectedNames })
        $missing = @($expectedNames | Where-Object { -not (Test-Path -LiteralPath (Join-Path $artifactRoot $_) -PathType Leaf) })
        if ($unexpected.Count -gt 0 -or $missing.Count -gt 0 -or $entries.Count -ne $expectedNames.Count) {
            throw "Release allowlist failed. Unexpected: [$((@($unexpected.Name) | Sort-Object) -join ', ')]. Missing: [$((@($missing) | Sort-Object) -join ', ')]."
        }

        $manifestPath = Join-Path $artifactRoot "SHA256SUMS"
        $manifestNames = [System.Collections.Generic.HashSet[string]]::new([System.StringComparer]::Ordinal)
        $manifestLines = @(Get-Content -LiteralPath $manifestPath)
        foreach ($line in $manifestLines) {
            if ($line -notmatch '^([0-9a-f]{64})  ([^\\/]+)$') {
                throw "Malformed SHA256SUMS line: $line"
            }
            $expectedHash = $Matches[1]
            $name = $Matches[2]
            if ($name -eq "SHA256SUMS" -or $name -notin $expectedNames) {
                throw "SHA256SUMS contains an unexpected filename: $name"
            }
            if (-not $manifestNames.Add($name)) {
                throw "SHA256SUMS contains a duplicate filename: $name"
            }
            $actualHash = (Get-FileHash -Algorithm SHA256 -LiteralPath (Join-Path $artifactRoot $name)).Hash.ToLowerInvariant()
            if ($actualHash -ne $expectedHash) {
                throw "SHA-256 mismatch for $name. Expected $expectedHash, got $actualHash"
            }
        }

        $expectedManifestNames = @($expectedNames | Where-Object { $_ -ne "SHA256SUMS" })
        $manifestMissing = @($expectedManifestNames | Where-Object { -not $manifestNames.Contains($_) })
        if ($manifestLines.Count -ne $expectedManifestNames.Count -or $manifestMissing.Count -gt 0) {
            throw "SHA256SUMS contract failed. Missing entries: [$((@($manifestMissing) | Sort-Object) -join ', ')]."
        }
    }

    if ($GitHubOutput) {
        @(
            "name=$($release.name)",
            "version=$($release.version)",
            "tag=$Tag",
            "release_note=$releaseNoteRelative",
            "prerelease=$($isPrerelease.ToString().ToLowerInvariant())",
            "commit=$((Invoke-Checked "git" @("rev-parse", "HEAD") | Select-Object -Last 1).Trim())",
            "artifact_name=$($release.name)-$($release.version)-windows-release"
        ) | Add-Content -LiteralPath $GitHubOutput -Encoding utf8
    }

    [pscustomobject]@{
        Name = $release.name
        Version = $release.version
        ExpectedTag = $(if ($Tag) { $Tag } else { $expectedStableTag })
        IsPrerelease = $isPrerelease
        ReleaseNote = $releaseNote
        ArtifactsVerified = [bool]$ArtifactDirectory
    }
} finally {
    Pop-Location
}
