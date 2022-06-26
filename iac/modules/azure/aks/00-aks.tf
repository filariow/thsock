resource "azurerm_kubernetes_cluster" "app" {
  name                = "aks-${var.suffix}"
  location            = var.location
  resource_group_name = var.rg_name
  dns_prefix          = "aks-${var.suffix}"
  node_resource_group = "${var.rg_name}-mc"

  #checkov:skip=CKV_AZURE_141:needed for easy development
  local_account_disabled = false

  api_server_authorized_ip_ranges = ["${data.http.ip.body}/32"]

  disk_encryption_set_id = azurerm_disk_encryption_set.aks_disk.id

  #checkov:skip=CKV_AZURE_115:"private cluster not needed"
  private_cluster_enabled = false

  azure_policy_enabled = true

  network_profile {
    network_plugin = "kubenet"
    network_policy = "calico"
  }

  auto_scaler_profile {
    balance_similar_node_groups      = false
    empty_bulk_delete_max            = "10"
    expander                         = "random"
    max_graceful_termination_sec     = "600"
    max_node_provisioning_time       = "15m"
    max_unready_nodes                = 3
    max_unready_percentage           = 45
    new_pod_scale_up_delay           = "0s"
    scale_down_delay_after_add       = "10m"
    scale_down_delay_after_delete    = "10s"
    scale_down_delay_after_failure   = "3m"
    scale_down_unneeded              = "10m"
    scale_down_unready               = "20m"
    scale_down_utilization_threshold = "0.5"
    scan_interval                    = "10s"
    skip_nodes_with_local_storage    = false
    skip_nodes_with_system_pods      = true
  }

  default_node_pool {
    name                = "default"
    min_count           = 1
    max_count           = 4
    enable_auto_scaling = true
    vm_size             = "Standard_B2s"
    max_pods            = 200
  }

  role_based_access_control_enabled = true

  identity {
    type         = "UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.aks_id.id]
  }

  #checkov:skip=CKV_AZURE_4:checkov does not handle azurerm 3.11.0 syntax for oms_agent
  oms_agent {
    log_analytics_workspace_id = azurerm_log_analytics_workspace.aks.id
  }

  depends_on = []

  lifecycle {}
}

## Key Vault Integration

resource "azurerm_key_vault_access_policy" "aks_to_shared" {
  key_vault_id = azurerm_key_vault.aks_kv.id
  tenant_id    = azurerm_key_vault.aks_kv.tenant_id
  object_id    = azurerm_user_assigned_identity.aks_id.principal_id

  secret_permissions = [
    "Get",
  ]

  certificate_permissions = [
    "Get",
  ]

  key_permissions = [
    "Get",
    "WrapKey",
    "UnwrapKey"
  ]
}

