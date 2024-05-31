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
	VERSION   = "1.1"
	ERROR_MSG = "ERROR: %s\n"
)

var (
	profile       string
	region        string
	logEventInput *cloudwatchlogs.GetLogEventsInput
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
	flag.StringVar(&from, "f", "", "From time in RFC3339 format. e.g.: 2024-02-13T14:25:60Z")
	flag.StringVar(&duration, "d", "1h", "(Optional) Duration of the log to be taken from the From time. \n"+
		"Valid time units are \"ns\", \"us\" (or \"µs\"), \"ms\", \"s\", \"m\", \"h\"")
	flag.StringVar(&profile, "p", "", "(Optional) Profile")
	flag.BoolVar(&help, "h", false, "Help")
	flag.Parse()

	if help {
		printHelp()
		os.Exit(0)
	}

	if (region == "") || (logGroupName == "") || (logStreamName == "") || (from == "") || (duration == "") {
		printHelp()
		fmt.Fprintf(os.Stderr, ERROR_MSG, fmt.Errorf("missing parameters"))
		os.Exit(1)
	}

	timeFrom, err := time.Parse(time.RFC3339, from)
	if err != nil {
		fmt.Fprintf(os.Stderr, ERROR_MSG, fmt.Errorf("error on parsing the From time: %s", err))
		os.Exit(2)
	}

	d, err := time.ParseDuration(duration)
	if err != nil {
		fmt.Fprintf(os.Stderr, ERROR_MSG, fmt.Errorf("error on parsing the Duration: %s", err))
		os.Exit(3)
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
		resp, err := client.GetLogEvents(context.TODO(), logEventInput)
		if err != nil {
			fmt.Fprintf(os.Stderr, ERROR_MSG, err)
			return
		}

		for _, event := range resp.Events {
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
