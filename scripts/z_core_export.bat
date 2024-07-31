@echo off
@setlocal
:: This is the core function on exporting the logs
:: Please define following varaibles before calling this script:
::
:: log_folder_prefix  - the folder prefix
:: profile             - (optional) whether to use the aws profile
:: region              - region of the log file stream
:: log_group           - log group of the stream
:: log_stream          - log stream
:: log_stream_prefixes - log stream prefix
:: from_date           - start date
:: duration            - duration
:: log_timestamp       - set value to "-t" to print the timestamp in each line
:: -------------------------------------------------------------------------------

:: set following to prevent the exception - 'charmap' codec can't encode character
:: ref: https://github.com/aws/aws-cli/issues/4778
set PYTHONUTF8=1

:: un-comment following to change to parent path if awsExportLog.exe is located in parent path
set script_dir=%~dp0
cd %script_dir%\..

:: define the base parameters
set base_param=-r=%region% -g=%log_group% -d=%duration% %log_timestamp%

set log_folder_suffix=_now
if not "%from_date%" equ "" (
  set log_folder_suffix=%from_date:~0,13%
  set base_param=-f=%from_date% %base_param%
)

:: set the log folder
set log_folder=%log_folder%\%log_folder_suffix%

mkdir %log_folder% >NUL 2>NUL

set cmd_test_identity=aws sts get-caller-identity
if not "%profile%" equ "" (
    set profile_param=-p=%profile%
    set cmd_test_identity=%cmd_test_identity% --profile %profile%
)

%cmd_test_identity% > NUL 2>&1
if ERRORLEVEL 1 (
  echo Error: Could not retrieve caller identity. Attempting SSO configuration.
  aws configure sso
) else (
  rem echo Success: AWS credentials are available.
)

:: loop to export each log stream
for %%i in (%log_streams%) do (
    call :exportStreamLog %%i
)

:: loop to export each log stream prefix
for %%i in (%log_stream_prefixes%) do (
    call :exportStreamPrefixLog %%i
)

echo done
goto :eof

:exportStreamLog
set "stream=%~1"
  rem echo Exporting %log_folder%\%stream%.log
  rem echo awsExportLog.exe %profile_param% %base_param% -s=%%i ^> %log_folder%\%%i.log
  awsExportLog.exe %profile_param% %base_param% -s=%stream% -o=%log_folder%
  rem awsExportLog.exe %profile_param% %base_param% -s=%stream% > %log_folder%\%stream%.log
goto :eof


:exportStreamPrefixLog
set "streamPrefix=%~1"
  rem echo Exporting %log_folder%\%stream%.log
  rem echo awsExportLog.exe %profile_param% %base_param% -sp=%stream% -o=%log_folder%
  awsExportLog.exe %profile_param% %base_param% -sp=%streamPrefix% -o=%log_folder%
goto :eof

