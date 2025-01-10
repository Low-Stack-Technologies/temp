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
$installDir = "$env:LOCALAPPDATA\temp"
$downloadUrl = "https://github.com/low-stack/temp/releases/latest/download/temp-windows-$arch.exe"

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

    Write-Host "Installation complete - please restart your terminal"
}
catch {
    Write-Host "An error occurred: $_"
    exit 1
}