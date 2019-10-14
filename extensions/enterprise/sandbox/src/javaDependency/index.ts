import { Unsubscribable } from 'rxjs'
import { javaDependencyManagementProviderRegistry } from './providers'
import { DependencyManagementCampaignContextCommon } from '../dependencyManagement/common'
import { registerDependencyManagementProviders } from '../dependencyManagement/register'
import { DependencyQuery, DependencyManagementProvider } from '../dependencyManagement'
import semver from 'semver'
import { combinedProvider } from '../dependencyManagement/combinedProvider'

// TODO!(sqs): https://github.com/kevcodez/gradle-upgrade-interactive/blob/master/ReplaceVersion.js

export interface JavaDependencyCampaignContext extends DependencyManagementCampaignContextCommon {}

export interface JavaDependencyQuery extends Required<DependencyQuery> {
    parsedVersionRange: semver.Range
}

export interface JavaDependencyManagementProvider extends DependencyManagementProvider<JavaDependencyQuery> {}

export function register(): Unsubscribable {
    return registerDependencyManagementProviders(
        'javaDependency',
        combinedProvider([gradleDependencyManagementProvider]),
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
