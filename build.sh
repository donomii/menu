mkdir build
rm build/universal_menu_main.exe build/universal_menu_command_line_toggle build/universal_menu_hotkey_monitor
go build -ldflags="-s -w" -o build/universal_menu_main.exe
go build -o build/universal_menu_main.exe ./glfw
go build -o build/tray_menu.exe ./tray
go build -ldflags="-s -w" -o build/universal_menu_command_line_toggle helpers/command_line_toggle.go
gcc -Ofast -Wall -o mac_launcher helpers/macGlobalKeyHook.c  -framework ApplicationServices -o build/universal_menu_hotkey_monitor
ls -la build/
