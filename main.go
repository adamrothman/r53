package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	log "github.com/sirupsen/logrus"
)

const recordType = route53.RRTypeA

func getExternalIP() (string, error) {
	res, err := http.Get("https://checkip.amazonaws.com/")
	if err != nil {
		return "", fmt.Errorf("sending request: %s", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %s", err)
	}

	return string(bytes.TrimSpace(body)), err
}

func getRegisteredIP(client *route53.Route53, hostedZone, recordName string) (*string, error) {
	req := client.ListResourceRecordSetsRequest(&route53.ListResourceRecordSetsInput{
		HostedZoneId:    aws.String(hostedZone),
		StartRecordName: aws.String(recordName),
		StartRecordType: recordType,
	})

	res, err := req.Send()
	if err != nil {
		return nil, fmt.Errorf("sending AWS request: %s", err)
	}

	if len(res.ResourceRecordSets) == 0 {
		return nil, nil
	}

	recordSet := res.ResourceRecordSets[0]
	if *recordSet.Name != recordName || recordSet.Type != recordType || len(recordSet.ResourceRecords) == 0 {
		return nil, nil
	}

	return recordSet.ResourceRecords[0].Value, nil
}

func updateRecord(client *route53.Route53, hostedZone, recordName, ip string, ttl int64) error {
	comment := time.Now().Format(time.RFC3339Nano)

	req := client.ChangeResourceRecordSetsRequest(&route53.ChangeResourceRecordSetsInput{
		HostedZoneId: &hostedZone,
		ChangeBatch: &route53.ChangeBatch{
			Comment: aws.String(comment),
			Changes: []route53.Change{
				route53.Change{
					Action: route53.ChangeActionUpsert,
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(recordName),
						ResourceRecords: []route53.ResourceRecord{
							route53.ResourceRecord{
								Value: aws.String(ip),
							},
						},
						TTL:  aws.Int64(ttl),
						Type: recordType,
					},
				},
			},
		},
	})

	_, err := req.Send()
	if err != nil {
		return fmt.Errorf("sending AWS request: %s", err)
	}

	return nil
}

func init() {
	log.SetOutput(os.Stdout)
}

func main() {
	var hostedZone, recordName string
	var ttl int64

	flag.StringVar(&hostedZone, "zone", "", "hosted zone ID")
	flag.StringVar(&recordName, "record", "", "record name")
	flag.Int64Var(&ttl, "ttl", 300, "record TTL")
	flag.Parse()

	if hostedZone == "" || recordName == "" {
		flag.Usage()
		os.Exit(1)
	}

	log := log.WithFields(log.Fields{
		"hosted_zone": hostedZone,
		"record_name": recordName,
		"ttl":         ttl,
	})

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		log.WithError(err).Fatal("Error loading SDK config")
	}

	r53 := route53.New(cfg)

	externalIP, err := getExternalIP()
	if err != nil {
		log.WithError(err).Fatal("Error getting external IP")
	}

	log = log.WithField("ip", externalIP)

	registeredIP, err := getRegisteredIP(r53, hostedZone, recordName)
	if err != nil {
		log.WithError(err).Fatal("Error getting registered IP")
	}

	if registeredIP != nil && *registeredIP == externalIP {
		log.Info("Not updating record; IP address unchanged")
		return
	}

	if err := updateRecord(r53, hostedZone, recordName, externalIP, ttl); err != nil {
		log.WithError(err).Fatal("Error updating record")
	} else {
		log.Info("Record updated")
	}
}
