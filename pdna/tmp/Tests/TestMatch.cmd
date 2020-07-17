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
::   File:          TestMatch.cmd
::   Description:   Tool to test MatchHashes sample
::
::   History:   2015.11.20   adrian chandley  Created test. 
::  
:::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::

::
:: TestMatch [executable to test]
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

:: compare sets of hashes looking for expected results
call :match "%_pr%" Test1.csv Test.csv Test03.jpg
if not defined _match set ERROR=%ERROR% Failed to match.&goto :ERROR
call :match "%_pr%" Test2.csv Test.csv Test04.png
if not defined _match set ERROR=%ERROR% Failed to match.&goto :ERROR

:FINISH
POPD
exit /b %_res%
goto :EOF
:ERROR
echo ERROR %ERROR%: %1
set _res=2
goto :FINISH

::
:: match [executable] [test candidate hashes] [test hash database] [Image name]
:: 
::   Uses the executable to compare the candidate hashes against the hash database
:: 
::   If a match is not found for [Image Name], the executable has failed the test 
::
:match
set _match=
set /a _ans=0
set ERROR=%~2:
set _pcmd=
set _cline="%~1" "%~2" "%~3" "" test
if /i "%~x1"==".java" goto :java
if /i "%~x1"==".class" goto :java
if /i "%~x1"==".py" set _pcmd=python
goto :runmatch
:java
set _pcmd=java
set _cline=-cp "%~dp1\." "%~n1" "%~2" "%~3" "" test
:runmatch
for /f "usebackq delims=^!" %%d in (`"%_pcmd% %_cline%" 2^>nul`) do call :parse %%d
if "%_ans%"=="0" set _match=&set ERROR=%ERROR% Failed to find match"
if %_ans% GTR 1 set _match=&set ERROR=%ERROR% match error"
if /i not "%_match%"=="%~4" echo FAIL: Match incorrect (%_match%)&goto :fail
echo PASS: Correct match found (%_match%)
goto :EOF
:fail
set _match=
goto :EOF

:: parse a line of output from the executable to collect the name and match distance
:parse
set __p=%~1
if /i not "%__p:~0,5%"=="Test0" goto :EOF
set /a _dist=%~2
if %_dist% GEQ 1000 goto :EOF
set _match=%~1
set /a _ans+=1
goto :EOF
