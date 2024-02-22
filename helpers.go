package main

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func ec2ClientWrapper(client string, action string, data string, dryrun string) (output string) {
	if client == "cli" {
		output = ec2ClientShell(action, data, dryrun)
	} else if client == "native" {
		output = ec2ClientNative(action, data, dryrun)
	} else {
		output = "Incompatible Client Selection"
	}
	return output
}

func ec2ClientNative(action string, data string, dryrun string) (output string) {
	var host_ip string

	log.Printf("INFO: Native Client Request Params: Action: %s, Data: %s, DryRun: %s", action, data, dryrun)
	if action == "" {
		log.Printf("WARN: Invalid Action provided: %s. No action shall be taken")
	}

	drflag, err := strconv.ParseBool(dryrun)
	if err != nil {
		log.Printf("WARN: Invalid DryRun flag: %s. Assumed value: true", dryrun)
		drflag = true
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := ec2.New(sess)

	if action == "stop" {
		input := &ec2.StopInstancesInput{
			InstanceIds: []*string{
				aws.String(data),
			},
			DryRun: aws.Bool(drflag),
		}
		result, err := svc.StopInstances(input)
		if err != nil {
			log.Printf("ERROR: %s", err)
		} else {
			log.Printf("INFO: Success: %s", result.StoppingInstances)
			output = "Stopped"
		}
	} else if action == "start" {
		input := &ec2.StartInstancesInput{
			InstanceIds: []*string{
				aws.String(data),
			},
			DryRun: aws.Bool(drflag),
		}
		result, err := svc.StartInstances(input)
		if err != nil {
			log.Printf("ERROR: %s", err)
		} else {
			log.Printf("INFO: Success: %s", result.StartingInstances)
			output = "Started"
		}
	} else if action == "findByIp" {
		ips, _ := net.LookupIP(data)
		for _, ip := range ips {
			if ipv4 := ip.To4(); ipv4 != nil {
				host_ip = ipv4.String()
			}
		}
		log.Printf("INFO: hostname: %s, host_ip: %s", data, host_ip)
		input := &ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("private-ip-address"),
					Values: []*string{
						aws.String(host_ip),
					},
				},
			},
		}
		result, err := svc.DescribeInstances(input)
		if err != nil {
			log.Printf("ERROR: %s", err)
		}
		for _, reservation := range result.Reservations {
			for _, i := range reservation.Instances {
				output = *i.InstanceId
			}
		}
	} else {
		log.Printf("ERROR: Invalid action: %s", action)
		output = "IncompatibleRequest"
	}

	return output
}

func ec2ClientShell(action string, data string, dryrun string) (output string) {
	var host_ip string
	var cmd *exec.Cmd

	log.Printf("INFO: CLI Client Request Params: Action: %s, Data: %s, DryRun: %s", action, data, dryrun)
	if action == "" {
		log.Printf("WARN: Invalid Action provided: %s. No action shall be taken")
	}

	drflag, err := strconv.ParseBool(dryrun)
	if err != nil {
		log.Printf("WARN: Invalid DryRun flag: %s. Assumed Value: true", dryrun)
		drflag = true
	}

	if drflag == true {
		dryrun = "--dry-run"
	} else {
		dryrun = "--no-dry-run"
	}

	if action == "stop" {
		cmd = exec.Command("aws", "ec2", "stop-instances", "--instance-ids", data, dryrun)
	} else if action == "start" {
		cmd = exec.Command("aws", "ec2", "start-instances", "--instance-ids", data, dryrun)
	} else if action == "findByIp" {
		ips, _ := net.LookupIP(data)
		for _, ip := range ips {
			if ipv4 := ip.To4(); ipv4 != nil {
				host_ip = ipv4.String()
			}
		}
		log.Printf("INFO: hostname: %s, host_ip: %s", data, host_ip)
		filter := "--filter=Name=private-ip-address,Values=" + host_ip
		query := "--query=Reservations[0].Instances[0].InstanceId"
		cmd = exec.Command("aws", "ec2", "describe-instances", filter, query)
		fmt.Println(cmd)
	} else {
		log.Printf("ERROR: Invalid action: %s", action)
		return "IncompatibleRequest"
	}

	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		log.Printf("INFO: %s", stdout)
	}
	return string(stdout)
}
