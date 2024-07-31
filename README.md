# awsExportLog
Rretrieve the logs from the AWS CloudWatch Logs

## Parameters
```
Version: 1.4
Usage: awsExportLog.exe -r REGION -g GROUP -s STREAM [options]
Retrieve the log content from the AWS CloudWatch

awsExportLog.exe -r REGION -g GROUP -s STREAM -f FROM_TIME
Retrieve the log content starting from FROM_TIME for previous 1 hour

awsExportLog.exe -r REGION -g GROUP -sp STREAM_PREFIX -o OUTPUT_FOLDER
Search the matched stream and export the log content to the OUTPUT_FOLDER with filename is $streamName.log for previous 1 hour

Parameters
  -d string
        (Optional) Duration of the log to be taken from the From time.
        Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h" (default "1h")
  -f string
        (Optional) From time in RFC3339 format. e.g.: 2024-02-13T14:25:60Z. (default current time)
  -g string
        AWS log group name
  -h    Help
  -o string
        (Optional) Output folder for the log file(s). Filename of the log file is $b.log
  -p string
        (Optional) Profile
  -r string
        AWS region
  -s string
        AWS log stream name
        Either Stream Name or Stream Prefix should has value
  -sp string
        (Optional) AWS Log stream prefix
        Either Stream Prefix or Stream Name should has value
  -t    (Optional) Whether to write the log entry received timestamp (UTC) to the log
```

## Example Usage
```
awsExportLog -r=us-west-2 -g=/aws/containerinsights/my-log-group -s=mytest.restapi -f=2024-01-13T14:25:00Z -d=1h > myLogfile.log
```
This will export the log content to the myLogfile.log

## Pre-requisite for development
Install the following libraries:
```
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs
```
