mkdir build
REM go build -ldflags -H=windowsgui -o build\universal_menu_main.exe .\gui
go build -o build\universal_menu_main.exe .\glfw
go build -o build\tray.exe .\tray
go build -o build\universal_menu_command_line_toggle.exe helpers\command_line_toggle.go
gcc helpers\windowsGlobalKeyHook.c -obuild/universal_menu_hotkey_monitor.exe
gcc helpers\hotcorner.c -obuild/universal_menu_hotcorner_monitor.exe
cp -r tray/config build/
rm -r traymenu
mv build traymenu
rm traymenu.zip
powershell Compress-Archive  traymenu traymenu.zip
