@echo off
setlocal

REM goany Windows Build Script
REM Requires: Go 1.24+, MinGW-w64 (g++), .NET 9 SDK, Rust, Node.js

echo ========================================
echo goany Windows Build Script
echo ========================================
echo.

REM Check prerequisites
echo Checking prerequisites...
where go >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo ERROR: Go not found. Install from https://go.dev/dl/
    exit /b 1
)

where g++ >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo ERROR: g++ not found. Install MinGW-w64 from https://github.com/niXman/mingw-builds-binaries/releases
    exit /b 1
)

where dotnet >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo WARNING: dotnet not found. C# backend tests will fail.
    echo Install from https://dotnet.microsoft.com/download/dotnet/9.0
)

where rustc >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo WARNING: rustc not found. Rust backend tests will fail.
    echo Install from https://rustup.rs/
)

where node >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo WARNING: node not found. JavaScript backend tests will fail.
    echo Install from https://nodejs.org/
)

echo.
echo Prerequisites check complete.
echo.

REM Build astyle library
echo ========================================
echo Building astyle library...
echo ========================================
cd compiler\astyle

if exist libastyle.a del libastyle.a
if exist *.o del *.o

g++ -c -Wall -O2 -DASTYLE_LIB -std=c++17 ASBeautifier.cpp -o ASBeautifier.o
if %ERRORLEVEL% NEQ 0 goto :error

g++ -c -Wall -O2 -DASTYLE_LIB -std=c++17 ASEnhancer.cpp -o ASEnhancer.o
if %ERRORLEVEL% NEQ 0 goto :error

g++ -c -Wall -O2 -DASTYLE_LIB -std=c++17 ASFormatter.cpp -o ASFormatter.o
if %ERRORLEVEL% NEQ 0 goto :error

g++ -c -Wall -O2 -DASTYLE_LIB -std=c++17 ASResource.cpp -o ASResource.o
if %ERRORLEVEL% NEQ 0 goto :error

g++ -c -Wall -O2 -DASTYLE_LIB -std=c++17 astyle_main.cpp -o astyle_main.o
if %ERRORLEVEL% NEQ 0 goto :error

ar rcs libastyle.a ASBeautifier.o ASEnhancer.o ASFormatter.o ASResource.o astyle_main.o
if %ERRORLEVEL% NEQ 0 goto :error

echo astyle library built successfully.
cd ..\..

REM Build goany
echo.
echo ========================================
echo Building goany...
echo ========================================
cd cmd

set CGO_ENABLED=1
set CC=gcc
set CXX=g++

go build -o goany.exe .
if %ERRORLEVEL% NEQ 0 goto :error

echo goany built successfully.
cd ..

echo.
echo ========================================
echo Build complete!
echo ========================================
echo.
echo Run tests with: cd cmd ^&^& go test -v ./...
echo.
cmd\goany.exe --help

exit /b 0

:error
echo.
echo ========================================
echo BUILD FAILED
echo ========================================
exit /b 1
