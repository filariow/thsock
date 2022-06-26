output "rg_name" {
  sensitive   = false
  description = "Name of the resource group"
  value       = azurerm_resource_group.rg.name
}
