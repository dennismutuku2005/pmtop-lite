@echo off
REM run_pmtop.bat — double-click this to open pmtop in a terminal window.
REM If pmtop crashes, you'll see the error before the window closes.

cd /d "%~dp0"
pmtop.exe
if errorlevel 1 (
    echo.
    echo pmtop exited with an error. See message above.
    pause
)
