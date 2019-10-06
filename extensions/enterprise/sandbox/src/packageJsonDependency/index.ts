import { Unsubscribable } from 'rxjs'
import semver from 'semver'
import { DependencyManagementCampaignContextCommon } from '../dependencyManagement/common'
import { registerDependencyManagementProviders } from '../dependencyManagement/register'
import { combinedProvider } from '../dependencyManagement/combinedProvider'
import { DependencyManagementProvider, DependencyQuery } from '../dependencyManagement'
import { yarnDependencyManagementProvider } from './yarn/provider'
import { npmDependencyManagementProvider } from './npm/provider'

export interface PackageJsonDependencyCampaignContext extends DependencyManagementCampaignContextCommon {}

export interface PackageJsonDependencyQuery extends Required<DependencyQuery> {
    parsedVersionRange: semver.Range
}

export interface PackageJsonDependencyManagementProvider
    extends DependencyManagementProvider<PackageJsonDependencyQuery> {}

export function register(): Unsubscribable {
    return registerDependencyManagementProviders(
        'packageJsonDependency',
        combinedProvider([npmDependencyManagementProvider, yarnDependencyManagementProvider]),
        context => {
            if (typeof context.packageName !== 'string') {
                throw new Error('invalid packageName')
            }
            if (typeof context.matchVersion !== 'string') {
                throw new Error('invalid matchVersion')
            }
            return {
                name: context.packageName,
                versionRange: context.matchVersion,
                parsedVersionRange: new semver.Range(context.matchVersion),
            }
        }
    )
}
