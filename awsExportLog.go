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
	VERSION   = "1.4"
	ERROR_MSG = "ERROR: %s\n"
)

var (
	profile          string
	region           string
	logGroupName     string
	logStreamName    string
	logStreamPreix   string
	startTime        int64
	endTime          int64
	showLogTimestamp bool
	outfolder        string
)

func main() {
	readConfigFromCommandLine()
	client, err := createAwsClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, ERROR_MSG, err)
		os.Exit(1)
	}

	streams := []string{}
	// streams := make([]string, 0)

	// if the logStreamName ends with *, then get all the streams
	// and append them to the streams list
	if logStreamPreix != "" {
		mystreams := retrieveAwsStreams(client, logStreamPreix)
		// append streams to the list
		streams = append(streams, mystreams...)
		// streams = append(streams, mystreams...)
	} else {
		streams = append(streams, logStreamName)
	}

	for _, stream := range streams {
		retrieveAwsLog(client, stream)
	}
}

func readConfigFromCommandLine() {

	var from, duration string
	var help bool

	flag.StringVar(&region, "r", "", "AWS region")
	flag.StringVar(&logGroupName, "g", "", "AWS log group name")
	flag.StringVar(&logStreamName, "s", "", "AWS log stream name \n"+
		"Either Stream Name or Stream Prefix should has value")
	flag.StringVar(&logStreamPreix, "sp", "", "(Optional) AWS Log stream prefix \n"+
		"Either Stream Prefix or Stream Name should has value")
	flag.StringVar(&from, "f", "", "(Optional) From time in RFC3339 format. e.g.: 2024-02-13T14:25:60Z. (default current time)")
	flag.StringVar(&duration, "d", "1h", "(Optional) Duration of the log to be taken from the From time. \n"+
		"Valid time units are \"ns\", \"us\" (or \"µs\"), \"ms\", \"s\", \"m\", \"h\"")
	flag.StringVar(&profile, "p", "", "(Optional) Profile")
	flag.BoolVar(&showLogTimestamp, "t", false, "(Optional) Whether to write the log entry received timestamp (UTC) to the log")
	flag.BoolVar(&help, "h", false, "Help")
	flag.StringVar(&outfolder, "o", "", "(Optional) Output folder for the log file(s). Filename of the log file is $b.log")
	flag.Parse()

	if help {
		printHelp()
		os.Exit(0)
	}

	if (region == "") || (logGroupName == "") || (logStreamName == "" && logStreamPreix == "") || (duration == "") {
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
	startTime = timeFrom.UnixMilli()
	endTime = timeTo.UnixMilli()
}

func createAwsClient() (*cloudwatchlogs.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithDefaultRegion(region),
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		//fmt.Fprintf(os.Stderr, ERROR_MSG, err)
		//fmt.Fprintf(os.Stderr, "Error when setting up credential\n")
		return nil, err
	}

	//create cloudwatchlogs client
	client := cloudwatchlogs.NewFromConfig(cfg)
	return client, nil
}

func retrieveAwsStreams(client *cloudwatchlogs.Client, streamPrefix string) []string {

	fmt.Fprintln(os.Stderr, "Retrieving log streams for", streamPrefix)

	input := &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName:        &logGroupName,
		LogStreamNamePrefix: &streamPrefix,
	}

	var logStreams []string
	for {
		resp, err := client.DescribeLogStreams(context.TODO(), input, func(o *cloudwatchlogs.Options) {
			o.Region = region
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, ERROR_MSG, err)
			os.Exit(1)
		}

		for _, stream := range resp.LogStreams {
			if startTime < *stream.LastEventTimestamp && endTime > *stream.CreationTime {
				logStreams = append(logStreams, *stream.LogStreamName)
			}
		}

		if resp.NextToken == nil ||
			(input.NextToken != nil && *input.NextToken == *resp.NextToken) {
			// We've reached the end of the stream
			break
		}

		input.NextToken = resp.NextToken
	}

	return logStreams
}

func retrieveAwsLog(client *cloudwatchlogs.Client, streamName string) {

	//if outFolder is not empty, then write the log to a file
	file, shouldReturn := createFile(streamName)
	if shouldReturn {
		return
	}
	defer file.Close()

	startFromHeader := true

	//get log events
	logEventInput := &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  &logGroupName,
		LogStreamName: &streamName,
		StartTime:     &startTime,
		EndTime:       &endTime,
		StartFromHead: &startFromHeader, // Start from the beginning of the log stream
	}

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
			if showLogTimestamp {
				timestamp := time.UnixMilli(int64(*event.Timestamp)).UTC()
				writeLogEntry(file, timestamp.Format(timestampeLayout)+" ")
			}
			writeLogEntry(file, *event.Message+"\n")
		}

		if resp.NextForwardToken == nil ||
			(logEventInput.NextToken != nil && *logEventInput.NextToken == *resp.NextForwardToken) {
			// We've reached the end of the stream
			break
		}

		logEventInput.NextToken = resp.NextForwardToken
	}
}

func createFile(streamName string) (*os.File, bool) {
	var file *os.File
	var err error

	if outfolder != "" {
		fileName := outfolder + "/" + streamName + ".log"
		fmt.Fprintln(os.Stderr, "creating file: "+fileName)
		//create the file
		file, err = os.Create(fileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, ERROR_MSG, err)
			return nil, true
		}
	}
	return file, false
}

func writeLogEntry(file *os.File, logEntry string) {
	if file != nil {
		file.WriteString(logEntry)
	} else {
		fmt.Print(logEntry)
	}
}

func printHelp() {
	fmt.Println("Version:", VERSION)
	fmt.Println("Usage:", os.Args[0]+" -r REGION -g GROUP -s STREAM [options]")
	fmt.Println("Retrieve the log content from the AWS CloudWatch")
	fmt.Println()
	fmt.Println(os.Args[0]+" -r REGION -g GROUP -s STREAM -f FROM_TIME")
	fmt.Println("Retrieve the log content starting from FROM_TIME for previous 1 hour")
	fmt.Println()
	fmt.Println(os.Args[0]+" -r REGION -g GROUP -sp STREAM_PREFIX -o OUTPUT_FOLDER")
	fmt.Println("Search the matched stream and export the log content to the OUTPUT_FOLDER with filename is $streamName.log for previous 1 hour")
	fmt.Println()
	fmt.Println("Parameters")
	flag.PrintDefaults()
}
