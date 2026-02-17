# Request For Proposal Parser

## Overview
This version of the application walks through a specified root directory locally on your machine, parses each docx, xlsx, and pdf file within each RFP Package directory that is located within the specified root directory, and then posts the data into Smartsheets. Within Smartsheets, you can easily create dashboards based off this data.

## Installation
1. Download the source code for this client app
2. Install pdftotext 
  - sudo apt update
  - sudo apt poppler-utils
    - pdftotext is included in the poppler-utils package
  - pdftotext -v
    - Verify installation
3. Create a Smartsheet with the following columns (in this order): Date Parsed, RFP Package Name, KPI Name, KPI Category, KPI Context
4. Create a directory on your file system. Within this root directory, place all RFP Packages that need to be parsed (each folder represents a RFP Package).
5. Create a .env file. Within the .env file add the following variables:
  - SMARTSHEET_TOKEN - A Smartsheet access token that can be generated in Smartsheet
  - SMARTSHEET_URL - The URL of the Smartsheet to push the KPI data into
6. In smartsheet_post.go, update the const variables by entering the column ID's for each column in your smartsheet. Will need to obtain these column ID's using a curl request to your Smartsheet using the SMARTSHEET_TOKEN and SMARTSHEET_URL
7. In parser/kpiDefinitions.json, update this file to include the KPI's you would like to parse for inside of docx and xlsx files
8. In main.go, update the "RfpPackageRootDir" const variable to the location of the root directory created in step #3
9. Run program from this applications root directory: go run ./