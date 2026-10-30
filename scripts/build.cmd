@echo off
setlocal EnableExtensions

rem PowerShell and cmd do not run .sh files. Windows asks which program
rem should open them, and the build never starts. This launcher runs the
rem same script with Git Bash, which can see the Windows Go and Node installs.

set "BASH="

for %%K in ("HKCU\SOFTWARE\GitForWindows" "HKLM\SOFTWARE\GitForWindows") do (
  if not defined BASH (
    for /f "tokens=2,*" %%A in ('reg query %%K /v InstallPath 2^>nul') do (
      if exist "%%B\bin\bash.exe" set "BASH=%%B\bin\bash.exe"
    )
  )
)

if not defined BASH if exist "%ProgramFiles%\Git\bin\bash.exe" set "BASH=%ProgramFiles%\Git\bin\bash.exe"
if not defined BASH if exist "%LocalAppData%\Programs\Git\bin\bash.exe" set "BASH=%LocalAppData%\Programs\Git\bin\bash.exe"
if not defined BASH if exist "%USERPROFILE%\scoop\apps\git\current\bin\bash.exe" set "BASH=%USERPROFILE%\scoop\apps\git\current\bin\bash.exe"

if not defined BASH (
  echo error: Git Bash was not found. 1>&2
  echo Install Git for Windows from https://git-scm.com/download/win 1>&2
  echo Then run: scripts\build.cmd build 1>&2
  exit /b 1
)

"%BASH%" --noprofile --norc "%~dp0build.sh" %*
exit /b %ERRORLEVEL%
