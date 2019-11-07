sudo apt-get install -y libegl1-mesa-dev libgles2-mesa-dev libx11-dev libxrandr-dev libxinerama-dev  libxinerama1 libxcursor-dev libxi-dev xorg-dev libglu1-mesa-dev freeglut3-dev mesa-common-dev
mkdir build
go build -ldflags="-s -w" -o build/universal_menu_main.exe
go build -ldflags="-s -w" -o build/universal_menu_command_line_toggle helpers/command_line_toggle.go
gcc -Ofast -Wall -o mac_launcher helpers/macGlobalKeyHook.c  -framework ApplicationServices -o build/universal_menu_hotkey_monitor
ls -la build/
