#!/bin/sh
export GO111MODULE=auto
go get github.com/mostlygeek/arp github.com/getlantern/systray golang.org/x/sync/semaphore github.com/emersion/go-imap github.com/mattn/go-shellwords github.com/schollz/closestmatch github.com/atotto/clipboard github.com/emersion/go-autostart  github.com/emersion/go-sasl
mkdir build
cd build
rm universal_menu_main.exe universal_menu_command_line_toggle universal_menu_hotkey_monitor universal_launcher_main.exe KeyTap_mac tray_menu.exe
cd ..
go build -o build/universal_launcher_main.exe ./launcher
go build -o build/universal_menu_main.exe ./glfw2
go build -o build/tray_menu.exe ./tray
go build -o build/tray_headless.exe ./trayheadless
go build -ldflags="-s -w" -o build/universal_menu_command_line_toggle helpers/command_line_toggle.go
gcc -Ofast -Wall -o mac_launcher helpers/macGlobalKeyHook.c  -framework ApplicationServices -o build/universal_menu_hotkey_monitor
gcc -Ofast -Wall -o mac_launcher helpers/KeyTap_mac.c  -framework ApplicationServices -o build/KeyTap_mac
cp -r build/* traymenu/
mkdir ~/.umh
cp -rn tray/config_examples traymenu/
cp -rn tray/config_examples ~/.umh
tar -czvf traymenu.tar.gz traymenu
ls -la traymenu/
