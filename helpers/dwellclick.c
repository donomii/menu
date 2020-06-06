#define WIN32_LEAN_AND_MEAN
#include <stdlib.h>
#include <windows.h>
#include <stdio.h>
#include <shellapi.h>
#include <fileapi.h>
#include <process.h>
#include <Tlhelp32.h>
#include <winbase.h>

#pragma comment(lib, "USER32")
#pragma comment(linker, "/SUBSYSTEM:WINDOWS")

#define KEYDOWN(k) ((k) & 0x80)

// This is a **very** minimal hotcorner app, written in C. Maybe its not the
// optimal way to do this, but it works for me.
//
// Zero state is stored anywhere, no registry keys or configuration files.
//
// - If you want to configure something, edit the code.
// - If you want to uninstall it, just delete it.
//
// Tavis Ormandy <taviso@cmpxchg8b.com> December, 2016
//
// https://github.com/taviso/hotcorner
//

// If the mouse enters this rectangle, activate the hot corner function.
// There are some hints about changing corners here
//      https://github.com/taviso/hotcorner/issues/7#issuecomment-269367351
static const RECT kHotCorner = {
    .top    = -20,
    .left   = -20,
    .right  = +20,
    .bottom = +20,
};

int menu_active = 0;
char * pidfile;
char * targetProgram = "universal_menu_main.exe";
// Input to inject when corner activated (Win+Tab by default).
static const INPUT kCornerInput[] = {
    { INPUT_KEYBOARD, .ki = { VK_LWIN, .dwFlags = 0 }},
    { INPUT_KEYBOARD, .ki = { VK_TAB,  .dwFlags = 0 }},
    { INPUT_KEYBOARD, .ki = { VK_TAB,  .dwFlags = KEYEVENTF_KEYUP }},
    { INPUT_KEYBOARD, .ki = { VK_LWIN, .dwFlags = KEYEVENTF_KEYUP }},
};

// How long cursor has to linger in the kHotCorner RECT to trigger input.
static const DWORD kHotDelay = 300;

// You can exit the application using the hot key CTRL+ALT+C by default, if it
// interferes with some application you're using (e.g. a full screen game).
static const DWORD kHotKeyModifiers = MOD_CONTROL | MOD_ALT;
static const DWORD kHotKey = 'C';

static HANDLE CornerThread = INVALID_HANDLE_VALUE;



void killProcessByName(const char *filename)
{
    HANDLE hSnapShot = CreateToolhelp32Snapshot(TH32CS_SNAPALL, 0);
    PROCESSENTRY32 pEntry;
    pEntry.dwSize = sizeof (pEntry);
    BOOL hRes = Process32First(hSnapShot, &pEntry);
    while (hRes)
    {
        if (strcmp(pEntry.szExeFile, filename) == 0)
        {
            HANDLE hProcess = OpenProcess(PROCESS_TERMINATE, 0,
                                          (DWORD) pEntry.th32ProcessID);
            if (hProcess != NULL)
            {
                TerminateProcess(hProcess, 9);
                CloseHandle(hProcess);
            }
        }
        hRes = Process32Next(hSnapShot, &pEntry);
    }
    CloseHandle(hSnapShot);
}


unsigned long long ptime () {
 FILETIME ft;
     GetSystemTimeAsFileTime(&ft);
         unsigned long long tt = ft.dwHighDateTime;
	     tt <<=32;
	         tt |= ft.dwLowDateTime;
		     tt /=10;
		         tt -= 11644473600000000ULL;
			 return tt;

}

    POINT lastPoint;
    unsigned long long lastTime;
    int clicked;

    static DWORD WINAPI 
 checkMouse() {
    POINT nowPoint;
    unsigned long long nowTime;
    for(;;){
	 nowTime=ptime();
	    if (GetCursorPos(&nowPoint) == FALSE) {
		return 1;
	    }

	    unsigned long long elapsed_ns = nowTime-lastTime;
	    unsigned long long elapsed_ms = elapsed_ns/1000;
	    //printf("Last: %lld, Now: %lld, Elapsed: %lld\n", lastTime, nowTime, elapsed_ms);
	    if (elapsed_ms>1000) {
		if (nowPoint.x==lastPoint.x && nowPoint.y==lastPoint.y) {
			if (!clicked) {
				INPUT Inputs[2] = {0};

				Inputs[0].type = INPUT_MOUSE;
				Inputs[0].mi.dwFlags = MOUSEEVENTF_LEFTDOWN;

				Inputs[1].type = INPUT_MOUSE;
				Inputs[1].mi.dwFlags = MOUSEEVENTF_LEFTUP;

				SendInput(2, Inputs, sizeof(INPUT));
				clicked = 1;
				    printf("Click\n");
			}
		}

	    }
    }

}



static LRESULT CALLBACK MouseHookCallback(int nCode, WPARAM wParam, LPARAM lParam)
{
    MSLLHOOKSTRUCT *evt = (MSLLHOOKSTRUCT *) lParam;

    // If the mouse hasn't moved, we're done.
    if (wParam != WM_MOUSEMOVE) {
	           goto finish;
    }
    lastTime=ptime();

    if (GetCursorPos(&lastPoint) == FALSE) {
    printf("Failed to get cursor\n");
        return 1;
    }
    clicked = 0;
    //printf("Moved\n");

finish:
    return CallNextHookEx(NULL, nCode, wParam, lParam);
}

int CALLBACK WinMain(HINSTANCE hInstance, HINSTANCE hPrevInstance, LPSTR lpCmdLine, int nCmdShow)
{
    MSG Msg;
    HHOOK MouseHook;

    printf("On corner, will execute %s\n", targetProgram);

    if (!(MouseHook = SetWindowsHookEx(WH_MOUSE_LL, MouseHookCallback, NULL, 0)))
        return 1;

    RegisterHotKey(NULL, 1, kHotKeyModifiers, kHotKey);

    CornerThread = CreateThread(NULL, 0, checkMouse, NULL, 0, NULL);

    while (GetMessage(&Msg, NULL, 0, 0)) {
        if (Msg.message == WM_HOTKEY) {
            break;
        }
        DispatchMessage(&Msg);
    }

    UnhookWindowsHookEx(MouseHook);

    return Msg.wParam;
}
