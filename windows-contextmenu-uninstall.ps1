Write-Host "Removing Context menu entry..."
try {
    # No admin rights needed for this version
    Remove-Item -Path "Registry::HKEY_CURRENT_USER\Software\Classes\*\shell\TempUpload" -Recurse -Force
    Write-Host "Context menu entry removed successfully."
}
catch {
    Write-Host "Error: $($_.Exception.Message)"
}