@echo off
echo Tidy modules...
go mod tidy
echo Installing...
go install .
if %errorlevel% neq 0 (
    echo Installation failed. Trying local build...
    go build -o gr.exe
    echo Built local gr.exe.
    echo To use 'gr' globally, add this folder or %%USERPROFILE%%\go\bin to your PATH.
) else (
    echo Installed 'gr' to %%USERPROFILE%%\go\bin
    echo Ensure %%USERPROFILE%%\go\bin is in your PATH.
    echo Try opening a new terminal and running 'gr'.
)
pause
