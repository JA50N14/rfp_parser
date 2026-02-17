# Request For Proposal Parser

## Overview
This application is designed for being hosted on Azure Functions. This application walks a Microsoft SharePoint Document Library to download docx, xlsx, and pdf files within each RFP Package directory, parses them for KPI's, and then posts the data into Smartsheets. Within Smartsheets, you can easily create dashboards based off this data.

## Installation
1. Download the source code for this client app
2. Create a SharePoint site. Within the Document Library have the following directory-sub directory set up:
  - Year (i.e. "2025", "2026", etc.)
    - Business Unit (i.e. "Facilities Management", etc.)
      - Division (i.e. "FM East", "FM West", etc.)
        - RFP Packages - Where your RFP Packages are located. Each directory needs to represent a RFP Package. Within the Package, there can be as many sub-folders and documents as needed. 
3. Generate a Private Key & Certificate
  -Generate Private Key:
    - Command: openssl genrsa -out graph-app.key 2048
    - graph-app.key -> Private key (keep this secret)
  - Create a self-signed certificate:
    - Command: openssl req -new -x509 -key graph-app.key -out graph-app.crt -days 365
      - Prompted for fields: Common Name (CN) - app name
      - This is your Public certificate that you need to upload into Entra ID when registering this client app
4. Register this client app in Microsoft's Entra ID UI - https://entra.microsoft.com/
  - Register this client app: App registration - Entra ID will provide a client ID and tenant ID
  - Upload Certificate: App registration -> your app -> Certificates & Secrets -> Certificates -> Upload certificate (graph-app.crt)
  - Grant client app permissions to Microsoft Graph: API permissions -> Add a permission -> Microsoft Graph -> Application permissions
      - Choose Sites.ReadWrite.All -> Click "Grant admin consent" (this step only requests permission, does not provide the permission)
      - Will need a Entra ID Admin to approve this Sites.ReadWrite.All for the specified SharePoint Site. Will need to provide the Entra ID Admin the SharePoint Site ID.
  - Have a Entra ID Admin grant the permission. Tell your Entra ID Admin that you need to register an Entra ID app for a backend service. The app requires Microsoft Graph application permission Sites.Selected. Once consented, the app must be granted read-write access to a single SharePoint site only. No user-delegated permissions are needed.
5. Create a Smartsheet with the following columns (in this order): Date Parsed, Year, Business Unit, Division, RFP Package Name, KPI Name, KPI Category, KPI Context  
6. Create the following environment variables and add into Azure Key Vault:
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
7. In target/smartsheet_post.go, update the const variables by entering the column ID's for each column in your smartsheet. Will need to obtain these column ID's using a curl request to your Smartsheet using the SMARTSHEET_TOKEN and SMARTSHEET_URL
8. In parser/parser_xlsx.go, update the "sharedStringsFilePath" constant variable. This is a temporary file that is created and deleted for larger xlsx files during the parsing process
9. In parser/kpiDefinitions.json, update this file to include the KPI's you would like to parse for inside of docx and xlsx files
10. Run program from this applications root directory: go run ./

