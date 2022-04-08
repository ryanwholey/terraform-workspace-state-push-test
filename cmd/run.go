package cmd

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"
)

func Execute() {
	cmd := &cobra.Command{
		Use:   "terrafrom-copy-state",
		Short: "Transfer state between Terraform Cloud backends",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			flags := cmd.Flags()

			organization, err := flags.GetString("organization")
			if err != nil {
				return fmt.Errorf("error parsing organization flag: %w", err)
			}

			if organization == "" {
				return fmt.Errorf("--organization,-o required")
			}

			srcName := args[0]

			dstName, err := flags.GetString("destination")
			if err != nil {
				return fmt.Errorf("error parsing destination flag: %w", err)
			}

			srcAddress, err := flags.GetString("source-address")
			if err != nil {
				return fmt.Errorf("error parsing source-address flag: %w", err)
			}

			dstAddress, err := flags.GetString("destination-address")
			if err != nil {
				return fmt.Errorf("error parsing destination-address flag: %w", err)
			}

			srcClient, err := tfe.NewClient(&tfe.Config{
				Token:   os.Getenv("TF_SRC_TOKEN"),
				Address: srcAddress,
			})
			if err != nil {
				return err
			}

			dstClient, err := tfe.NewClient(&tfe.Config{
				Token:   os.Getenv("TF_DST_TOKEN"),
				Address: dstAddress,
			})
			if err != nil {
				return err
			}

			src, err := srcClient.Workspaces.Read(ctx, organization, srcName)
			if err != nil {
				return err
			}

			b, err := downloadStateFile(ctx, srcClient, src)
			if err != nil {
				return err
			}

			dst, err := dstClient.Workspaces.Read(ctx, organization, dstName)
			if err != nil {
				return err
			}

			if _, err := dstClient.Workspaces.Lock(ctx, dst.ID, tfe.WorkspaceLockOptions{
				Reason: tfe.String("Pushing initial state"),
			}); err != nil {
				return err
			}

			_, err = dstClient.StateVersions.Create(ctx, dst.ID, tfe.StateVersionCreateOptions{
				Serial: tfe.Int64(1),
				State:  tfe.String(base64.StdEncoding.EncodeToString(b)),
				MD5:    tfe.String(fmt.Sprintf("%x", md5.Sum(b))),
			})
			if err != nil {
				return err
			}

			if _, err := dstClient.Workspaces.Unlock(ctx, dst.ID); err != nil {
				return err
			}

			return nil
		},
	}

	flags := cmd.Flags()

	flags.StringP("destination", "d", "", "Destination workspace name. If not provided, source workspace name will be used")
	flags.String("source-address", "https://app.terraform.io", "Source host")
	flags.String("destination-address", "https://app.terraform.io", "Destination host")
	flags.StringP("organization", "o", "", "Terraform Cloud organization")

	cobra.CheckErr(cmd.Execute())
}

func downloadStateFile(ctx context.Context, tf *tfe.Client, workspace *tfe.Workspace) ([]byte, error) {
	srcVersion, err := tf.StateVersions.ReadCurrent(ctx, workspace.ID)
	if err != nil {
		return []byte{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", srcVersion.DownloadURL, nil)
	if err != nil {
		return []byte{}, err
	}

	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return []byte{}, err
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return []byte{}, err
	}

	return b, nil
}
