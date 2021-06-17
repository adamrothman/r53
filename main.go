package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	log "github.com/sirupsen/logrus"
)

const recordType = types.RRTypeA

func getExternalIP() (string, error) {
	res, err := http.Get("https://checkip.amazonaws.com/")
	if err != nil {
		return "", fmt.Errorf("sending request: %w", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	}

	return string(bytes.TrimSpace(body)), err
}

func getRegisteredIP(client *route53.Client, ctx context.Context, hostedZone, recordName string) (*string, error) {
	out, err := client.ListResourceRecordSets(
		ctx,
		&route53.ListResourceRecordSetsInput{
			HostedZoneId:    aws.String(hostedZone),
			StartRecordName: aws.String(recordName),
			StartRecordType: recordType,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("sending AWS request: %w", err)
	}

	if len(out.ResourceRecordSets) == 0 {
		return nil, nil
	}

	recordSet := out.ResourceRecordSets[0]
	if *recordSet.Name != recordName || recordSet.Type != recordType || len(recordSet.ResourceRecords) == 0 {
		return nil, nil
	}

	return recordSet.ResourceRecords[0].Value, nil
}

func updateRecord(client *route53.Client, ctx context.Context, hostedZone, recordName, ip string, ttl int64) error {
	comment := time.Now().Format(time.RFC3339Nano)

	_, err := client.ChangeResourceRecordSets(
		ctx,
		&route53.ChangeResourceRecordSetsInput{
			HostedZoneId: &hostedZone,
			ChangeBatch: &types.ChangeBatch{
				Comment: aws.String(comment),
				Changes: []types.Change{
					{
						Action: types.ChangeActionUpsert,
						ResourceRecordSet: &types.ResourceRecordSet{
							Name: aws.String(recordName),
							ResourceRecords: []types.ResourceRecord{
								{Value: aws.String(ip)},
							},
							TTL:  aws.Int64(ttl),
							Type: recordType,
						},
					},
				},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("sending AWS request: %w", err)
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

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.WithError(err).Fatal("Error loading SDK config")
	}

	client := route53.NewFromConfig(cfg)

	externalIP, err := getExternalIP()
	if err != nil {
		log.WithError(err).Fatal("Error getting external IP")
	}

	log = log.WithField("ip", externalIP)

	registeredIP, err := getRegisteredIP(client, ctx, hostedZone, recordName)
	if err != nil {
		log.WithError(err).Fatal("Error getting registered IP")
	}

	if registeredIP != nil && *registeredIP == externalIP {
		log.Info("Not updating record; IP address unchanged")
		return
	}

	if err := updateRecord(client, ctx, hostedZone, recordName, externalIP, ttl); err != nil {
		log.WithError(err).Fatal("Error updating record")
	} else {
		log.Info("Record updated")
	}
}
