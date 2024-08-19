// This is a placeholder file to trigger environment variable setting via New-TestResources.ps1

@description('The base resource name.')
param baseName string = resourceGroup().name

@description('Which Azure Region to deploy the resource to. Defaults to the resource group location.')
param location string = resourceGroup().location

@description('The principal to assign the role to. This is application object id.')
param testApplicationOid string
