# VHS Tape - RepoLens TUI Recording
# Render with: vhs < demo.tape

Output demo.gif

Set Shell "zsh"
Set FontSize 15
Set FontFamily "DejaVu Sans Mono"
Set Width 1200
Set Height 700
Set Padding 20
Set Theme "Chalk"

# Start the application
Type "./repolens" Sleep 1s Enter
Sleep 1.5s

# Navigate and interact
Down Sleep 800ms
Down Sleep 800ms

# Collapse packer directory
Left Sleep 1.2s
Sleep 500ms

# Go to README.md
Down Sleep 800ms

# Toggle select
Space Sleep 1s

# Move to main.go
Down Sleep 800ms

# Toggle code comment compression c
Type "c" Sleep 1.5s

# Toggle back
Type "c" Sleep 1s

# Pack and Copy
Enter
Sleep 1.5s
