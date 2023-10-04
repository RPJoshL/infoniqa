@ECHO OFF

:: Bypass the "Terminate Batch Job" prompt
if "%~1"=="-FIXED_CTRL_C" (
   :: Remove the -FIXED_CTRL_C parameter
   SHIFT
) ELSE (
   :: Run the batch with <NUL and -FIXED_CTRL_C
   CALL <NUL %0 -FIXED_CTRL_C %*
   GOTO :EOF
)

SET PATH=%PATH%;C:\Windows\System32

:: Custom go temp path
if exist .\scripts\temp_path (
   set /p tempDir=< scripts\temp_path
   set GOTMPDIR=%tempDir%
)

set args=%1
shift
:start
if [%1] == [] goto done
set args=%args% %1
shift
goto start
:done

:: Debug settings
set LOGGER_LEVEL=DEBUG
set INFONIQA_CONFIG=./config-dev.yaml

nodemon --delay 1s -e go,html,yaml --signal SIGKILL --ignore web/app/ --quiet ^
--exec "echo [Restarting] && go run ./cmd/infoniqa" -- %args% || "exit 1"