# Check if running with administrator privileges
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)

# Function to handle errors consistently
function Write-ErrorAndExit {
    param([string]$message)
    Write-Host "Error: $message" -ForegroundColor Red
    Write-Host "Please contact support if this issue persists." -ForegroundColor Yellow
    exit 1
}

# Function to remove profile function and aliases
function Remove-ProfileAliases {
    $profilePaths = @(
        $PROFILE.CurrentUserAllHosts,
        $PROFILE.CurrentUserCurrentHost
    )

    foreach ($profilePath in $profilePaths) {
        if (Test-Path $profilePath) {
            $content = Get-Content $profilePath -Raw
            $modified = $false

            # Array of patterns to remove
            $aliasPatterns = @(
                'Set-Alias.*temp.*',
                'Set-Alias.*lowstack-temp.*',
                'Set-Alias.*upload-temp.*',
                'New-Alias.*temp.*',
                'New-Alias.*lowstack-temp.*',
                'New-Alias.*upload-temp.*',
                'function\s+temp\s*\{[\s\S]*?\}',
                'function\s+upload-temp\s*\{[\s\S]*?\}'
            )

            foreach ($pattern in $aliasPatterns) {
                if ($content -match $pattern) {
                    $content = $content -replace "(?m)^\s*$pattern\s*$", ""
                    $modified = $true
                }
            }

            # Remove the function block if it exists
            if ($content -match '# Alias function for (temp|upload-temp)') {
                $content = $content -replace '(?ms)# Alias function for (temp|upload-temp).*?}\r?\n?', ''
                $modified = $true
            }

            if ($modified) {
                # Remove any double blank lines created by removals
                $content = $content -replace "`n\s*`n\s*`n", "`n`n"
                Set-Content -Path $profilePath -Value $content
                Write-Host "Removed aliases and functions from $profilePath" -ForegroundColor Green
            }
        }
    }
}

# Set the installation directory path
$installDir = "$env:LOCALAPPDATA\lowstack-temp"
$exePath = "$installDir\temp.exe"

try {
    # Always attempt to remove aliases regardless of exe presence
    Write-Host "Removing PowerShell aliases and functions..."
    Remove-ProfileAliases

    # Check if the service exists
    if (-not (Test-Path $exePath)) {
        Write-Host "temp.exe not found at $exePath. It may have already been uninstalled." -ForegroundColor Yellow
    }
    else {
        # Check for running processes and attempt to stop them
        $processes = Get-Process | Where-Object { $_.Path -eq $exePath }
        if ($processes) {
            Write-Host "Found running temp processes. Attempting to stop them..."
            foreach ($process in $processes) {
                try {
                    $process | Stop-Process -Force
                    Write-Host "Stopped process with ID $($process.Id)"
                }
                catch {
                    Write-ErrorAndExit "Failed to stop process with ID $($process.Id). Please close all temp applications and try again."
                }
            }
            # Wait a moment for processes to fully terminate
            Start-Sleep -Seconds 2
        }

        # Remove from PATH
        Write-Host "Removing temp from PATH..."
        $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
        if ($userPath) {
            $pathEntries = $userPath -split ';' | Where-Object { $_ -and $_ -ne $installDir }
            $newPath = $pathEntries -join ';'
            [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
            Write-Host "Removed from PATH successfully" -ForegroundColor Green
        }
        else {
            Write-Host "No User PATH found to modify" -ForegroundColor Yellow
        }

        # Remove the binary and installation directory
        Write-Host "Removing temp binary and installation directory..."
        
        # Try to remove the exe file first
        if (Test-Path $exePath) {
            try {
                Remove-Item $exePath -Force -ErrorAction Stop
                Write-Host "Removed temp.exe successfully" -ForegroundColor Green
            }
            catch {
                Write-ErrorAndExit "Failed to remove temp.exe. Error: $_"
            }
        }

        # Then try to remove the directory
        if (Test-Path $installDir) {
            try {
                # Check if directory is empty before removal
                $remainingFiles = Get-ChildItem -Path $installDir -Force
                if ($remainingFiles) {
                    Write-Host "Warning: Additional files found in $installDir" -ForegroundColor Yellow
                    foreach ($file in $remainingFiles) {
                        Write-Host "- $($file.Name)"
                    }
                    $confirmation = Read-Host "Do you want to remove all files in the directory? (Y/N)"
                    if ($confirmation -eq 'Y') {
                        Remove-Item $installDir -Force -Recurse
                        Write-Host "Removed installation directory successfully" -ForegroundColor Green
                    }
                    else {
                        Write-Host "Directory removal skipped by user" -ForegroundColor Yellow
                    }
                }
                else {
                    Remove-Item $installDir -Force -Recurse
                    Write-Host "Removed empty installation directory successfully" -ForegroundColor Green
                }
            }
            catch {
                Write-ErrorAndExit "Failed to remove installation directory. Error: $_"
            }
        }

        # Clean up any remaining environment variables
        $envVars = [Environment]::GetEnvironmentVariables('User')
        foreach ($var in $envVars.Keys) {
            if ($envVars[$var] -like "*$installDir*") {
                try {
                    [Environment]::SetEnvironmentVariable($var, $null, 'User')
                    Write-Host "Removed environment variable: $var" -ForegroundColor Green
                }
                catch {
                    Write-Host "Warning: Failed to remove environment variable: $var" -ForegroundColor Yellow
                }
            }
        }
    }

    Write-Host "`nUninstallation completed successfully!" -ForegroundColor Green
    Write-Host "Please restart your terminal for the changes to take effect." -ForegroundColor Yellow
}
catch {
    Write-ErrorAndExit "An unexpected error occurred: $_"
}