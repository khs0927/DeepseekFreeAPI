$startupDir = [Environment]::GetFolderPath('Startup')
$shortcutPath = Join-Path $startupDir "DeepSeek Free API Bridge.lnk"

if (Test-Path $shortcutPath) {
    Remove-Item -Path $shortcutPath -Force
    Write-Output "SUCCESS: Removed auto-start shortcut from $shortcutPath"
} else {
    Write-Output "INFO: Shortcut does not exist in Startup folder."
}
