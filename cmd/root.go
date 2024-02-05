/*
 * MIT License
 *
 * Copyright (c) 2024 PaoloB
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package cmd

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/objectstorage"
)

// Constants used throughout the cli
const (
	// Change the version when write new code or modify it
	cliVersion = "1.0.5"
	// DO NOT CHANGE the values below
	cliName            = "oci-reports-download"
	cfgDirName         = ".oci"
	cfgFileName        = "config"
	reportingNamespace = "bling"
	costPrefix         = "reports/cost-csv/"
	usagePrefix        = "reports/usage-csv/"
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:     cliName,
	Short:   "Download usage and cost report files from your OCI tenancy",
	Long:    `Download usage and cost report files from your OCI tenancy`,
	Version: cliVersion,
	Run: func(cmd *cobra.Command, args []string) {

		// Defined parameters for the cli
		var tenancyOcid string
		var reportType string // This is required
		var profileName string
		var reportInterval string
		var uncompressFiles bool

		// Read the flags from the command line
		reportType, _ = cmd.Flags().GetString("report-type")
		profileName, _ = cmd.Flags().GetString("profile")
		reportInterval, _ = cmd.Flags().GetString("report-interval")
		uncompressFiles, _ = cmd.Flags().GetBool("uncompress")

		// Get home folder for user and get the config file path
		homeFolder := getHomeFolder()
		configFilePath := filepath.Join(homeFolder, cfgDirName, cfgFileName)

		// Create a new OCI Object Storage client to invoke the APIs
		configurationProvider := common.CustomProfileConfigProvider(configFilePath, profileName)
		// Read the tenancy from the .oci/config specified profile
		tenancyOcid, _ = configurationProvider.TenancyOCID()
		osClient, ociErr := objectstorage.NewObjectStorageClientWithConfigurationProvider(configurationProvider)
		exitOnError(ociErr)

		// Get a context for the cli
		ctx := context.Background()

		// Create a ListObjectRequest with fields filled with values coming from cli flags
		loReq := objectstorage.ListObjectsRequest{
			NamespaceName: common.String(reportingNamespace),
			BucketName:    common.String(tenancyOcid),
			Fields:        common.String("name,size,timeCreated"),
		}

		// Check which reports should be downloaded
		switch reportType {
		case "cost":
			loReq.Prefix = common.String(costPrefix)
		case "usage":
			loReq.Prefix = common.String(usagePrefix)
		default:
			exitOnError(errors.New("Invalid argument for parameter report-type. Allowed values are: cost, usage"))
		}

		// Invoke the ListObjects APIs to get the selected report files
		for loRsp, ociErr := osClient.ListObjects(ctx, loReq); ; loRsp, ociErr = osClient.ListObjects(ctx, loReq) {
			exitOnError(ociErr)
			for _, value := range loRsp.ListObjects.Objects {
				// Filter files by date, if needed
				if strings.Contains(value.TimeCreated.Format("2006-01-02"), reportInterval) {
					// Download a report file
					fmt.Print("Downloading file ", *value.Name)
					goReq := objectstorage.GetObjectRequest{
						NamespaceName: common.String(reportingNamespace),
						BucketName:    common.String(tenancyOcid),
						ObjectName:    common.String(*value.Name),
					}

					goRes, ociErr := osClient.GetObject(ctx, goReq)
					exitOnError(ociErr)

					// Save the file locally
					fileName := strings.TrimPrefix(*value.Name, *loReq.Prefix)
					if uncompressFiles {
						outFile, err := os.Create(strings.TrimSuffix(fileName, ".gz"))
						exitOnError(err)
						defer outFile.Close()

						reader, err := gzip.NewReader(goRes.Content)
						exitOnError(err)
						defer reader.Close()

						_, err = io.Copy(outFile, reader)
						exitOnError(err)
					} else {
						outCompressedFile, err := os.Create(fileName)
						exitOnError(err)
						defer outCompressedFile.Close()

						written, err := io.Copy(outCompressedFile, goRes.Content)
						if err != nil || written != *goRes.ContentLength {
							exitOnError(err)
						}
					}
					fmt.Println(" Done!")
				}
			}
			if loRsp.ListObjects.NextStartWith != nil {
				loReq.Start = loRsp.ListObjects.NextStartWith
			} else {
				break
			}
		}

	},
}

// Configure and run the cli
func Execute() {
	// Print a banner
	banner := cliName + " " + cliVersion
	fmt.Println(banner)
	fmt.Println(strings.Repeat("-", len(banner)))

	// Execute the cli
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// Initialize the cli with supported flags and their default values
func init() {
	rootCmd.Flags().StringP("report-type", "t", "", "the type of report to download - allowed values: usage, cost")
	rootCmd.MarkFlagRequired("report-type")
	rootCmd.Flags().StringP("report-interval", "i", "", "the period of time to consider for reports - allowed values: yyyy-mm-dd, yyyy-mm, yyyy")
	rootCmd.Flags().StringP("profile", "p", "DEFAULT", "the profile defined in ~/.oci/config to use to connect to OCI (case-sensitive)")
	rootCmd.Flags().BoolP("uncompress", "u", false, "uncompress the downloaded files")
	rootCmd.Flags().SortFlags = false
}

// Utility functions

// Get the user's home folder
func getHomeFolder() string {
	current, e := user.Current()
	if e != nil {
		// If the user.Current() method does not work, it reads the user's home folder from the enviroment
		home := os.Getenv("HOME") // Linux/MacOS
		if home == "" {
			home = os.Getenv("USERPROFILE") // Windows
		}
		return home
	}
	return current.HomeDir
}

// Fatal exit on error
func exitOnError(err error) {
	if err != nil {
		log.Fatalln(err.Error())
	}
}
