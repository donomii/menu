# menu

An experimental system for displaying user menus


##  Build

	go build -ldflags -H=windowsgui .
	g++ helpers\windowsGlobalKeyHook.c -o menu_launcher
    gcc -Wall -o mac_launcher testkeyhandler.c -framework ApplicationServices


