# Scheduler to start/stop EC2 instances at given time

This Golang project aims to provide a utility to automate starting/stopping of AWS EC2 instances.

The following tags needs to be assigned to an EC2 instance to make this work:
* Day (Mon,Tue,Wed,Thu,Fri,Sat,Sun)
* StartTime (HH:MM)
* StopTime (HH:MM)

In case the EC2 instance is stopped and it should run according to the schedule, it will be started and vice-versa.

How to build the image:

    docker build -t aws-ec2-scheduler:latest .

Helm chart installation:
    helm install -n aws-ec2-scheduler --create-namespace --set-file config=/path/to/config aws-ec2-scheduler ./helm

Do not forget to create the config at /path/to/config (an example filepath) and put inside the roles of different AWS accounts this scheduler can assume and control EC2 instances.
