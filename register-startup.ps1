$startupDir = [Environment]::GetFolderPath('Startup')
$shortcutPath = Join-Path $startupDir "DeepSeek Free API Bridge.lnk"
$targetPath = Join-Path $PSScriptRoot "start-deepseek.bat"
$workDir = $PSScriptRoot

$wshShell = New-Object -ComObject WScript.Shell
$shortcut = $wshShell.CreateShortcut($shortcutPath)
$shortcut.TargetPath = $targetPath
$shortcut.WorkingDirectory = $workDir
$shortcut.WindowStyle = 7
$shortcut.Description = "Auto-start DeepSeek Free API Bridge (Port 8000)"
$shortcut.Save()

Write-Output "SUCCESS: Created auto-start shortcut at $shortcutPath"
