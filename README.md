# awsExportLog
GoLang progame to export the log from the AWS.

## Pre-requisite
1. AWS CLI must be installed

## Parameters
```
  -d string
        Duration of the log to be taken from the From time. e.g. 1m1s = 1 minute 1 second (default "1h")
  -f string
        From time in RFC3339 format. e.g.: 2024-02-13T14:25:60Z
  -g string
        AWS log group name
  -r string
        AWS region
  -s string
        AWS log stream name
```

## Example Usage
```
awsExportLog -r=us-west-2 -g=/aws/containerinsights/my-log-group s=mytest.restapi -f=2024-01-13T14:25:00Z -d=1h > myLogfile.log
```
This will export the log content to the myLogfile.log


