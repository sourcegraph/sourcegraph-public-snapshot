import { DependencyQuery, DependencyManagementProvider } from '../dependencyManagement'
import semver from 'semver'
import { npmPackageManager } from './npm/npm'
import { yarnPackageManager } from './yarn/yarn'
import { combinedProvider } from '../dependencyManagement/combinedProvider'

const PROVIDERS = [npmPackageManager, yarnPackageManager]

export const packageJsonDependencyManagementProviderRegistry = combinedProvider(PROVIDERS)

export interface PackageJsonDependencyQuery extends Required<DependencyQuery> {
    parsedVersionRange: semver.Range
}

export interface PackageJsonDependencyManagementProvider
    extends DependencyManagementProvider<PackageJsonDependencyQuery> {
    type: string
}
