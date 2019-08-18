#include <Windows.h>
#include <stdio.h>
#include <fileapi.h>
#include <process.h>
#include <Tlhelp32.h>
#include <winbase.h>
#include <string.h>
//#define _WIN32_WINNT 0x050

int menu_active = 0;


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

LRESULT CALLBACK LowLevelKeyboardProc(int nCode, WPARAM wParam, LPARAM lParam)
{
    BOOL fEatKeystroke = FALSE;
	PKBDLLHOOKSTRUCT p;

    if (nCode == HC_ACTION)
    {
        switch (wParam)
        {
        case WM_KEYDOWN:
        case WM_SYSKEYDOWN:
        case WM_KEYUP:
        case WM_SYSKEYUP:
             p = (PKBDLLHOOKSTRUCT)lParam;
            printf("%i\n", p->vkCode);
            if (fEatKeystroke = (p->vkCode == 20)) {
		if ( (wParam == WM_KEYUP) || (wParam == WM_SYSKEYUP) ) // Keyup
			                {
		    printf("Menu active: %i\n", menu_active);
		    if (menu_active==1) {
			    menu_active = 0;
				killProcessByName("menu.exe");

		    } else {
			    //system("menu.exe");
			    menu_active=1;
			    ShellExecute(NULL, "open", "universal_menu_main.exe", NULL, NULL, SW_SHOWDEFAULT);
		    }

	    }
	    }
            /*if (fEatKeystroke = (p->vkCode == 0x41)) {     //redirect a to b
            keybd_event('B', 0, 0, 0);
            keybd_event('B', 0, KEYEVENTF_KEYUP, 0);
            break;
            }
	    */
            break;
        }
    }
    return(fEatKeystroke ? 1 : CallNextHookEx(NULL, nCode, wParam, lParam));
}

int CheckOneInstance()
{

    HANDLE  m_hStartEvent = CreateEventW( NULL, FALSE, FALSE, L"Global\\CSAPP" );

    if(m_hStartEvent == NULL)
    {
    CloseHandle( m_hStartEvent ); 
        return 0;
    }


    if ( GetLastError() == ERROR_ALREADY_EXISTS ) {

        CloseHandle( m_hStartEvent ); 
        m_hStartEvent = NULL;
        // already exist
        // send message from here to existing copy of the application
        return 0;
    }
    // the only instance, start in a usual way
    return 1;
}

int main()
{
	if (!CheckOneInstance()) {
		exit(1);
		abort();
	}
    // Install the low-level keyboard & mouse hooks
    HHOOK hhkLowLevelKybd = SetWindowsHookEx(WH_KEYBOARD_LL, LowLevelKeyboardProc, 0, 0);

    // Keep this app running until we're told to stop
    MSG msg;
    while (!GetMessage(&msg, NULL, 0, 0)) {    //this while loop keeps the hook
        TranslateMessage(&msg);
        DispatchMessage(&msg);
    }

    UnhookWindowsHookEx(hhkLowLevelKybd);

    return(0);
}
