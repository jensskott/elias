package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	region       = kingpin.Flag("region", "Aws region of cluster").Required().Short('r').String()
	cluster      = kingpin.Flag("cluster", "Cluster name").Required().Short('c').String()
	awsCredsFile = kingpin.Flag("file", "Aws credentials file full url").Required().Short('f').String()
	awsProfile   = kingpin.Flag("profile", "Aws profile to use in the creds file").Default("default").Short('p').String()
)

func main() {
	kingpin.Version("0.0.1")
	kingpin.Parse()
	latestAgentVersion := "1.17.0"

	creds := credentials.NewSharedCredentials(*awsCredsFile, *awsProfile)

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(*region),
		Credentials: creds,
	})
	if err != nil {
		log.Fatal(err)
	}

	client := ecs.New(sess)

	resp, err := client.ListContainerInstances(&ecs.ListContainerInstancesInput{
		Cluster: aws.String(*cluster),
	})
	if err != nil {
		log.Fatal(err)
	}

	instances, err := client.DescribeContainerInstances(&ecs.DescribeContainerInstancesInput{
		Cluster:            aws.String(*cluster),
		ContainerInstances: resp.ContainerInstanceArns,
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, c := range instances.ContainerInstances {
		if *c.VersionInfo.AgentVersion != latestAgentVersion {
			fmt.Println("Updating container agent")
			_, err := client.UpdateContainerAgent(&ecs.UpdateContainerAgentInput{
				Cluster:           aws.String(*cluster),
				ContainerInstance: c.ContainerInstanceArn,
			})
			if err != nil {
				log.Fatal(err)
			}
		} else {
			fmt.Println("Container agent up to date")
		}
	}
}
