# Request For Proposal Parser

## Overview
The RFP Parser is an application that traverses a Microsoft SharePoint Document Library, downloads .docx, .xlsx, and .pdf files within each Request for Proposal (RFP) package, extracts key performance indicators (KPIs), and posts the parsed data into Smartsheet. Within Smartsheet, dashboards can be easily created to visualize this data.

This application is designed to run locally.

## Motivation
The tool provides analytics into client priorities, such as specific ISO certifications or implemented programs. By analyzing RFP documents, leadership can prioritize which certifications or programs to pursue, improving the company's ability to win bids.

## 🚀 Setup - Local Configuration
1. Clone this repository to your local machine
- git clone https://github.com/JA50N14/rfp_parser.git

2. Create a SharePoint Site & Library
  - Within the Document Library have the following directory tree set up:
    - Year (e.g., 2025, 2026)
      - Business Unit (e.g., "Facilities Management")
        - Division (e.g., "FM East", "FM West")
          - RFP Packages (directories representing each RFP Package)
  - Add a dropdown column named ProcessStatus with options:
    - InProgress
    - Complete
    - Failed

3. Generate Private Key & Certificate
- Generate Private Key:
  - openssl genrsa -out graph-app.key 2048
- Generate self-signed certificate:
  - openssl req -new -x509 -key graph-app.key -out graph-app.crt -days 365
    - Common Name (CN): rfp_parser
- graph-app.key is private and must be kept secret.
- graph-app.crt is the public certificate uploaded to Entra ID.

4. Register Client App in Entra ID
- Go to: https://entra.microsoft.com/
- Register a new application:
  - App registration → obtain Client ID and Tenant ID
- Upload Certificate:
  - App registration -> your app -> Certificates & Secrets -> Certificates -> Upload certificate (graph-app.crt)
- Grant permissions:
  - API permissions -> Add a permission -> Microsoft Graph -> Application permissions
    - Select Sites.ReadWrite.All -> Grant admin consent (this step only requests permission, does not provide the permission)
    - An Entra ID Admin must approve this permission for the specific SharePoint Site. Will need to provide the Entra ID Admin the SharePoint Site ID.
    - App requires Microsoft Graph application permission Sites.Selected. Once consented, app must be granted read-write access. No user-delegated permissions needed.

5. Create Smartsheet
  - Columns (in order):
    - Date Parsed, Year, Business Unit, Division, RFP Package Name, KPI Name, KPI Category, KPI Context
  - Generate a Smartsheet access token: Account → Apps & Integrations

6. Configure Column IDs
  - Update walk/result_to_smartsheet_transform.go constants with your Smartsheet column IDs.
  - Use Smartsheet API (curl) to retrieve column IDs.

7. Create a .env file. Within the .env file add the following variables:
  - SMARTSHEET_TOKEN - A Smartsheet access token that can be generated in Smartsheet
  - SMARTSHEET_URL - The URL of the Smartsheet to push the KPI data into
  - GRAPH_PRIVATE_KEY_PATH - The path to where your Private Key is stored
  - GRAPH_CERTIFICATE_PATH - The path to where your certificate is stored
  - GRAPH_CLIENT_ID - The Client ID provided via Entra ID UI
  - GRAPH_TENANT_ID - The Tenant ID provided via Entra ID UI
  - GRAPH_SITE_ID - The ID of the SharePoint Site where the Document Library to be walked resides
  - GRAPH_LIBRARY_NAME - The name of the Document Library to walk
  - GRAPH_DRIVE_ID - The Drive ID of the Document Library to walk
  - SHAREPOINT_LIST_ID - The List ID of the Document Library to walk

8. In parser/kpiDefinitions.json, update this file to include the KPI's you would like to parse for inside of docx, xlsx, and pdf files

## Usage
- Run program from this applications root directory: go run ./

## Maintenance - Upload New Certificate & Private Key
1. Generate a public-private key pair
  - cmd: openssl genrsa -out graph-app.key 2048
    - graph-app.key -> Private key (keep this secret)
  - cmd: openssl req -new -x509 -key graph-app.key -out graph-app.crt -days 365
    - Prompted for fields: Common Name (CN) -> rfp_parser
2. Login to entra.microsoft.com
3. "App registration" -> Select "rfp_parser" -> "Certificates & secrets"
4. Upload new public key/certificate and remove old public key/certificate
5. Update GRAPH_PRIVATE_KEY and GRAPH_CERTIFICATE environment variables in Container App Job

## Contributing
- Clone the repo
  - cmd: git clone https://github.com/JA50N14/rfp_parser.git
