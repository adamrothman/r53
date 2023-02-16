package cmd

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"

	"github.com/adamrothman/r53/internal/api"
)

var (
	// Flags
	recordName string
	ttl        int64

	updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update a Route 53 record with the system's public IP",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := slog.With(
				slog.String("hosted_zone", hostedZone),
				slog.String("record_name", recordName),
				slog.Int64("ttl", ttl),
			)

			ctx := context.Background()

			cfg, err := config.LoadDefaultConfig(ctx)
			if err != nil {
				log.Error("Failed to load AWS SDK config", err)
				return err
			}

			client := route53.NewFromConfig(cfg)

			publicIP, err := api.GetPublicIP()
			if err != nil {
				log.Error("Failed to get public IP", err)
				return err
			}

			log = log.With(slog.String("ip", publicIP))

			registeredIP, err := api.GetRecordValue(client, ctx, hostedZone, recordName)
			if err != nil {
				log.Error("Failed to get registered IP", err)
				return err
			}

			if registeredIP != nil && *registeredIP == publicIP {
				log.Info("Not updating record; IP address unchanged")
				return nil
			}

			if err := api.UpdateRecord(client, ctx, hostedZone, recordName, publicIP, ttl); err != nil {
				log.Error("Record update failed", err)
				return err
			} else {
				log.Info("Record updated")
			}

			return nil
		},
	}
)

func init() {
	f := updateCmd.Flags()

	f.StringVarP(&recordName, "record", "r", "", "record name")
	updateCmd.MarkFlagRequired("record")

	f.Int64VarP(&ttl, "ttl", "t", 300, "record TTL")

	rootCmd.AddCommand(updateCmd)
}
