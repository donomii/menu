mkdir build
go build -o build/universal_menu_main.exe
go build -o build/universal_menu_command_line_toggle helpers/command_line_toggle.go
gcc -Wall -o mac_launcher helpers/macGlobalKeyHook.c  -framework ApplicationServices -o build/universal_menu_hotkey_monitor
