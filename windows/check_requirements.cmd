@echo off

:--------------------------------------
set missing_req=0

echo.
echo Checking system requirements...
echo.

:--------------------------------------
:GIT
git -v 1>nul 2>nul
IF ERRORLEVEL 1 GOTO MISSING_GIT
echo git     OK
GOTO NPM

:MISSING_GIT
set missing_req=1
echo git     MISSING
GOTO NPM

:--------------------------------------
:NPM
call npm -v 1>nul 2>nul
IF ERRORLEVEL 1 GOTO MISSING_NPM
echo npm     OK
GOTO PNPM

:MISSING_NPM
set missing_req=1
echo npm     MISSING
GOTO PNPM

:--------------------------------------
:PNPM
call pnpm -v 1>nul 2>nul
IF ERRORLEVEL 1 GOTO MISSING_PNPM
echo pnpm    OK
GOTO NODE

:MISSING_PNPM
set missing_req=1
echo pnpm    MISSING
GOTO NODE

:--------------------------------------
:NODE
call node -v 1>nul 2>nul
IF ERRORLEVEL 1 GOTO MISSING_NODE
echo node    OK
GOTO GCC

:MISSING_NODE
set missing_req=1
echo node    MISSING
GOTO GCC

:--------------------------------------
:GCC
call gcc -v 1>nul 2>nul
IF ERRORLEVEL 1 GOTO MISSING_GCC
echo gcc     OK
GOTO MSBUILD

:MISSING_GCC
set missing_req=1
echo gcc     MISSING
GOTO MSBUILD

:--------------------------------------
:MSBUILD
call msbuild -h 1>nul 2>nul
IF ERRORLEVEL 1 GOTO MISSING_MSBUILD
echo MSBuild OK
GOTO JSIGN

:MISSING_MSBUILD
set missing_req=1
echo MSBuild MISSING
GOTO JSIGN

:--------------------------------------
:JSIGN
call jsign 1>nul 2>nul
IF ERRORLEVEL 1 GOTO MISSING_JSIGN
echo jsign   OK
GOTO END

:MISSING_JSIGN
set missing_req=1
echo jsign   MISSING
GOTO END

:--------------------------------------
:END
IF %missing_req% == 1 GOTO FAILED
echo.
echo INFO: All Windows build requirements pass.
echo.
GOTO EOF

:FAILED
echo.
echo FATAL: Windows build requirements missing.
echo.
exit /b 1

:EOF
