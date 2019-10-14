import { Unsubscribable } from 'rxjs'
import semver from 'semver'
import { packageJsonDependencyManagementProviderRegistry } from './providers'
import { DependencyManagementCampaignContextCommon } from '../dependencyManagement/common'
import { registerDependencyManagementProviders } from '../dependencyManagement/register'

export interface PackageJsonDependencyCampaignContext extends DependencyManagementCampaignContextCommon {}

export function register(): Unsubscribable {
    return registerDependencyManagementProviders(
        'packageJsonDependency',
        packageJsonDependencyManagementProviderRegistry,
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
