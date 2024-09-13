# **Disclaimer:**  
This keylogger was developed strictly for educational and research purposes within the context of cybersecurity. The primary intent of this project is to study how keyloggers are built and how they function, in order to better understand the techniques used by malicious actors. This code is *not* intended for illegal or unethical use, and I do not condone or support the misuse of this project in any form. By accessing or using this repository, you agree to use the information responsibly and solely for educational purposes.

# **Liability:**  
I am not liable for any actions taken by individuals who download, modify, or use this repository. Any misuse, illegal activity, or unethical behavior conducted with this software is solely the responsibility of the user. It is your obligation to ensure that your use of this repository complies with all applicable laws and regulations.




## Overview

Skippy is a simple keylogger application for Windows, implemented in Go. It captures keystrokes and logs them to a file. This document explains the purpose of each function in Skippy and provides instructions for using the program.

## Functions

1. **`SetWindowsHookEx(idHook int, lpfn uintptr, hmod uintptr, dwThreadId uint32) HHOOK`**
   - **Purpose**: Sets a hook procedure to monitor keyboard events.
   - **Usage**: Called to install a low-level keyboard hook that allows Skippy to capture keystrokes.

2. **`CallNextHookEx(hhk HHOOK, nCode int, wParam WPARAM, lParam LPARAM) LRESULT`**
   - **Purpose**: Passes the hook information to the next hook procedure in the chain.
   - **Usage**: Invoked within the `LowLevelKeyboardProc` function to ensure other hooks and applications receive the keyboard events.

3. **`LowLevelKeyboardProc(nCode int, wParam WPARAM, lParam LPARAM) LRESULT`**
   - **Purpose**: Callback function that processes keyboard input events.
   - **Usage**: Handles key presses, adjusts for Shift and Caps Lock states, and writes keystroke information to the log file.

4. **`getKeyState(vkCode int) int`**
   - **Purpose**: Retrieves the state of a specific virtual key.
   - **Usage**: Used in `LowLevelKeyboardProc` to determine whether keys like Shift or Caps Lock are pressed.

5. **`transform(s string, shift bool) string`**
   - **Purpose**: Adjusts character output based on Shift and Caps Lock states.
   - **Usage**: Converts characters to their shifted or non-shifted counterparts and handles special symbols.

6. **`UnhookWindowsHookEx(hhk HHOOK) bool`**
   - **Purpose**: Removes the installed hook procedure.
   - **Usage**: Not used in the current implementation but would typically be called to clean up the hook procedure when it is no longer needed.

## How to Use Skippy

1. **Build the Program:**
   Normally when you build the program in windows it opens a terminal window on startup. To prevent this we build the program as a GUI application
     ```sh
    go build -ldflags="-H=windowsgui" -o myapp.exe main.go
     ```

 2. **Run Skippy**
  
     ```sh
     ./skippy.exe
     ```

 3. **Log File**
   - Skippy will create a file named `keyfile.txt` in the same directory where it is executed. This file will contain logs of all captured keystrokes, including timestamps and key details.
   you can change the name or location of the file and compress it then send it to your c2 or keep data in-memory and periodically send it your c2.

 4. **Stopping the Program**
   - To stop Skippy, close the application window or terminate the process from the task manager.
   - or add a killswitch

# Important Notes

- **Permissions**: Ensure that you have appropriate permissions to run and access files on your system.
- **Ethical Use**: Use Skippy responsibly and only on machines where you have explicit permission to monitor keystrokes. Unauthorized use of keyloggers is illegal and unethical.


