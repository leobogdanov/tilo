package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
)

const ec2RunningStatus int64 = 16 // EC2 status code

// EC2 primary struct
type EC2 struct {
	*Status
}

// NewEC2 EC2 constructor
func NewEC2() *EC2 {
	return &EC2{
		Status: &Status{}}
}

// ShutdownInactive scans through and stops unused EC2 instances
func (c *EC2) ShutdownInactive(sess *session.Session, period int,
	cycles int, threshold float64, dryRun bool) error {

	instances, err := c.listEc2(sess)
	if err != nil {
		log.Println("Error listing ec2 instances", err)
		return err
	}

	cw := cloudwatch.New(sess)

	for _, instance := range instances {

		endTime := time.Now()
		lookBackMinutes := period * (cycles + 3)
		duration, _ := time.ParseDuration(fmt.Sprintf("-%dm", lookBackMinutes))
		startTime := endTime.Add(duration)

		query := &cloudwatch.MetricDataQuery{
			Id: aws.String("cpu"),
			MetricStat: &cloudwatch.MetricStat{
				Metric: &cloudwatch.Metric{
					Namespace:  aws.String("AWS/EC2"),
					MetricName: aws.String("CPUUtilization"),
					Dimensions: []*cloudwatch.Dimension{
						&cloudwatch.Dimension{
							Name:  aws.String("InstanceId"),
							Value: instance.InstanceId,
						},
					},
				},
				Period: aws.Int64(int64(period) * 60),
				Stat:   aws.String("Maximum"),
			},
		}

		res, err := cw.GetMetricData(&cloudwatch.GetMetricDataInput{
			EndTime:           &endTime,
			StartTime:         &startTime,
			MetricDataQueries: []*cloudwatch.MetricDataQuery{query},
		})
		if err != nil {
			log.Println("Error metrics query instances ", err)
			return err
		}
		thresholdBreached := false
		offset := 0
		metricdata := res.MetricDataResults[0]
		for ; offset < len(metricdata.Timestamps) && offset < cycles; offset++ {
			if *metricdata.Values[offset] > threshold {
				thresholdBreached = true
				break
			}

		}
		if thresholdBreached {
			log.Printf("EC2 InstanceId %s: threshold breached. Usage detected. Ignoring\n", *instance.InstanceId)
			c.Status.SkippedCount++
		} else if offset < cycles {
			log.Printf("EC2 InstanceId %s: not enough data\n", *instance.InstanceId)
			c.Status.SkippedCount++
		} else if offset >= cycles {
			log.Printf("EC2 InstanceId %s: threshold not breached. Stopping instance\n", *instance.InstanceId)
			res, err := c.stopEc2(sess, instance.InstanceId, dryRun)
			if err != nil {
				log.Printf("EC2 InstanceId %s: error stopping %v\n", *instance.InstanceId, err)
				c.Status.ErrorCount++
			} else {
				log.Printf("EC2 InstanceId %s: stopped. %v\n", *instance.InstanceId, res)
				c.Status.StoppedCount++
			}

		}
	}
	return nil

}

func (c *EC2) listEc2(session *session.Session) ([]*ec2.Instance, error) {

	// Create new EC2 client
	ec2Svc := ec2.New(session)

	// Call to get detailed information on each instance
	result, err := ec2Svc.DescribeInstances(nil)
	if err != nil {
		fmt.Println("Error", err)
		return nil, err
	}
	var runningInstances []*ec2.Instance
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			if *instance.State.Code == ec2RunningStatus {
				runningInstances = append(runningInstances, instance)
			} else {
				c.Status.SkippedCount++
			}
		}

	}
	return runningInstances, nil
}

func (c *EC2) stopEc2(session *session.Session, instanceID *string, dryRun bool) (*ec2.StopInstancesOutput, error) {

	// Create new EC2 client
	ec2Svc := ec2.New(session)

	return ec2Svc.StopInstances(&ec2.StopInstancesInput{
		InstanceIds: []*string{instanceID},
		DryRun:      aws.Bool(dryRun)})

}
