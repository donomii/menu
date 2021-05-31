# menu

A cross-platform experimental system for displaying user menus, that turned into a popup launcher and more.

There are currently two programs available: a tray menu, and a popup launcher.

# Features

UMH has picked up a lot of different abilities.  It can scan your network for other machines and services, publish services of its own, launch applications, open files to edit, open urls and run shell commands.

# Tray menu

The tray menu sits in the system tray(Windows) or the menu bar (Mac).  It displays the user-defined menu, as well as several generated menus, like the Applications menu, and the network services menu.

# Popup launcher

The popup launcher works similar to the built-in search on macosx.  Press a button (the spotlight button on mac, CAPS LOCK on windows), and the launcher will appear.  Type your search, then use the up and down arrows to select your choice.

The options that appear in the popup launcher are NOT automatically generated, they come from a file (~/.menu.recall.text).  You can always find this file by searching for "menu settings", then selecting that option from the list.  

The different ways to launch a file or program are described below.

# Use

Run tray_menu.exe.  This starts the tray menu, and from there you can start the popup launcher from the user menu.

For command line scripting, there is a command line program that does the same as the hotkey: ```universal_menu_command_line_toggle.exe```.  This works with AppleScript and other scripting platforms.  I use it with the program "My TouchBar, My Rules" to add a touchbar button that will open and close the launcher.

##  Build

Use the provided build scripts, build.bat and build.sh

# Command format

UMH uses urls where ever possible, and adds its own extensions.  The common supported ones are:

- http://,https:// - Will open a web browser to this url
- shell:// - will run a shell command exactly as written.  The shell will be cmd.exe or bash.
- exec:// - attempt to run this program.  You can add simple arguments after the program name, as if you were in the shell (but in a really dumb shell)
- file:// - Open the file for display or editing using the default program
- internal:// - There are several internal commands provided, e.g. RunAtStartup
- clipboard:// - Copy the rest of the url to the clipboard(minus the clipboard:// part).  Handy if you know you have to keep typing something.  You probably shouldn't put your passwords here, but I do anyway