# awsExportLog
Rretrieve the logs from the AWS CloudWatch Logs

## Parameters
```
  -d string
        (Optional) Duration of the log to be taken from the From time.
        Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h" (default "1h")
  -f string
        From time in RFC3339 format. e.g.: 2024-02-13T14:25:60Z
  -g string
        AWS log group name
  -h    Help
  -p string
        (Optional) Profile
  -r string
        AWS region
  -s string
        AWS log stream name
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
