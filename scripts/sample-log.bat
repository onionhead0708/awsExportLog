@echo off
@setlocal

set log_folder=logs\sample-logs
rem set profile=my_profile
set region=ap-southeast-1
set log_group=/aws/my-group
set log_streams=%log_streams% my-application
rem set log_stream_prefixes=%log_stream_prefixes% my-app

set from_date=2024-06-04T19:30:00Z
set duration=1h

call z_core_export.bat
