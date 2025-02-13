# Function to detect architecture
function Get-Architecture {
    $arch = [System.Environment]::GetEnvironmentVariable("PROCESSOR_ARCHITECTURE")
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default {
            Write-Host "Unsupported architecture: $arch"
            exit 1
        }
    }
}

# Set variables
$arch = Get-Architecture
$installDir = "$env:LOCALAPPDATA\lowstack-temp"

# Fetch latest release URL from GitHub API
try {
    Write-Host "Fetching latest release..."
    $releases = Invoke-RestMethod -Uri "https://api.github.com/repos/low-stack-technologies/temp/releases"
    $downloadUrl = $releases[0].assets |
    Where-Object { $_.browser_download_url -like "*temp-windows-$arch.exe" } |
    Select-Object -ExpandProperty browser_download_url -First 1
    if (-not $downloadUrl) {
        Write-Host "Failed to find appropriate release for windows-$arch"
        exit 1
    }
}
catch {
    Write-Host "Failed to fetch latest release: $_"
    exit 1
}

# Create installation directory if it doesn't exist
New-Item -ItemType Directory -Force -Path $installDir | Out-Null

try {
    # Download the binary
    Write-Host "Downloading temp binary..."
    Invoke-WebRequest -Uri $downloadUrl -OutFile "$installDir\temp.exe"
    
    # Add to PATH if not already present
    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    if ($userPath -notlike "*$installDir*") {
        [Environment]::SetEnvironmentVariable('Path', "$userPath;$installDir", 'User')
        Write-Host "Added to PATH successfully"
    }

    # Create or update PowerShell profile to add the alias function
    $profilePath = $PROFILE.CurrentUserAllHosts
    $profileDir = Split-Path $profilePath -Parent
    
    # Create profile directory if it doesn't exist
    if (-not (Test-Path $profileDir)) {
        New-Item -ItemType Directory -Force -Path $profileDir | Out-Null
    }

    # Create or append to profile
    $aliasFunction = @'

# Alias function for upload-temp
function upload-temp {
    $exePath = "$env:LOCALAPPDATA\lowstack-temp\temp.exe"
    if (Test-Path $exePath) {
        & $exePath $args
    } else {
        Write-Host "temp.exe not found. Please run the installer again."
    }
}
'@

    if (Test-Path $profilePath) {
        # Check if the function already exists in the profile
        $profileContent = Get-Content $profilePath -Raw
        if ($profileContent -notlike "*function upload-temp*") {
            Add-Content $profilePath $aliasFunction
        }
    }
    else {
        # Create new profile with the function
        Set-Content $profilePath $aliasFunction
    }

    Write-Host "Installation complete - please restart your PowerShell terminal"
}
catch {
    Write-Host "An error occurred: $_"
    exit 1
}