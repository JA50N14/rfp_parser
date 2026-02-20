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
      - Prompted for fields: Common Name (CN) - rfp_parser
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
8. In parser/kpiDefinitions.json, update this file to include the KPI's you would like to parse for inside of docx, xlsx and pdf files
9. Run program from this applications root directory: go run ./


## Microsoft Entra ID Cert/Private Key
1. Generate a public-private key pair
  - See step #3 in "Installation" section
2. Log into entra.microsoft.com
3. "App registration" -> Select "rfp_parser" -> "Certificates & secrets"
4. Upload the new public key/certificate and remove the old public key/certificate
5. Upload the new private key into Azure Functions Key Vault

## Deployment to Azure Functions Instructions
1. Build and Push Docker Image
  - Create Azure Container Registry (ACR)
  - Build your Image from project root
    - docker build -t yourRegistryName.azurecr.io/rfp-parser:1.0.0 .
      - Verify Dockerfile builds: go build -o function .
        - Must match "defaultExecutablePath": "function"
  - Push Image
    - docker push yourRegistryName.azurecr.io/rfp-parser:1.0.0
      - Image is now in Azure
2. Create the Function App (Container-Based)
 - az functionapp plan create --name rfp-plan --resource-group Your_RG --location westus2 --sku EP1 --is-linux
 - Create Function App Using Container
  - az functionapp create --name rfp-function --storage-account Your_storage --resource-group Your_RG --plan rfp-plan --deployment-container-image-name yourRegistryName.azurecr.io/rfp-parser:1.0.0 --functions-version 4
3. Allow Function to Pull from ACR
- Enable Managed Identity:
  - az functionapp identity assign --name rfp-function --resource-group Your_RG
  - Get the identity principal ID
  - Then grant pull access:
    - az role assignment create --assignee principal-id --scope $(az acr show --name yourRegistryName --query id --output tsv) --role AcrPull
    - Now Azure can securely pull your container without storing credentials
4. Configure App Settings (Environment Variables)
- In Azure Portal: Function App -> Settings -> Configuration -> Application Settings
  - Add:
    - SMARTSHEET_TOKEN
    - SMARTSHEET_URL
    - GRAPH_CLIENT_ID
    - GRAPH_TENANT_ID
    - GRAPH_SITE_ID
    - GRAPH_LIBRARY_NAME
    - GRAPH_DRIVE_ID
    - SHAREPOINT_LIST_ID
5. Create Azure Key Vault
- az keyvault create --name rfp-keyvault --resource-group Your_RG --location westus2
- Add Secrets
  - az keyvault secret set --vault-name rfp-keyvault --name GRAPH_PRIVATE_KEY --value "your-secret"
  - az keyvault secret set --vault-name rfp-keyvault --name GRAPH_CERTIFICATE --value "your-secret"
  - az keyvault secret set --vault-name rfp-keyvault --name SMARTSHEET_TOKEN --value "access-token"
6. Give Function Access to Key Vault
- Enable Managed Identity (should already be done)
- az role assignment create --asignee principal-id --role "Key Vault Secrets User" --scope $(az keyvault show --name rfp-keyvault --query id --output tsv)
  - Now your function can read secrets
7. Reference Key Vault Secrets in App Settings
- Add these secrets as value in: Function App -> Configuration -> Application Settings
  - GRAPH_PRIVATE_KEY = @Microsoft.KeyVault(SecretUri=https://rfp-keyvault.vault.azure.net/secrets/GRAPH_PRIVATE_KEY/)
  - GRAPH_CERTIFICATE = @Microsoft.KeyVault(SecretUri=https://rfp-keyvault.vault.azure.net/secrets/GRAPH_CERTIFICATE/)
  - SMARTSHEET_TOKEN = @Microsoft.KeyVault(SecretUri=https://rfp-keyvault.vault.azure.net/secrets/SMARTSHEET_TOKEN/)
8. Configure Timer Trigger
- Should already have rfp_parser_timer_function/function.json and host.json
9. Logging Configuration
- Auzre Functions automatically streams to stdout and stderr
- Go code uses slog.NewJSONHandler(os.Stdout, nil), so logging will work properly
- Logs go to Application Insights and Log Streams
- Enable Application Insights: Portal -> Function App -> Application Insights -> Turn On
10. Verify Deployment
- Go to Function App -> Functions
  - Ensure your timer function appears
- Check "Monitor" tab
- Check logs
  - Should see "Custom Handler listening on port ..."
  - At trigger time, should see "Timer fired!"


## How this App Works on Azure Functions - Using a Custom Handler
1. Azure decides if it needs an instance
- This happens when:
	- Timer fires and no instance exists
2. Container is created
- Azure pulls your container image (if not cached)
- Azure creates a container
- Azure mounts storage
- Azure starts the Function Host
3. Azure Functions Host Starts
- The Function runtime starts first
- Then reads host.json and all function.json files
- It identifies Timer Triggers and Custom Handler Configuration
4. Azure Starts Your Executable
- Because your host.json contains: "defaultExecutablePath": "function"
- Azure runs ./function
- Now your Go main() starts
- Your program reads FUNCTIONS_CUSTOMHANDLER_PORT
- Starts HTTP server
- Listens for request
- Container is now "warm"
5. Timer Fires (if that's why it started)
- Azure sends: POST http://localhost:{port}/ with the TimerTrigger payload
- Your handler runs

# Azure Function Plans
- Consumption Plan: 1-10+ seconds (cold start)
	- Scales to zero
	- After idle period (usually 20 min) - Azure stops the container, memory wiped, next trigger causes cold start
	- Execution timeout: 5 min (max 10 min)
- Premium Plan: Usually 0 seconds (if pre-warmed)
	- Container stays running, waits for next trigger, keeps memory in RAM, keeps HTTP server alive
	- Cold starts only for first deployment and after manual restart
	- Execution timeout: unlimited (configurable)
- Basic/Dedicated Plan: Only on first deploy or restart
	- Container stays alive, behaves like a normal always-on web app
	- Cold starts happens only when app restarts, you redeploy
	- Execution timeout: unlimited


