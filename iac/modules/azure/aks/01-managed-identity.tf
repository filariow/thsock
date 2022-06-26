resource "azurerm_user_assigned_identity" "aks_id" {
  resource_group_name = var.rg_name
  location            = var.location

  name = "id-aks-${var.suffix}"
}
