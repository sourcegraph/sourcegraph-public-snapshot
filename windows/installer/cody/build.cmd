@echo off

cd windows\installer\cody

:: TODO(william): Make this call based on branch
GOTO TESTCERT
::GOTO REALCERT

:TESTCERT
call %SYSTEMDRIVE%\sign\testcert
powershell -ExecutionPolicy Unrestricted .\build.ps1
GOTO EOF

:REALCERT
call powershell -ExecutionPolicy Unrestricted %STSTEMDRIVE%\sign\realcert.ps1
GOTO EOF

:EOF

cd ..\..\..