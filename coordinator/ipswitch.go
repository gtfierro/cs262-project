package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type IPSwitcher interface {
	// attempts to acquire the IP address, returning true
	// if successful
	AcquireIP() error
}

type AWSIPSwitcher struct {
	// the AWS region
	region string
	// our instance id
	iid string
	// elastic IP
	ipaddr string
	// identifier for elastic ip
	eipid *string
	// the id to use to disassociate the address from its old instance
	associd *string
	// reference to EC2 API
	ec2 *ec2.EC2
}

func NewAWSIPSwitcher(iid, region, ipaddr string) (*AWSIPSwitcher, error) {
	ip := &AWSIPSwitcher{
		iid:    iid,
		region: region,
		ipaddr: ipaddr,
		ec2:    ec2.New(session.New(), &aws.Config{Region: aws.String(region)}),
	}
	// need to grab the allocation reference for the given elastic IP
	describeAddressesParams := &ec2.DescribeAddressesInput{
		Filters: []*ec2.Filter{&ec2.Filter{Name: aws.String("public-ip"), Values: []*string{aws.String(ipaddr)}}},
	}
	addresses, err := ip.ec2.DescribeAddresses(describeAddressesParams)
	if err != nil {
		return ip, err
	}
	address := addresses.Addresses[0]
	ip.associd = address.AssociationId
	ip.eipid = address.AllocationId

	return ip, nil
}

func (ip *AWSIPSwitcher) AcquireIP() error {
	// need to grab the allocation reference for the given elastic IP
	describeAddressesParams := &ec2.DescribeAddressesInput{
		Filters: []*ec2.Filter{&ec2.Filter{Name: aws.String("public-ip"), Values: []*string{aws.String(ip.ipaddr)}}},
	}
	addresses, err := ip.ec2.DescribeAddresses(describeAddressesParams)
	if err != nil {
		return err
	}
	address := addresses.Addresses[0]
	associd := address.AssociationId

	disassocParams := &ec2.DisassociateAddressInput{
		AssociationId: associd,
	}

	// these are not equal if we lost control
	if associd != ip.associd {
		ip.ec2.DisassociateAddress(disassocParams)
	}

	assocParams := &ec2.AssociateAddressInput{
		AllocationId:       ip.eipid,
		DryRun:             aws.Bool(false),
		InstanceId:         aws.String(ip.iid),
		AllowReassociation: aws.Bool(true),
	}

	_, err = ip.ec2.AssociateAddress(assocParams)
	if err != nil {
		return err
	}

	return nil
}
