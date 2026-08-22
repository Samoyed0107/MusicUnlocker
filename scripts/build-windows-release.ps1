[CmdletBinding()]
param(
    [switch]$PortableOnly,
    [switch]$AllowDirty,
    [string]$ArtifactDirectory = "artifacts"
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$repoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot ".."))
$artifactRoot = [System.IO.Path]::GetFullPath((Join-Path $repoRoot $ArtifactDirectory))
$repoPrefix = $repoRoot.TrimEnd([System.IO.Path]::DirectorySeparatorChar) + [System.IO.Path]::DirectorySeparatorChar
if (-not $artifactRoot.StartsWith($repoPrefix, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "Artifact directory must resolve inside the repository: $artifactRoot"
}

function Invoke-Checked {
    param([Parameter(Mandatory)][string]$Command, [Parameter(Mandatory)][string[]]$Arguments)
    & $Command @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "$Command failed with exit code $LASTEXITCODE"
    }
}

function New-DeterministicZip {
    param(
        [Parameter(Mandatory)][string]$SourceDirectory,
        [Parameter(Mandatory)][string]$Destination,
        [Parameter(Mandatory)][DateTimeOffset]$Timestamp
    )
    Add-Type -AssemblyName System.IO.Compression
    $stream = [System.IO.File]::Open($Destination, [System.IO.FileMode]::CreateNew)
    try {
        $archive = [System.IO.Compression.ZipArchive]::new($stream, [System.IO.Compression.ZipArchiveMode]::Create, $false)
        try {
            Get-ChildItem -LiteralPath $SourceDirectory -File -Recurse | Sort-Object FullName | ForEach-Object {
                $relative = [System.IO.Path]::GetRelativePath($SourceDirectory, $_.FullName).Replace("\", "/")
                $entry = $archive.CreateEntry($relative, [System.IO.Compression.CompressionLevel]::Optimal)
                $entry.LastWriteTime = $Timestamp
                $input = [System.IO.File]::OpenRead($_.FullName)
                $output = $entry.Open()
                try { $input.CopyTo($output) } finally { $output.Dispose(); $input.Dispose() }
            }
        } finally { $archive.Dispose() }
    } finally { $stream.Dispose() }
}

Push-Location $repoRoot
try {
    $worktreeStatus = (& git status --porcelain=v1)
    if ($LASTEXITCODE -ne 0) { throw "Unable to inspect the Git worktree" }
    if ($worktreeStatus -and -not $AllowDirty) {
        throw "Release builds require a clean Git worktree. Commit or remove scoped changes, or use -AllowDirty only for local engineering validation."
    }

    $release = (Invoke-Checked "go" @("run", "./cmd/releaseinfo") | ConvertFrom-Json)
    if (-not $release.name -or -not $release.version) { throw "Release identity is incomplete" }
    if (-not $PortableOnly -and -not (Get-Command "makensis" -ErrorAction SilentlyContinue)) {
        throw "NSIS 3.x is required to build the installer; makensis was not found on PATH. No software was installed and existing artifacts were left untouched."
    }

    $commit = (Invoke-Checked "git" @("rev-parse", "HEAD") | Select-Object -Last 1).Trim()
    $commitTimestampText = (Invoke-Checked "git" @("show", "-s", "--format=%cI", "HEAD") | Select-Object -Last 1).Trim()
    $commitTimestamp = [DateTimeOffset]::Parse($commitTimestampText).ToUniversalTime()
    $buildTimestamp = $commitTimestamp.ToString("yyyy-MM-ddTHH:mm:ssZ")
    $buildIdentifier = $commit.Substring(0, 12)

    if (Test-Path -LiteralPath $artifactRoot) {
        Remove-Item -LiteralPath $artifactRoot -Recurse -Force
    }
    New-Item -ItemType Directory -Path $artifactRoot | Out-Null

    $ldflags = "-X musicunlocker/internal/version.BuildTimestamp=$buildTimestamp -X musicunlocker/internal/version.BuildIdentifier=$buildIdentifier"
    $wailsArgs = @("build", "-clean", "-trimpath", "-platform", "windows/amd64", "-webview2", "embed", "-installscope", "user", "-ldflags", $ldflags)
    if (-not $PortableOnly) {
        $wailsArgs += "-nsis"
    }
    Invoke-Checked "wails" $wailsArgs

    $binary = Join-Path $repoRoot "build/bin/$($release.name).exe"
    if (-not (Test-Path -LiteralPath $binary -PathType Leaf)) { throw "Wails executable was not produced: $binary" }

    $stageRoot = Join-Path $artifactRoot ".portable-stage"
    $packageRoot = Join-Path $stageRoot "$($release.name)-$($release.version)-win64-portable"
    New-Item -ItemType Directory -Path $packageRoot | Out-Null
    Copy-Item -LiteralPath $binary -Destination (Join-Path $packageRoot "$($release.name).exe")
    Copy-Item -LiteralPath (Join-Path $repoRoot "LICENSE") -Destination (Join-Path $packageRoot "LICENSE.txt")
    Copy-Item -LiteralPath (Join-Path $repoRoot "THIRD_PARTY_NOTICES.md") -Destination (Join-Path $packageRoot "THIRD_PARTY_NOTICES.txt")
    Copy-Item -LiteralPath (Join-Path $repoRoot "build/windows/portable/README.txt") -Destination (Join-Path $packageRoot "README.txt")

    $portableName = "$($release.name)-$($release.version)-win64-portable.zip"
    $portablePath = Join-Path $artifactRoot $portableName
    New-DeterministicZip -SourceDirectory $stageRoot -Destination $portablePath -Timestamp $commitTimestamp
    Remove-Item -LiteralPath $stageRoot -Recurse -Force

    $outputs = @($portablePath)
    if (-not $PortableOnly) {
        $installerName = "$($release.name)-Setup-$($release.version)-win64.exe"
        $installerSource = Join-Path $repoRoot "build/bin/$installerName"
        if (-not (Test-Path -LiteralPath $installerSource -PathType Leaf)) { throw "NSIS installer was not produced: $installerSource" }
        $installerPath = Join-Path $artifactRoot $installerName
        Copy-Item -LiteralPath $installerSource -Destination $installerPath
        $outputs = @($installerPath, $portablePath)
    }

    $checksumPath = Join-Path $artifactRoot "SHA256SUMS"
    $checksumLines = foreach ($output in $outputs | Sort-Object { Split-Path -Leaf $_ }) {
        $hash = (Get-FileHash -Algorithm SHA256 -LiteralPath $output).Hash.ToLowerInvariant()
        "$hash  $(Split-Path -Leaf $output)"
    }
    [System.IO.File]::WriteAllLines($checksumPath, $checksumLines, [System.Text.UTF8Encoding]::new($false))

    $outputs + $checksumPath | ForEach-Object { Get-Item -LiteralPath $_ | Select-Object Name, Length, LastWriteTimeUtc }
} finally {
    Pop-Location
}
