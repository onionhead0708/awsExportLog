package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func readConfigFromCommandLine() ([]string, error) {

	var region, logGroupName, logStreamName, from, duration, profile string

	flag.StringVar(&region, "r", "", "AWS region")
	flag.StringVar(&logGroupName, "g", "", "AWS log group name")
	flag.StringVar(&logStreamName, "s", "", "AWS log stream name")
	flag.StringVar(&from, "f", "", "From time in RFC3339 format. e.g.: 2024-02-13T14:25:60Z")
	flag.StringVar(&duration, "d", "1h", "Duration of the log to be taken from the From time. e.g. 1m1s = 1 minute 1 second")
	flag.StringVar(&profile, "p", "", "Profile (Optional)")
	flag.Parse()

	if (region == "") || (logGroupName == "") || (logStreamName == "") || (from == "") || (duration == "") {
		flag.PrintDefaults()
		return nil, fmt.Errorf("missing parameters")
	}

	timeFrom, err := time.Parse(time.RFC3339, from)
	if err != nil {
		return nil, fmt.Errorf("error on parsing the From time: %s", err)
	}

	d, err := time.ParseDuration(duration)
	if err != nil {
		return nil, fmt.Errorf("error on parsing the Duration: %s", err)
	}

	timeTo := timeFrom.Add(d)

	var cmdParams []string
	cmdParams = append(cmdParams, "logs", "get-log-events",
		"--region", region,
		"--log-group-name", logGroupName,
		"--log-stream-name", logStreamName,
		"--start-time", fmt.Sprint(timeFrom.UnixMilli()),
		"--end-time", fmt.Sprint(timeTo.UnixMilli()),
		"--start-from-head")

	if profile != "" {
		cmdParams = append(cmdParams, "--profile", profile)
	}

	return cmdParams, nil
}

func retrieveAwsLog(cmdParams []string) {
	/*
		AWS Logs documentation:
		https://awscli.amazonaws.com/v2/documentation/api/latest/reference/logs/get-log-events.html
	*/
	out, err := exec.Command("aws", cmdParams...).Output()
	if err != nil {
		fmt.Println("ERROR when running following aws command:")
		fmt.Println("aws", strings.Join(cmdParams, " "))
		fmt.Println("Result:", err)
		fmt.Println("Check https://docs.aws.amazon.com/cli/latest/userguide/cli-usage-returncodes.html for the exit status code")
		return
	}

	//-- for debug the output
	// f, _ := os.OpenFile("output.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// f.Write(out)
	// f.Close()

	// Parse JSON response
	var res map[string]interface{}
	json.Unmarshal(out, &res)

	// Extract log messages
	events := res["events"].([]interface{})
	if len(events) == 0 {
		fmt.Println("----:: NO_MORE_LOG_FOR_THE_TIME_RANGE ::----->")
		return
	}

	for _, event := range events {
		m := event.(map[string]interface{})["message"]
		fmt.Println(m.(string))
	}

	// Get next token
	nextToken := res["nextForwardToken"].(string)
	// fmt.Println("nextToken:", nextToken)
	idx := -1
	for i, param := range cmdParams {
		if param == "--next-token" {
			idx = i
			break
		}
	}
	if idx == -1 {
		cmdParams = append(cmdParams, "--next-token", nextToken)
	} else {
		cmdParams = append(cmdParams[:idx+1], cmdParams[idx+2:]...)
		cmdParams = append(cmdParams, nextToken)
	}

	retrieveAwsLog(cmdParams)
}

func main() {
	cmdParams, err := readConfigFromCommandLine()
	if err != nil {
		fmt.Println(err)
		return
	}
	retrieveAwsLog(cmdParams)
}
