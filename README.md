# Scheduler to start/stop EC2 instances at given time

This Golang project aims to provide a utility to automate starting/stopping of AWS EC2 instances.

The following tags needs to be assigned to an EC2 instance to make this work:
Day (Mon,Tue,Wed,Thu,Fri,Sat,Sun)
StartTime (HH:MM)
StopTime (HH:MM)

In case the EC2 instance is stopped and it should run according to the schedule, it will be started and vice-versa.
