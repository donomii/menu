sudo apt-get install -y libegl1-mesa-dev libgles2-mesa-dev libx11-dev libxrandr-dev libxinerama-dev  libxinerama1 libxcursor-dev libxi-dev xorg-dev libglu1-mesa-dev freeglut3-dev mesa-common-dev
mkdir build
go get github.com/AllenDang/giu/imgui
go build -o build/universal_menu_main.exe ./glfw/
go build -o build/universal_menu_command_line_toggle helpers/command_line_toggle.go
go build -o build/tray.exe ./tray/
#FIXME compile touchbar/global hotkey program here
