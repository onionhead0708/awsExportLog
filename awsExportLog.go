package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
)

const (
	VERSION   = "1.3"
	ERROR_MSG = "ERROR: %s\n"
)

var (
	profile       string
	region        string
	logEventInput *cloudwatchlogs.GetLogEventsInput
	logTimestamp  bool
)

func main() {
	readConfigFromCommandLine()
	retrieveAwsLog()
}

func readConfigFromCommandLine() {

	var logGroupName, logStreamName, from, duration string
	var help bool

	flag.StringVar(&region, "r", "", "AWS region")
	flag.StringVar(&logGroupName, "g", "", "AWS log group name")
	flag.StringVar(&logStreamName, "s", "", "AWS log stream name")
	flag.StringVar(&from, "f", "", "(Optional) From time in RFC3339 format. e.g.: 2024-02-13T14:25:60Z. By default is current time.")
	flag.StringVar(&duration, "d", "1h", "(Optional) Duration of the log to be taken from the From time. \n"+
		"Valid time units are \"ns\", \"us\" (or \"µs\"), \"ms\", \"s\", \"m\", \"h\"")
	flag.StringVar(&profile, "p", "", "(Optional) Profile")
	flag.BoolVar(&logTimestamp, "t", false, "(Optional) Whether to write the log entry received timestamp (UTC) to the log")
	flag.BoolVar(&help, "h", false, "Help")
	flag.Parse()

	if help {
		printHelp()
		os.Exit(0)
	}

	if (region == "") || (logGroupName == "") || (logStreamName == "") || (duration == "") {
		fmt.Fprintln(os.Stderr, "Region: ", region)
		fmt.Fprintln(os.Stderr, "Group: ", logGroupName)
		fmt.Fprintln(os.Stderr, "Stream: ", logStreamName)
		fmt.Fprintln(os.Stderr, "Duration: ", duration)
		printHelp()
		fmt.Fprintf(os.Stderr, ERROR_MSG, fmt.Errorf("missing parameters"))
		os.Exit(1)
	}

	d, err := time.ParseDuration(duration)
	if err != nil {
		fmt.Fprintf(os.Stderr, ERROR_MSG, fmt.Errorf("error on parsing the Duration: %s", err))
		os.Exit(3)
	}

	var timeFrom time.Time
	if from != "" {
		timeFrom1, err := time.Parse(time.RFC3339, from)
		if err != nil {
			fmt.Fprintf(os.Stderr, ERROR_MSG, fmt.Errorf("error on parsing the From time: %s", err))
			os.Exit(2)
		}
		timeFrom = timeFrom1
	} else {
		timeFrom = time.Now().Add(-d)
	}

	timeTo := timeFrom.Add(d)

	startTime := timeFrom.UnixMilli()
	endTime := timeTo.UnixMilli()

	startFromHeader := true

	//get log events
	logEventInput = &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  &logGroupName,
		LogStreamName: &logStreamName,
		StartTime:     &startTime,
		EndTime:       &endTime,
		StartFromHead: &startFromHeader, // Start from the beginning of the log stream
	}
}

func retrieveAwsLog() {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithDefaultRegion(region),
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, ERROR_MSG, err)
		fmt.Fprintf(os.Stderr, "Error when setting up credential\n")
	}
	//create cloudwatchlogs client
	client := cloudwatchlogs.NewFromConfig(cfg)

	for {
		resp, err := client.GetLogEvents(context.TODO(), logEventInput, func(o *cloudwatchlogs.Options) {
			o.Region = region
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, ERROR_MSG, err)
			return
		}

		timestampeLayout := "2006-01-02 15:04:05.000"
		for _, event := range resp.Events {
			if logTimestamp {
				timestamp := time.UnixMilli(int64(*event.Timestamp)).UTC()
				fmt.Print(timestamp.Format(timestampeLayout), " ")
			}
			fmt.Println(*event.Message)
		}

		if resp.NextForwardToken == nil ||
			(logEventInput.NextToken != nil && *logEventInput.NextToken == *resp.NextForwardToken) {
			// We've reached the end of the stream
			break
		}

		logEventInput.NextToken = resp.NextForwardToken
	}
}

func printHelp() {
	fmt.Println("Version:", VERSION)
	fmt.Println("Usage:", os.Args[0]+" -r REGION -g GROUP -s STREAM -f FROM_TIME [options]")
	fmt.Println("Rretrieve the logs from the AWS CloudWatch Logs")
	fmt.Println()
	fmt.Println("Parameters")
	flag.PrintDefaults()
}
