Write-Host "Installing Context menu entry..."
try {
    # No admin rights needed for this version
    $registryPath = "Registry::HKEY_CURRENT_USER\Software\Classes\*\shell\TempUpload"

    # Create the main menu item quietly
    $null = New-Item -Path $registryPath -Force
    $null = New-ItemProperty -Path $registryPath -Name "MUIVerb" -Value "Upload to TEMP (24h)" -PropertyType String -Force

    # Create the command subkey quietly
    $null = New-Item -Path "$registryPath\command" -Force
    $commandScript = @'
powershell.exe -WindowStyle Hidden -Command "
    Add-Type -AssemblyName System.Windows.Forms;
    try {
        # Get the full file path
        $filePath = [System.IO.Path]::GetFullPath('%1');
        
        # Run the command and get the output
        $result = & upload-temp $filePath --expiration 24h | Out-String;
        
        # Extract only the URL using regex
        $url = $result | Select-String -Pattern 'https://.*' | ForEach-Object { $_.Matches[0].Value };
        
        if ($url) {
            [System.Windows.Forms.Clipboard]::SetText($url);
            [System.Windows.Forms.MessageBox]::Show('Download URL copied to clipboard', 'Upload complete', [System.Windows.Forms.MessageBoxButtons]::OK, [System.Windows.Forms.MessageBoxIcon]::Information);
        } else {
            [System.Windows.Forms.MessageBox]::Show('Upload failed - No URL found in output', 'Warning', [System.Windows.Forms.MessageBoxButtons]::OK, [System.Windows.Forms.MessageBoxIcon]::Warning);
        }
    }
    catch {
        [System.Windows.Forms.MessageBox]::Show($_.Exception.Message, 'Error', [System.Windows.Forms.MessageBoxButtons]::OK, [System.Windows.Forms.MessageBoxIcon]::Error);
    }
"
'@
    $null = New-ItemProperty -Path "$registryPath\command" -Name "(Default)" -Value $commandScript -PropertyType String -Force

    Write-Host "Context menu entry added successfully."
}
catch {
    Write-Host "Error: $($_.Exception.Message)"
}