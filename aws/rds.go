package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/rds"
)

// RDS primary struct
type RDS struct {
	*Status
}

// NewRDS RDS constructor
func NewRDS() *RDS {
	return &RDS{
		Status: &Status{}}
}

// ShutdownInactive scans through and stops unused RDS instances
func (c *RDS) ShutdownInactive(sess *session.Session, period int,
	cycles int, threshold float64) error {

	instances, err := c.listRDS(sess)
	if err != nil {
		log.Println("Error listing rds instances", err)
		return err
	}

	cw := cloudwatch.New(sess)

	for _, instance := range instances {
		//log.Printf("Instance is %v", instance)
		endTime := time.Now()
		lookBackMinutes := period * (cycles + 3)
		duration, _ := time.ParseDuration(fmt.Sprintf("-%dm", lookBackMinutes))
		startTime := endTime.Add(duration)

		query := &cloudwatch.MetricDataQuery{
			Id: aws.String("connections"),
			MetricStat: &cloudwatch.MetricStat{
				Metric: &cloudwatch.Metric{
					Namespace:  aws.String("AWS/RDS"),
					MetricName: aws.String("DatabaseConnections"),
					Dimensions: []*cloudwatch.Dimension{
						&cloudwatch.Dimension{
							Name:  aws.String("DBInstanceIdentifier"),
							Value: instance.DBInstanceIdentifier,
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
			log.Printf("RDS InstanceId %s: threshold breached. Usage detected. Ignoring\n", *instance.DBInstanceIdentifier)
			c.Status.SkippedCount++
		} else if offset < cycles {
			log.Printf("RDS InstanceId %s: not enough data\n", *instance.DBInstanceIdentifier)
			c.Status.SkippedCount++
		} else if offset >= cycles {
			log.Printf("RDS InstanceId %s: threshold not breached. Stopping instance\n", *instance.DBInstanceIdentifier)
			res, err := c.stopRDS(sess, instance.DBInstanceIdentifier)
			if err != nil {
				log.Printf("RDS InstanceId %s: error stopping %v\n", *instance.DBInstanceIdentifier, err)
				c.Status.ErrorCount++
			} else {
				log.Printf("RDS InstanceId %s: stopped. %v\n", *instance.DBInstanceIdentifier, res)
				c.Status.StoppedCount++
			}

		}

	}
	return nil

}

func (c *RDS) listRDS(session *session.Session) ([]*rds.DBInstance, error) {

	// Create new EC2 client
	rdsSvc := rds.New(session)

	// Call to get detailed information on each instance

	result, err := rdsSvc.DescribeDBInstances(nil)
	if err != nil {
		fmt.Println("Error", err)
		return nil, err
	}
	var runningInstances []*rds.DBInstance
	for _, instance := range result.DBInstances {
		if *instance.DBInstanceStatus == "available" {
			runningInstances = append(runningInstances, instance)
		} else {
			c.Status.SkippedCount++
		}
	}

	return runningInstances, nil
}

func (c *RDS) stopRDS(session *session.Session, instanceID *string) (*rds.StopDBInstanceOutput, error) {

	// Create new RDS client
	rdsSvc := rds.New(session)

	return rdsSvc.StopDBInstance(&rds.StopDBInstanceInput{
		DBInstanceIdentifier: instanceID,
	})

}
