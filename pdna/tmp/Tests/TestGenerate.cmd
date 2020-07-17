@echo off
:::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
:: 
:: IMPORTANT NOTICE
:: ================
:: 
::   Copyright (C) 2016, Microsoft Corporation
::   All Rights Reserved.
:: 
::   THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS STRICTLY UNDER THE TERMS
::   OF A SEPARATE LICENSE AGREEMENT.  NO SOURCE CODE DISTRIBUTION OR SOURCE 
::   CODE DISCLOSURE RIGHTS ARE PROVIDED WITH RESPECT TO THIS SOFTWARE OR ANY 
::   DERIVATIVE WORKS THEREOF.  USE AND NON-SOURCE CODE REDISTRIBUTION OF THE 
::   SOFTWARE IS PERMITTED ONLY UNDER THE TERMS OF SUCH LICENSE AGREEMENT.  
::   THIS SOFTWARE AND ANY WORKS/MODIFICATIONS/TRANSLATIONS DERIVED FROM THIS 
::   SOFTWARE ARE COVERED BY SUCH LICENSE AGREEMENT AND MUST ALSO CONTAIN THIS 
::   NOTICE.  THIS SOFTWARE IS CONFIDENTIAL AND SENSITIVE AND MUST BE SECURED 
::   WITH LIMITED ACCESS PURSUANT TO THE TERMS OF SUCH LICENSE AGREEMENT.
:: 
::   THE SOFTWARE IS PROVIDED "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, 
::   INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY 
::   AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE
::   COPYRIGHT HOLDER BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, 
::   EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, 
::   PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS;
::   OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, 
::   WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR 
::   OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF 
::   ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
::   
::   Project:       PhotoDNA Robust Hashing
::   File:          TestGenerate.cmd
::   Description:   Tool to test GenerateHashes sample
::
::   History:   2015.11.20   adrian chandley  Created test. 
::  
:::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::

::
:: TestGenerate [executable to test]
::
setlocal ENABLEDELAYEDEXPANSION
set _res=0
set ERROR=

set _pr=%~dpnx1
if not exist "%_pr%" set ERROR=Program not found&goto :ERROR

:: get last modified time for the executable (region independent)
set fdt=
for /F "usebackq tokens=1,2 delims==" %%i in (`wmic datafile where name^=^"%_pr:\=\\%^" get LastModified /VALUE 2^>nul`) do if "%%i"=="LastModified" set fdt=%%j
:: display the name and last modified time for the executable
if defined fdt echo TESTING: %~1 (File time: %fdt:~0,4%.%fdt:~4,2%.%fdt:~6,2% %fdt:~8,2%:%fdt:~10,2%)
if not defined fdt echo TESTING: %~1

pushd "%~dp0"

:: Get the base test hash
call :base

:: Calculate hashes and distance from base
call :Calculate "%_pr%" "Test01.jpg" _distance "GEQ 1000"
if %_res% GTR 0 set ERROR=%ERROR% Distance too great.&goto :ERROR
call :Calculate "%_pr%" "Test02.jpg" _distance "LSS 50000"
if %_res% GTR 0 set ERROR=%ERROR% Distance too small.&goto :ERROR
call :Calculate "%_pr%" "Test03.jpg" _distance "GEQ 1000"
if %_res% GTR 0 set ERROR=%ERROR% Distance too great.&goto :ERROR
call :Calculate "%_pr%" "Test04.png" _distance "GEQ 1000"
if %_res% GTR 0 set ERROR=%ERROR% Distance too great.&goto :ERROR

:CLEANUP
FOR /L %%v IN (1,1,144) DO set B_%%v=
:FINISH
POPD
exit /b %_res%
goto :EOF
:ERROR
echo ERROR %ERROR%: %1
set _res=2
goto :FINISH

::
:: calculate [executable] [image file] [result variable name] [condition]
::   
::   Uses the executable to generate a hash for the given image
::   The resulting hash will be compared to the the base test hash using the given condition
::
::   If the condition is not met (e.g. GEQ 1000), the executable has failed the test
::
:calculate
set _pcmd="%~1"
set ERROR=%~2:
set /a %~3=-1

if /i "%~x1"==".py" set _pcmd=python "%~1"
if /i "%~x1"==".java" set _pcmd=java -cp "%~dp1\." "%~n1"
if /i "%~x1"==".class" set _pcmd=java -cp "%~dp1\." "%~n1"

:: call executable to calculate hash, use :parse to unpack the result into variables D_(1..144)
for /f "usebackq delims=^!" %%d in (`"%_pcmd% "%~2""`) do call :parse D %%d
if not defined D_144 set ERROR=%ERROR% Failed to calculate hash&goto :EOF
set _pcmd=

:: compare to the base hash
set /a dist=0
FOR /L %%v IN (1,1,144) DO (
  set /A diff=!B_%%v!-!D_%%v!
  set /a dist+= !diff! * !diff!
  set D_%%v=
)
if %dist% %~4 set _res=1&echo FAIL: Distance for %~2: %dist%
if NOT %dist% %~4 echo PASS: Distance for %~2: %dist%
set %~3=%dist%
set dist=
goto :EOF

:: set the base hash and parse into B_(1..144)
:base
set BASE=B,5,70,1,78,27,18,8,58,14,24,16,55,15,28,63,51,32,38,57,68,42,14,48,102,1,53,214,42,6,34,209,64,4,22,178,81,6,15,146,76,20,10,127,61,45,0,128,57,5,47,255,6,1,57,255,6,0,52,255,13,21,28,255,13,51,14,255,5,77,4,255,14,13,41,17,92,24,33,29,88,29,52,71,31,60,18,91,26,75,2,69,12,45,17,63,32,28,86,67,44,124,18,62,46,56,97,75,69,64,46,58,108,101,17,32,66,54,22,36,65,62,53,88,31,32,80,44,73,81,59,42,227,81,63,133,101,128,43,85,68,88,15,183,21
call :parse B %BASE%
goto :EOF

:: parse [csv string] [Prefix]
::
:: Parse a comma seperated string into individual variables Prefix_0 Prefix_1 ... 
:: 
:: Implementation note: this is a recursive method to overcome the limited number 
:: of tokens that can be unpacked by the windows FOR command
::
:parse
SET NAME=%~1
set DATA=%*
set DATA=%DATA:~2%
set /a count=0
:: unpack the first value and recurse through the remainder
FOR /F "tokens=1* delims=," %%i IN ("%DATA%") DO (
    set !NAME!_Name=%%i
    CALL :parsedatatokens2 "%%j"
)
goto :EOF
:: unpack the next value and recurse through the remainder
:: note that the use of /A forces a numeric assignment
:parsedatatokens2
set /a count+=1
FOR /F "tokens=1* delims=," %%i IN ("%~1") DO (
    set /A %NAME%_!count!=%%i
    CALL :parsedatatokens2 "%%j"
    )
goto :EOF
