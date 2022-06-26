resource "azurerm_resource_group" "rg" {
  name     = "rg-${var.suffix}"
  location = var.location
}
