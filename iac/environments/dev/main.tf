terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = ">= 3.11.0"
    }
  }
}

provider "azurerm" {
  features {}
}

locals {
  location = "westeurope"
  suffix   = "filario-garden"
}

module "aks_rg" {
  source   = "../../modules/azure/rg"
  location = local.location
  suffix   = local.suffix
}

module "aks" {
  source   = "../../modules/azure/aks"
  location = local.location
  suffix   = local.suffix
  rg_name  = module.aks_rg.rg_name

  depends_on = [module.aks_rg]
}
