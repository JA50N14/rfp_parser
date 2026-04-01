# Request For Proposal Parser

## Overview
This application walks a Microsoft SharePoint Document Library to download docx, xlsx, and pdf files within each Request for Proposal (RFP) Package, parses for KPI's, and then posts the data into Smartsheets. Within Smartsheets, you can easily create dashboards based off this data. This applications architecture is designed to run on Azure Container App Job. 


## Motivation
Our organization wanted analytics into what is important to our clients (i.e. specific ISO certifications, programs implemented, etc.). This program provides these client analytics, so the leadership team can prioritize which certifications or programs to pursue. The overall purpose of this application is to provide analytics to better position the company to win bids.


## 🚀 Setup - Part 1
1. Clone this repository to your local machine

2. Create a SharePoint site. Within the Document Library have the following directory tree set up:
  - Year (i.e. "2025", "2026", etc.)
    - Business Unit (i.e. "Facilities Management", etc.)
      - Division (i.e. "FM East", "FM West", etc.)
        - RFP Packages - Where your RFP Packages are located. Each directory needs to represent a RFP Package. 
  - Create a dropdown type column named "ProcessStatus" with "InProgress", "Complete", "Failed" as selection options

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

6. In walk/result_to_smartsheet_transform.go, update the const variables by entering the column ID's for each column in your smartsheet. Will need to obtain these column ID's using a curl request to your Smartsheet using the SMARTSHEET_TOKEN and SMARTSHEET_URL

7. In parser/kpiDefinitions.json, update this file to include the KPI's you would like to parse for inside of docx, xlsx, and pdf files


## 🚀 Setup - Part 2: Required Azure Resources
Before creating a Container App Job resource, you need the following Azure resources created:
1. Subscription
  - Create a subscription with the following resource providers added (if not already created):
    - Microsoft.Web, Microsoft.Storage, Microsoft.ContainerRegistry, Microsoft.Insights, Microsoft.ManagedIdentity, Microsoft.App, Microsoft.OperationalInsights
  - If unable to create a resource (i.e. Azure Container Registry, etc.), it is likely because you are missing resource providers on your subscription. The error message will tell you which resource providers your subscription is missing, so you can add them.

2. Resource Group
  - Logical container for all resources
  - Create a Resource Group before creating other Azure resources
  - Need "Owner" level access 
  - Example: rfp-parser

3. Azure Container Registry (ACR)
  - Stores your Docker Image
  - Must have a unique registry name
  - Configure network access as needed (public)
  - Example: rfpparserregistry

4. Azure Container Apps Environment
  - Managed environment where Container Apps run
  - Provides networking, scaling, and logs infrastructure
  - Example: managedEnvironmnet-rfpparser

5. Azure Log Analytics Workspace
  - Logs from the parser binary go to stdout. Azure captures them automatically, but they are ephemeral and stored temporarily.
  - This resource allows you to persist logs long-term and view them in the Azure Portal
  - Need to connect this resource to your Azure Container Apps Environment resource for persistent log storage
  - This resource is not required, but recommended


## 🚀 Setup - Part 3: Create a Azure Container App Job
1. Open a bash terminal and set variables
  # Resource Identifiers
  - RG="rfp-parser" # Resource Group
  - ENV="managedEnvironment-rfpparser" # Container Apps Envrionment name
  - JOB="rfpparsercontainerappjob" # Name of your Container App Job
  # Container Registry and Image
  - ACR_NAME="rfpparserregistrya8d7ecctd7bsgfbj" # Must match existing ACR registry name
  - ACR_LOGIN_SERVER="$ACR_NAME.azurecr.io"
  - IMAGE_NAME="rfp-parser" # Docker Image name
  - IMAGE_TAG="v1" # Initial Docker version tag
  # Resources for Container App Job
  - CPU="0.5"
  - MEMORY="1Gi"
  - REPLICA_TIMEOUT="1800" # seconds
  - REPLICA_RETRY_LIMIT="1"
  # Cron schedule
  - CRON_EXPR="0 1 * * 0" #Runs Sunday at 1AM
  # Environment / Secrets
  - SMARTSHEET_TOKEN - A Smartsheet access token that can be generated in Smartsheet
  - SMARTSHEET_URL - The URL of the Smartsheet to POST the KPI data
  - GRAPH_PRIVATE_KEY - Your Private Key
  - GRAPH_CERTIFICATE - Your certificate
  - GRAPH_CLIENT_ID - The Client ID provided via Entra ID UI
  - GRAPH_TENANT_ID - The Tenant ID provided via Entra ID UI
  - GRAPH_SITE_ID - The ID of the SharePoint Site where the Document Library to be walked resides
  - GRAPH_LIBRARY_NAME - The name of the Document Library to walk
  - GRAPH_DRIVE_ID - The Drive ID of the Document Library to walk
  - SHAREPOINT_LIST_ID - The List ID of the Document Library to walk
  - Explanation: These variables keep commands short and easy to update.

2. Build and push your Docker Image to ACR
  # Log in to ACR
  az acr login --name $ACR_NAME
  # Build Docker image
  docker build -t $ACR_LOGIN_SERVER/$IMAGE_NAME:$IMAGE_TAG .
  # Push Docker image
  docker push $ACR_LOGIN_SERVER/$IMAGE_NAME:$IMAGE_TAG
  - Explanation: We're building the Go binary and packaging it with runtime dependencies (like poppler-utils) into a container. Then we push the image to Azure Container Registry so the job can pull it.

3. Create the Container App Job
  az containerapp job create \
    --name $JOB \
    --resource-group $RG \
    --environment $ENV \
    --trigger-type Schedule \
    --cron-expression "$CRON_EXPR" \
    --image "$ACR_LOGIN_SERVER/$IMAGE_NAME:$IMAGE_TAG" \
    --cpu $CPU \
    --memory $MEMORY \
    --replica-timeout $REPLICA_TIMEOUT \
    --replica-retry-limit $REPLICA_RETRY_LIMIT
  - Explanation: This creates a scheduled Azure Container App Job resource. It pulls the Docker image from ACR and sets up resource limits and retry behaviour.

4. Assign AcrPull to the job's system-assigned identity
  # Get the job's system-assigned principal ID
  PRINCIPAL_ID=$(az containerapp job show \
    --name $JOB \
    --resource-group $RG \
    --query identity.principalId -o tsv)
  # Get the ACR resource ID
  ACR_RESOURCE_ID=$(az acr show \
    --name $ACR_NAME \
    --query id -o tsv)
  # Assign AcrPull role
  az role assignment create \
    --assignee $PRINCIPAL_ID \
    --role AcrPull \
    --scope $ACR_RESOURCE_ID
  - Explanation: The job uses a system-assigned managed identity to pull images from ACR. Without AcrPull permissions, the container cannot start. 

5. Add secrets to the job
  az containerapp job secret set \
    --name $JOB \
    --resource-group $RG \
    --secrets graph-private-key="$GRAPH_PRIVATE_KEY" \
              graph-certificate="$GRAPH_CERTIFICATE" \
              smartsheet-token="$SMARTSHEET_TOKEN"
  - Explanation: Secrets like API keys are stored securely and can be referenced in environment variables

6. Set environment variables
  az containerapp job env set \
    --name $JOB \
    --resource-group $RG \
    --env-vars  GRAPH_PRIVATE_KEY=secretref:graph-private-key \
                GRAPH_CERTIFICATE=secretref:graph-certificate \
                SMARTSHEET_TOKEN=secretref:smartsheet-token \
                SMARTSHEET_URL="$SMARTSHEET_URL"
                GRAPH_CLIENT_ID="$GRAPH_CLIENT_ID" \
                GRAPH_TENANT_ID="$GRAPH_TENANT_ID" \
                GRAPH_SITE_ID="$GRAPH_SITE_ID" \
                GRAPH_LIBRARY_NAME="$GRAPH_LIBRARY_NAME" \
                GRAPH_DRIVE_ID='$GRAPH_DRIVE_ID' \
                SHAREPOINT_LIST_ID="$SHAREPOINT_LIST_ID"
  - Explanation: Creating environment variables that will be availble to the container app.

7. Start a manual execution to test
  EXECUTION=$(az containerapp job start \
    --name $JOB \
    --resource-group $RG \
    --query name -o tsv)

  echo "Started execution: $EXECUTION"
  - Explanation: This triggers the job immediately, instead of waiting for the cron schedule. Good for testing. This command also captures the execution ID, which is used in the next step to view logs.

8. View logs from execution
  az containerapp job logs show \
    --name $JOB \
    --resource-group $RG \
    --execution $EXECUTION \
    --container $JOB \
    --follow
  - Explanation: You can now watch live logs from your Go binary. The slog JSON logs should appear here if the binary is running correctly.

9. Update the job with a new image version (future releases)
# Update the job to use a new image version (i.e. v2)
  az containerapp job update \
    --name $JOB \
    --resource-group $RG \
    --image $ACR_LOGIN_SERVER/$IMAGE_NAME:v2
  - Explanation: After pushing a new image to ACR, this command updates the job to point to the new image.


## Maintenance - Upload Certificate & Private Key Instructions (when Certificate/Private Key Expire)
1. Generate a public-private key pair
  - Command: openssl genrsa -out graph-app.key 2048
    - graph-app.key -> Private key (keep this secret)
  - Command: openssl req -new -x509 -key graph-app.key -out graph-app.crt -days 365
    - Prompted for fields: Common Name (CN) -> rfp_parser
2. Log into entra.microsoft.com
3. "App registration" -> Select "rfp_parser" -> "Certificates & secrets"
4. Upload the new public key/certificate and remove the old public key/certificate
5. Update the GRAPH_PRIVATE_KEY environment variable with the new private key
6. Create a new docker image, push to the Azure Container Registry, and ensure the Azure Container App Job resource references the new docker image


## Contributing
- Clone the repo
  - bash: git clone https://github.com/JA50N14/rfp_parser.git



