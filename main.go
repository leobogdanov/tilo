package main

import (
	"flag"
	"strings"

	"github.com/leobogdanov/tilo/aws"
)

func main() {

	region := flag.String("region", "us-east-1", "AWS Region")
	role := flag.String("role", "", "Role to assume. Will default to AWS credentials permissions if not specified")
	services := flag.String("services", "ec2", "Comma-separated list of services")
	period := flag.Int("period", 5, "Period of cloudwatch metric sampling in minutes")
	samples := flag.Int("samples", 36, "Number of samples to look back on")
	threshold := flag.Float64("threshold", 0, "Metric threshold")
	dryRun := flag.Bool("dryRun", false, "Dry run mode won't stop any instances but will check if all permissions are present. Only supported for EC2")
	flag.Parse()

	var roleInput *string
	if *role != "" {
		roleInput = role
	}

	aws.ShutdownInactive(region, roleInput,
		strings.Split(*services, ","), *period, *samples, *threshold, *dryRun)

}
