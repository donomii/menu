#ifndef KEYWATCHER
#define KEYWATCHER 1
#include <Windows.h>
#include <stdio.h>
#include <fileapi.h>
#include <process.h>
#include <Tlhelp32.h>
#include <winbase.h>
#include <string.h>
//#define _WIN32_WINNT 0x050


int menu_active = 0;
char * pidfile;

 void HandleKey(int i);


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
           
            
            
		if ( (wParam == WM_KEYUP) || (wParam == WM_SYSKEYUP) ) // Keyup
			                {
                                 printf("c: %i\n", p->vkCode);
                                HandleKey(p->vkCode);
		  //fEatKeystroke=1;
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

char * pidPath() {
	/*
	printf("USERPROFILE = %s\n", getenv("USERPROFILE"));
    printf("HOMEDRIVE   = %s\n", getenv("HOMEDRIVE"));
    printf("HOMEPATH    = %s\n", getenv("HOMEPATH"));
	*/
	
	char * pidfile = calloc(1024,1);
	snprintf(pidfile, 1023, "%s/universalmenu.pid", getenv("USERPROFILE"));
	return pidfile;
}

int watchKeys( )
{
	pidfile = pidPath();
	
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
#endif
