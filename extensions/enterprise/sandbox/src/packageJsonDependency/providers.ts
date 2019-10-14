import { npmPackageManager } from './npm/npm'
import { yarnPackageManager } from './yarn/yarn'
import { combinedProvider } from '../dependencyManagement/combinedProvider'

const PROVIDERS = [npmPackageManager, yarnPackageManager]

export const packageJsonDependencyManagementProviderRegistry = combinedProvider(PROVIDERS)
