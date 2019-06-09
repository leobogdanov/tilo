# tilo
Command line tool that stops dormant AWS resources

Supported AWS services: EC2, RDS

### Build 

1. ```git clone https://github.com/leobogdanov/tilo```
2. ```cd tilo```
3. ```docker build -t tilo .```

### Test installation 

1. ```docker run tilo tilo -h```
2. You should see the help screen

### Metrics
The metrics being checked are ```CPUUtilization``` for EC2 instances or ```DatabaseConnections``` for RDS instances

The tool requires AWS api credentials. ```AWS_ACCESS_KEY_ID``` and ```AWS_SECRET_ACCESS_KEY```. See ```tilo_policy.json``` 
for minimum required permissions to run the tool. You may either assign the policy to the user directly, or to a role that
the tool will assume via AssumeRole api call.

### Running without external role
1. ```docker run -e AWS_ACCESS_KEY_ID=<key> -e AWS_SECRET_ACCESS_KEY=<secret> tilo tilo --services=ec2 --period=5 
--samples=36 --threshold=1
This will stop all EC2 instances whose CPU Utilization hasn't gone above 1% for the last 3 hours (36 5-minute samples)

### Running with external role
1. ```docker run -e AWS_ACCESS_KEY_ID=<key> -e AWS_SECRET_ACCESS_KEY=<secret> tilo tilo --services=ec2 --period=5 
--samples=36 --threshold=1 --role=arn:aws:iam::<account>:role/<role-name>
This will stop all EC2 instances whose CPU Utilization hasn't gone above 1% for the last 3 hours (36 5-minute samples)
