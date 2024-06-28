# awsExportLog
Rretrieve the logs from the AWS CloudWatch Logs

## Parameters
```
Version: 1.3
Usage: awsExportLog.exe -r REGION -g GROUP -s STREAM -f FROM_TIME [options]
Rretrieve the logs from the AWS CloudWatch Logs

Parameters
  -d string
        (Optional) Duration of the log to be taken from the From time.
        Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h" (default "1h")
  -f string
        (Optional) From time in RFC3339 format. e.g.: 2024-02-13T14:25:60Z. By default is current time.
  -g string
        AWS log group name
  -h    Help
  -p string
        (Optional) Profile
  -r string
        AWS region
  -s string
        AWS log stream name
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
