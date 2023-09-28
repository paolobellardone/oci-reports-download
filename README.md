# oci_reports_download

A cli to download the usage and costs reports from OCI tenancy
Developed and tested on Mac, it works also on Linux and Windows

## Prerequisites

In order to be able to use the cli, you have to implement the following prerequisites:

- Create an OCI configuration file (<https://docs.oracle.com/en-us/iaas/Content/API/Concepts/sdkconfig.htm>) and optionally install oci-cli (<https://docs.oracle.com/en-us/iaas/Content/API/SDKDocs/cliinstall.htm>)
- Configure the OCI policies needed to access to the reports (<https://docs.oracle.com/en-us/iaas/Content/Billing/Concepts/usagereportsoverview.htm>)
- Compile or download the compiled binary (see below) and into a directory of your choice, ideally in your path
- Make the cli executable with the command `chmod +x oci_reports_download`

## Compile and Build

To compile and build the cli, please follow these steps:

- Prerequisite
  - a working **go** installation
- Download or clone this repository
  - <https://github.com/paolobellardone/oci_reports_download/archive/refs/heads/main.zip>
  - `git clone https://github.com/paolobellardone/oci_reports_download.git`
- Run this command to compile and build the cli
  - `make clean build`
- Copy your brand new cli in a directory of your choice, ideally on your binary path
- (MacOS only) At the first run you have to authorize the execution of the cli by allowing it from "System Settings" --> "Privacy & Security"

## Download

The latest version and the previous ones are available on [Releases](<https://github.com/paolobellardone/oci_reports_download/releases>) page

## Usage

Usage:  
&ensp;oci_reports_download [flags]

Flags:  
&ensp;-t, --tenancy string&emsp;&emsp;           the OCID of your tenancy (required)
&ensp;-r, --report-type string&emsp;&emsp;       the type of report to download - allowed values: all, usage, cost (default "all")
&ensp;-i, --report-interval string&emsp;   the period of time to consider for reports - allowed values: yyyy-mm-dd, yyyy-mm, yyyy
&ensp;-p, --profile string&emsp;&emsp;           the profile defined in ~/.oci/config to use to connect to OCI (default "DEFAULT")
&ensp;-u, --uncompress&emsp;&emsp;               uncompress the downloaded files
&ensp;-h, --help&emsp;&emsp;                     help for oci_reports_download
&ensp;-v, --version&emsp;&emsp;                  version for oci_reports_download

Date formats:

- YYYY: all the files for the specified year
- YYYY-MM: all the files for the specified month
- YYYY-MM-DD: all the files for the specified day
- If the argument --report-interval|-i is not specified, the cli will download all the available files in the usage and cost pools
