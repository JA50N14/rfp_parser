# Request For Proposal Parser

## Overview
The RFP Parser is an application that traverses a Microsoft SharePoint Document Library, downloads .docx, .xlsx, and .pdf files within each Request for Proposal (RFP) package, extracts key performance indicators (KPIs), and posts the parsed data into Smartsheet. Within Smartsheet, dashboards can be easily created to visualize this data.

This application is designed to run on Azure Container App Job. Familiarity with software development and Azure is recommended for successful deployment.


## Motivation
The tool provides analytics into client priorities, such as specific ISO certifications or implemented programs. By analyzing RFP documents, leadership can prioritize which certifications or programs to pursue, improving the company's ability to win bids.


## Prerequisites
Before beginning, ensure you have the following:
- Software Development Knowledge – familiarity with Go, Docker, and scripting.
- Azure Knowledge – access to a subscription, resource groups, and the ability to deploy Container App Jobs.
- Azure CLI Installed – to run deployment commands.
- Docker Installed – to build and push images to Azure Container Registry.
- Smartsheet Account – to store KPI data and generate access tokens.
- Entra ID Admin Access – to register applications and grant Microsoft Graph permissions.

## 🚀 Setup - Part 1: Local Configuration
1. Clone this repository to your local machine
  - git clone https://github.com/JA50N14/rfp_parser.git

2. Create a SharePoint Site & Library
  - Within the Document Library have the following directory tree set up:
  Year (e.g., 2025, 2026)
  └─ Business Unit (e.g., "Facilities Management")
      └─ Division (e.g., "FM East", "FM West")
          └─ RFP Packages (directories representing each RFP Package)
 
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

7. Define KPIs
  - Update parser/kpiDefinitions.json to include the KPIs to parse from .docx, .xlsx, and .pdf files.


## 🚀 Setup - Part 2: Deploy Application on Azure
1. Subscription & Resource Providers
  - Ensure the following providers exist:
    - Microsoft.Web, Microsoft.Storage, Microsoft.ContainerRegistry, Microsoft.Insights, Microsoft.ManagedIdentity, Microsoft.App, Microsoft.OperationalInsights
  - Add any missing providers if prompted.

2. Create Resource Group
  - Logical container for all resources
  - Require "Owner" access 
  - Example name: rfp-parser

3. Set Environment Variables
  - Open a bash shell session - Need to run commands to create and configure Azure Resources.
  - Set Variables:
    # Resource Identifiers
    - RG="rfp-parser" # Resource Group
    - ENV="managedenvironment-rfpparser" # Container Apps Envrionment name
    - JOB="rfpparsercontainerappjob" # Name of your Container App Job
    # Container Registry and Image
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
  - Additional variables will be set throughout this process

4. Create Log Analytics Workspace
  az monitor log-analytics workspace create --resource-group $RG --workspace-name $LOG_ANALYTICS_NAME --location $LOCATION
  - This enables persistent log storage for container execution logs.

5. Create Container Apps Environment
  # Get Log Analytic Credentials
  LOG_ANALYTICS_ID=$(az monitor log-analytics workspace show --resource-group $RG --workspace-name $LOG_ANALYTICS_NAME --query customerId -o tsv)
  LOG_ANALYTICS_KEY=$(az monitor log-analytics workspace get-shared-keys --resource-group $RG --workspace-name $LOG_ANALYTICS_NAME --query primarySharedKey -o tsv)
  # Create the Container Apps Environment
  az containerapp env create --name $ENV_NAME --resource-group $RG --location $LOCATION --logs-workspace-id $LOG_ANALYTICS_ID --logs-workspace-key $LOG_ANALYTICS_KEY

6. Create Container Registry (ACR)
  # Create the ACR
  az acr create --resource-group $RG --name $ACR_NAME --sku $SKU --location $LOCATION --admin-enabled false
  # Get the login server - This must match EXACTLY what you use in Docker tags and Container App Jobs image
  ACR_LOGIN_SERVER=$(az acr show --name $ACR_NAME --query loginServer -o tsv)
  # Login to ACR (for pushing image)
  az acr login --name $ACR_NAME
  # Get ACR resource ID - Used later to allow the Container App Job to pull images securely
  ACR_RESOURCE_ID=$(az acr show --name $ACR_NAME --query id -o tsv)

7. Build and Push your Docker Image to ACR
    # Build Docker image
    docker build --no-cache -t $ACR_LOGIN_SERVER/$IMAGE_NAME:$IMAGE_TAG .
    # Push Docker image
    docker push $ACR_LOGIN_SERVER/$IMAGE_NAME:$IMAGE_TAG
    - Explanation: We're building the Go binary and packaging it with runtime dependencies (like poppler-utils) into a container. Then we push the image to Azure Container Registry so the job can pull it.

8. Create Container App Job
  az containerapp job create --name $JOB --resource-group $RG --environment $ENV --trigger-type Schedule --cron-expression "$CRON_EXPR" --image mcr.microsoft.com/k8se/quickstart:latest --cpu $CPU --memory $MEMORY --replica-timeout $REPLICA_TIMEOUT --replica-retry-limit $REPLICA_RETRY_LIMIT --system-assigned
  - Explanation: This creates a scheduled Azure Container App Job resource. It pulls the Docker image from ACR and sets up resource limits and retry behaviour.

9. Assign AcrPull role
# Get the job's system-assigned principal ID
  PRINCIPAL_ID=$(az containerapp job show --name $JOB --resource-group $RG --query identity.principalId -o tsv)
  # Get the ACR resource ID
  ACR_RESOURCE_ID=$(az acr show --name $ACR_NAME --query id -o tsv)
  # Assign AcrPull role
  az role assignment create --assignee $PRINCIPAL_ID --role AcrPull --scope $ACR_RESOURCE_ID

10. Add Secrets to Container App Job
  az containerapp job secret set --name $JOB --resource-group $RG --secrets graph-private-key="$GRAPH_PRIVATE_KEY" graph-certificate="$GRAPH_CERTIFICATE" smartsheet-token="$SMARTSHEET_TOKEN"

11. Set Environment Variables for Container App Job
  az containerapp job env set --name $JOB --resource-group $RG --env-vars \ 
    GRAPH_PRIVATE_KEY=secretref:graph-private-key \
    GRAPH_CERTIFICATE=secretref:graph-certificate \
    SMARTSHEET_TOKEN=secretref:smartsheet-token \
    SMARTSHEET_URL="$SMARTSHEET_URL" \
    GRAPH_CLIENT_ID="$GRAPH_CLIENT_ID" \
    GRAPH_TENANT_ID="$GRAPH_TENANT_ID" \
    GRAPH_SITE_ID="$GRAPH_SITE_ID" \
    GRAPH_LIBRARY_NAME="$GRAPH_LIBRARY_NAME" \
    GRAPH_DRIVE_ID='$GRAPH_DRIVE_ID' \
    SHAREPOINT_LIST_ID="$SHAREPOINT_LIST_ID"
  - Explanation: Creating environment variables that will be availble to the container app.

12. Start Manual Execution (testing)
  EXECUTION=$(az containerapp job start --name $JOB --resource-group $RG --query name -o tsv)

  echo "Started execution: $EXECUTION"
  - Explanation: This triggers the job immediately, instead of waiting for the cron schedule. Good for testing. This command also captures the execution ID, which is used in the next step to view logs.

13. View Logs
  az containerapp job logs show --name $JOB --resource-group $RG --execution $EXECUTION --container $JOB --follow


## Maintenance - Viewing Logs in Log Analytics
- Go to Log Analytic workspace resource -> Logs -> KQL mode
  - Application logs (parsed output):
  ContainerAppConsoleLogs_CL
| where ContainerJobName_s == "rfpparsercontainerappjob"
| sort by TimeGenerated desc

  - System logs (startup failures, container crashes, pull errors):
  ContainerAppSystemLogs_CL
| where JobName_s == "rfpparsercontainerappjob"
| sort by TimeGenerated desc


## Maintenance - Updating Job with New Image Version
1. Build new Docker image (v2, v3, ...)
- Set your variables:
ACR_LOGIN_SERVER=<your-acr-login-server>   # e.g., myregistry.azurecr.io
IMAGE_NAME=rfp-parser
IMAGE_TAG=v2
- Build the Docker image:
docker build --no-cache -t $ACR_LOGIN_SERVER/$IMAGE_NAME:$IMAGE_TAG .
2. Login to Azure Container Registry
az acr login --name <your-acr-name>
3. Push image to ACR
docker push $ACR_LOGIN_SERVER/$IMAGE_NAME:$IMAGE_TAG
4. Verify image is listed
az acr repository show-tags --name <your-acr-name> --repository $IMAGE_NAME --output table
5. Update Container App Job to use new image
az containerapp job update --name <job-name> --resource-group <rg-name> --image $ACR_LOGIN_SERVER/$IMAGE_NAME:$IMAGE_TAG


## Maintenance - Upload New Certificate & Private Key
1. Generate a public-private key pair
  - Command: openssl genrsa -out graph-app.key 2048
    - graph-app.key -> Private key (keep this secret)
  - Command: openssl req -new -x509 -key graph-app.key -out graph-app.crt -days 365
    - Prompted for fields: Common Name (CN) -> rfp_parser
2. Login to entra.microsoft.com
3. "App registration" -> Select "rfp_parser" -> "Certificates & secrets"
4. Upload new public key/certificate and remove old public key/certificate
5. Update GRAPH_PRIVATE_KEY and GRAPH_CERTIFICATE environment variables in Container App Job


## Contributing
- Clone the repo
  - bash: git clone https://github.com/JA50N14/rfp_parser.git
