package api

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

const recordType = types.RRTypeA

func GetRecordValue(client *route53.Client, ctx context.Context, hostedZone, recordName string) (*string, error) {
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

func UpdateRecord(client *route53.Client, ctx context.Context, hostedZone, recordName, ip string, ttl int64) error {
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
