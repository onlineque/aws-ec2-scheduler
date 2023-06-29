package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"log"
	"strings"
	"time"
)

func getConfig() aws.Config {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}

func isActionNeeded(days string, startTime string, stopTime string, instanceState string) bool {
	// abbreviation of the day today
	today := time.Now().Weekday().String()[:3]
	// current time

	if !strings.Contains(days, today) {
		return false
	}

	now := time.Now()
	timeToConvert := fmt.Sprintf("%v:%v", now.Hour(), now.Minute())
	timeNow, err := time.Parse("15:04", timeToConvert)
	if err != nil {
		log.Fatal(err)
	}
	timeStartTime, err := time.Parse("15:04", startTime)
	if err != nil {
		log.Fatal(err)
	}
	timeStopTime, err := time.Parse("15:04", stopTime)
	if err != nil {
		log.Fatal(err)
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

func main() {
	cfg := getConfig()
	ec2Client := ec2.NewFromConfig(cfg)

	result, err := ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
	if err != nil {
		log.Fatal(err)
	}

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			days := getTag(instance.Tags, "Day")
			startTime := getTag(instance.Tags, "StartTime")
			stopTime := getTag(instance.Tags, "StopTime")
			instanceState := string(instance.State.Name)
			instanceId := *instance.InstanceId

			fmt.Printf("Instance id: %s, name: %s, state: %s\nScheduler Days: %s, StartTime: %s, StopTime: %s\n",
				*instance.InstanceId,
				getTag(instance.Tags, "Name"),
				instanceState,
				days,
				startTime,
				stopTime)
			isActionNeeded := isActionNeeded(days, startTime, stopTime, instanceState)
			fmt.Printf("Action neeeded: %v\n", isActionNeeded)
			if isActionNeeded && instanceState == "running" {
				fmt.Println("-> stopping the instance")
				_, err := ec2Client.StopInstances(context.TODO(), &ec2.StopInstancesInput{
					InstanceIds: []string{instanceId},
				})
				if err != nil {
					log.Fatal(err)
				}
			}
			if isActionNeeded && instanceState == "stopped" {
				fmt.Println("-> starting the instance")
				_, err := ec2Client.StartInstances(context.TODO(), &ec2.StartInstancesInput{
					InstanceIds: []string{instanceId},
				})
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}
