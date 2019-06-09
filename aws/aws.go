package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
)

// Status summary results of the run
type Status struct {
	StoppedCount int64
	SkippedCount int64
	ErrorCount   int64
}

// ShutdownInactive main entry point for the job
func ShutdownInactive(region *string, role *string, inServices []string, period int,
	samples int, threshold float64, dryRun bool) error {

	sess, err := session.NewSession(&aws.Config{
		Region: region},
	)
	if err != nil {
		log.Println("Error creating session", err)
		return err
	}

	if role != nil {
		creds := stscreds.NewCredentials(sess, *role)

		sessWithRole, err := session.NewSession(&aws.Config{
			Region:      region,
			Credentials: creds})
		if err != nil {
			log.Println("Error creating session with role", err)
			return err
		}
		sess = sessWithRole
	}
	for _, service := range inServices {
		switch service {
		case "ec2":
			c := NewEC2()
			err := c.ShutdownInactive(sess, period, samples, threshold, dryRun)
			if err != nil {
				log.Printf("EC2 scan aborted with error: %v", err)
			} else {
				log.Printf("EC2 results %+v", *c.Status)
			}
		case "rds":
			c := NewRDS()
			err := c.ShutdownInactive(sess, period, samples, threshold)
			if err != nil {
				log.Printf("RDS scan aborted with error: %v", err)
			} else {
				log.Printf("RDS results %+v", *c.Status)
			}
		default:
			log.Printf("Unsupported service %s", service)
		}
	}

	return nil
}
