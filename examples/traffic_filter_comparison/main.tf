# =============================================================================
# Traffic Filter Comparison: Non-Serverless vs Serverless
# =============================================================================
# This file demonstrates how traffic filters work in both deployment types.
#
# NOTE ON UI vs API MODEL:
# The Elastic Cloud UI shows traffic filters with an option to "attach projects"
# to them (filter-centric view). However, the API models this the opposite way:
# projects have a `traffic_filters` attribute containing filter IDs (project-centric).
# Both approaches achieve the same result - this provider follows the API model.
# =============================================================================

# -----------------------------------------------------------------------------
# OPTION A: NON-SERVERLESS DEPLOYMENTS (ec_deployment)
# -----------------------------------------------------------------------------
# There are TWO ways to associate traffic filters with deployments:
#
# A1. Inline on the deployment resource (recommended for full Terraform control)
# A2. Separate association resource (for deployments managed outside Terraform)
# -----------------------------------------------------------------------------

# --- A1: Inline Traffic Filter Association ---
# Use this when the deployment is fully managed by Terraform.
# The traffic_filter attribute directly on ec_deployment manages the full set.

data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "us-east-1"
}

resource "ec_deployment_traffic_filter" "deployment_allow_office" {
  name   = "Allow office IP"
  region = "us-east-1"
  type   = "ip"

  rule {
    source      = "192.168.1.0/24"
    description = "Office network"
  }
}

resource "ec_deployment_traffic_filter" "deployment_allow_vpn" {
  name   = "Allow VPN"
  region = "us-east-1"
  type   = "ip"

  rule {
    source      = "10.0.0.0/8"
    description = "VPN network"
  }
}

resource "ec_deployment" "example_inline" {
  name                   = "deployment-with-inline-filters"
  region                 = "us-east-1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-io-optimized-v2"

  # Inline association: Terraform manages ALL traffic filters for this deployment
  traffic_filter = [
    ec_deployment_traffic_filter.deployment_allow_office.id,
    ec_deployment_traffic_filter.deployment_allow_vpn.id,
  ]

  elasticsearch = {
    hot = {
      autoscaling = {}
    }
  }

  kibana = {}
}

# --- A2: Separate Association Resource ---
# Use this when the deployment exists outside of Terraform, or you want to
# manage traffic filter associations independently.
# NOTE: Cannot mix with inline traffic_filter on the same deployment!

data "ec_deployment" "existing" {
  id = "320b7b540dfc967a7a649c18e2fce4ed" # Pre-existing deployment ID
}

resource "ec_deployment_traffic_filter" "separate_filter" {
  name   = "Separate filter for existing deployment"
  region = "us-east-1"
  type   = "ip"

  rule {
    source = "203.0.113.0/24"
  }
}

resource "ec_deployment_traffic_filter_association" "example" {
  # Links ONE filter to ONE deployment - need multiple resources for multiple filters
  traffic_filter_id = ec_deployment_traffic_filter.separate_filter.id
  deployment_id     = data.ec_deployment.existing.id
}


# -----------------------------------------------------------------------------
# OPTION B: SERVERLESS PROJECTS (ec_elasticsearch_project, etc.)
# -----------------------------------------------------------------------------
# There are TWO ways to associate traffic filters with serverless projects:
#
# B1. Inline on the project resource (recommended for full Terraform control)
# B2. Separate association resource (for projects managed outside Terraform)
# -----------------------------------------------------------------------------

# --- AWS Region Example ---

resource "ec_serverless_traffic_filter" "serverless_allow_office_aws" {
  name   = "Allow office IP (serverless AWS)"
  region = "aws-us-east-1"
  type   = "ip"

  rule {
    source      = "192.168.1.0/24"
    description = "Office network"
  }
}

resource "ec_serverless_traffic_filter" "serverless_allow_vpn_aws" {
  name   = "Allow VPN (serverless AWS)"
  region = "aws-us-east-1"
  type   = "ip"

  rule {
    source      = "10.0.0.0/8"
    description = "VPN network"
  }
}

resource "ec_elasticsearch_project" "my_project_aws" {
  name      = "my-serverless-project-aws"
  region_id = "aws-us-east-1"

  # Inline association only - set of traffic filter IDs
  traffic_filters = [
    ec_serverless_traffic_filter.serverless_allow_office_aws.id,
    ec_serverless_traffic_filter.serverless_allow_vpn_aws.id,
  ]
}

# --- GCP Region Example ---
# Serverless supports AWS, GCP, and Azure regions

resource "ec_serverless_traffic_filter" "serverless_allow_office_gcp" {
  name   = "Allow office IP (serverless GCP)"
  region = "gcp-us-central1"
  type   = "ip"

  rule {
    source      = "192.168.1.0/24"
    description = "Office network"
  }
}

resource "ec_elasticsearch_project" "my_project_gcp" {
  name      = "my-serverless-project-gcp"
  region_id = "gcp-us-central1"

  traffic_filters = [
    ec_serverless_traffic_filter.serverless_allow_office_gcp.id,
  ]
}

# --- Azure Region Example ---

resource "ec_serverless_traffic_filter" "serverless_allow_office_azure" {
  name   = "Allow office IP (serverless Azure)"
  region = "azure-eastus2"
  type   = "ip"

  rule {
    source      = "192.168.1.0/24"
    description = "Office network"
  }
}

resource "ec_observability_project" "my_obs_project" {
  name      = "my-observability-project"
  region_id = "azure-eastus2"

  traffic_filters = [
    ec_serverless_traffic_filter.serverless_allow_office_azure.id,
  ]
}

resource "ec_security_project" "my_sec_project" {
  name      = "my-security-project"
  region_id = "aws-us-east-1"

  traffic_filters = [
    ec_serverless_traffic_filter.serverless_allow_vpn_aws.id,
  ]
}

# --- B2: Separate Association Resource ---
# Use this when you want to manage traffic filter associations as separate resources.
# This allows you to define filters and attach them from the project side of your config.

resource "ec_serverless_traffic_filter" "separate_filter_serverless" {
  name   = "Separate filter for association"
  region = "aws-us-east-1"
  type   = "ip"

  rule {
    source = "203.0.113.0/24"
  }
}

resource "ec_serverless_traffic_filter_association" "example" {
  traffic_filter_id = ec_serverless_traffic_filter.separate_filter_serverless.id
  project_id        = ec_elasticsearch_project.my_project_aws.id
  project_type      = "elasticsearch" # elasticsearch, observability, or security
}


# =============================================================================
# COMPARISON SUMMARY
# =============================================================================
#
# | Aspect                    | Non-Serverless (ec_deployment)       | Serverless (ec_*_project)              |
# |---------------------------|--------------------------------------|----------------------------------------|
# | Traffic Filter Resource   | ec_deployment_traffic_filter         | ec_serverless_traffic_filter           |
# | Inline Association        | traffic_filter = [...]               | traffic_filters = [...]                |
# | Separate Association      | ec_deployment_traffic_filter_assoc.  | ec_serverless_traffic_filter_assoc.    |
# | Association Type          | List of filter IDs                   | Set of filter IDs                      |
# | Region Format             | "us-east-1"                          | "aws-us-east-1", "gcp-us-central1"     |
#
# =============================================================================
