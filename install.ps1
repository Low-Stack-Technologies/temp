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

    Write-Host "Installing Context menu entry..."
    try {
        # No admin rights needed for this version
        $registryPath = "Registry::HKEY_CURRENT_USER\Software\Classes\*\shell\TempUpload"
    
        # Create the main menu item quietly
        $null = New-Item -Path $registryPath -Force
        $null = New-ItemProperty -Path $registryPath -Name "MUIVerb" -Value "Temporary Upload (24h)" -PropertyType String -Force
    
        # Create the command subkey quietly
        $null = New-Item -Path "$registryPath\command" -Force
    
        # Define the command script separately to avoid nesting issues
        $commandScript = 'powershell.exe -Command "' + `
            'Add-Type -AssemblyName System.Windows.Forms;' + `
            'try {' + `
            '    $filePath = [System.IO.Path]::GetFullPath(''%1'');' + `
            '    Write-Host "Uploading $filePath";' + `
            '    $result = & upload-temp $filePath --expiration 24h | Out-String;' + `
            '    $url = $result | Select-String -Pattern ''(https?://[^\s]+)'' | ForEach-Object { $_.Matches[0].Value };' + `
            '    if ($url) {' + `
            '        [System.Windows.Forms.Clipboard]::SetText($url);' + `
            '        [System.Windows.Forms.MessageBox]::Show(''Download URL copied to clipboard'', ''Upload complete'', [System.Windows.Forms.MessageBoxButtons]::OK, [System.Windows.Forms.MessageBoxIcon]::Information);' + `
            '    } else {' + `
            '        [System.Windows.Forms.MessageBox]::Show(''Upload failed - No URL found in output'', ''Warning'', [System.Windows.Forms.MessageBoxButtons]::OK, [System.Windows.Forms.MessageBoxIcon]::Warning);' + `
            '    }' + `
            '} catch {' + `
            '    [System.Windows.Forms.MessageBox]::Show($_.Exception.Message, ''Error'', [System.Windows.Forms.MessageBoxButtons]::OK, [System.Windows.Forms.MessageBoxIcon]::Error);' + `
            '}"'
    
        # Set the command
        $null = New-ItemProperty -Path "$registryPath\command" -Name "(Default)" -Value $commandScript -PropertyType String -Force
    
        Write-Host "Context menu entry added successfully."
    }
    catch {
        Write-Host "Error: $($_.Exception.Message)"
    }

    Write-Host "Installation complete - please restart your terminal and Windows Explorer"
}
catch {
    Write-Host "An error occurred: $_"
    exit 1
}
