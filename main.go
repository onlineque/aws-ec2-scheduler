package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"log"
	"os"
	"strings"
	"time"
)

func Log(l *log.Logger, msg string) {
	l.SetPrefix(time.Now().Format("2006-01-02 15:04:05") + " ")
	l.Print(msg)
}

func getConfig(l *log.Logger) aws.Config {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		Log(l, err.Error())
		os.Exit(1)
	}
	return cfg
}

func assumeRole(roleArn string, l *log.Logger) aws.Config {
	cfg := getConfig(l)

	stsSvc := sts.NewFromConfig(cfg)
	creds := stscreds.NewAssumeRoleProvider(stsSvc, roleArn)

	cfg.Credentials = aws.NewCredentialsCache(creds)
	return cfg
}

func isActionNeeded(days string, startTime string, stopTime string, instanceState string, l *log.Logger) bool {
	// abbreviation of the day today
	today := time.Now().Weekday().String()[:3]
	// current time

	if !strings.Contains(days, today) {
		return false
	}

	now := time.Now()
	timeToConvert := fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())
	timeNow, err := time.Parse("15:04", timeToConvert)
	if err != nil {
		Log(l, err.Error())
		return false
	}
	timeStartTime, err := time.Parse("15:04", startTime)
	if err != nil {
		Log(l, err.Error())
		return false
	}
	timeStopTime, err := time.Parse("15:04", stopTime)
	if err != nil {
		Log(l, err.Error())
		return false
	}

	if instanceState == "running" && timeStopTime.Before(timeNow) {
		return true
	}
	if instanceState == "stopped" && timeStartTime.Before(timeNow) {
		return true
	}

	return false
}

func getTag(tags []types.Tag, key string) string {
	for _, tag := range tags {
		if *tag.Key == key {
			return *tag.Value
		}
	}
	return ""
}

func checkInstances(l *log.Logger, roleArn string, doAssume bool) {
	// in case the check needs to run cross-account, assume first the target role
	var cfg aws.Config

	if doAssume {
		cfg = assumeRole(roleArn, l)
	} else {
		cfg = getConfig(l)
	}

	stsSvc := sts.NewFromConfig(cfg)
	resultAssume, err := stsSvc.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
	if err != nil {
		Log(l, err.Error())
		os.Exit(4)
	}

	Log(l, fmt.Sprintf("Successfully assumed role: %s", *resultAssume.Arn))

	ec2Client := ec2.NewFromConfig(cfg)

	result, err := ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
	if err != nil {
		Log(l, err.Error())
		os.Exit(2)
	}

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			days := getTag(instance.Tags, "Day")
			startTime := getTag(instance.Tags, "StartTime")
			stopTime := getTag(instance.Tags, "StopTime")
			instanceState := string(instance.State.Name)
			instanceId := *instance.InstanceId

			actionNeeded := isActionNeeded(days, startTime, stopTime, instanceState, l)
			Log(l, fmt.Sprintf("Instance id: %s, name: %s, state: %s, scheduler days: %s, start time: %s, stop time: %s, action needed: %v",
				*instance.InstanceId,
				getTag(instance.Tags, "Name"),
				instanceState,
				days,
				startTime,
				stopTime,
				actionNeeded))

			if actionNeeded && instanceState == "running" {
				Log(l, "-> stopping the instance")
				_, err := ec2Client.StopInstances(context.TODO(), &ec2.StopInstancesInput{
					InstanceIds: []string{instanceId},
				})
				if err != nil {
					Log(l, err.Error())
				}
			}
			if actionNeeded && instanceState == "stopped" {
				Log(l, "-> starting the instance")
				_, err := ec2Client.StartInstances(context.TODO(), &ec2.StartInstancesInput{
					InstanceIds: []string{instanceId},
				})
				if err != nil {
					Log(l, err.Error())
				}
			}
		}
	}
}

func main() {
	l := log.New(os.Stdout, "", 0)

	// check the AWS account where the pod is running first
	checkInstances(l, "", false)

	readFile, err := os.Open("/config")
	if err != nil {
		Log(l, err.Error())
		os.Exit(5)
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		// check all AWS accounts supplied from a config
		roleToBeAssumed := fileScanner.Text()
		Log(l, fmt.Sprintf("Role to be assumed: %v", roleToBeAssumed))
		checkInstances(l, roleToBeAssumed, true)
	}
}
