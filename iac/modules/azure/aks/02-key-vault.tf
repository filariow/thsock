data "azurerm_client_config" "current" {}

resource "azurerm_key_vault" "aks_kv" {
  name                        = "kv-aks-${var.suffix}"
  location                    = var.location
  resource_group_name         = var.rg_name
  tenant_id                   = data.azurerm_client_config.current.tenant_id
  sku_name                    = "premium"
  enabled_for_disk_encryption = true

  #checkov:skip=CKV_AZURE_42:suppressing purge protection for development purposes
  #checkov:skip=CKV_AZURE_110:suppressing purge protection for development purposes
  purge_protection_enabled    = false

  network_acls {
    default_action = "Deny"
    bypass         = "AzureServices"
    ip_rules       = [data.http.ip.body]
  }
}

resource "azurerm_key_vault_key" "aks_disk" {
  #checkov:skip=CKV_AZURE_112:suppressing HSM requirement
  name         = "des-key-aks-disk"
  key_vault_id = azurerm_key_vault.aks_kv.id
  key_type     = "RSA"
  key_size     = 2048

  expiration_date = "2024-06-26T16:00:00Z"

  key_opts = [
    "decrypt",
    "encrypt",
    "sign",
    "unwrapKey",
    "verify",
    "wrapKey",
  ]

  depends_on = [azurerm_key_vault.aks_kv, azurerm_key_vault_access_policy.current_user]
}

resource "azurerm_disk_encryption_set" "aks_disk" {
  name                = "des-aks-disk"
  resource_group_name = var.rg_name
  location            = var.location
  key_vault_key_id    = azurerm_key_vault_key.aks_disk.id

  identity {
    type = "SystemAssigned"
  }

  depends_on = [azurerm_key_vault_access_policy.current_user]
}

resource "azurerm_key_vault_access_policy" "aks_disk" {
  key_vault_id = azurerm_key_vault.aks_kv.id

  tenant_id = azurerm_disk_encryption_set.aks_disk.identity.0.tenant_id
  object_id = azurerm_disk_encryption_set.aks_disk.identity.0.principal_id

  key_permissions = [
    "Create",
    "Delete",
    "Get",
    "Purge",
    "Recover",
    "Update",
    "List",
    "Decrypt",
    "Sign",
    "WrapKey",
    "UnwrapKey"
  ]
}

resource "azurerm_key_vault_access_policy" "current_user" {
  key_vault_id = azurerm_key_vault.aks_kv.id

  tenant_id = data.azurerm_client_config.current.tenant_id
  object_id = data.azurerm_client_config.current.object_id

  key_permissions = [
    "Create",
    "Delete",
    "Get",
    "Purge",
    "Recover",
    "Update",
    "List",
    "Decrypt",
    "Sign"
  ]
}

