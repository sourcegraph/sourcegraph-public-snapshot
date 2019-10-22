import { Unsubscribable } from 'rxjs'
import { DependencyManagementCampaignContextCommon } from '../dependencyManagement/common'
import { registerDependencyManagementProviders } from '../dependencyManagement/register'
import { DependencyQuery, DependencyManagementProvider } from '../dependencyManagement'
import semver from 'semver'
import { combinedProvider } from '../dependencyManagement/combinedProvider'
import { gradleDependencyManagementProvider } from './gradle/provider'

// TODO!(sqs): https://github.com/kevcodez/gradle-upgrade-interactive/blob/master/ReplaceVersion.js

export interface JavaDependencyCampaignContext extends DependencyManagementCampaignContextCommon {
    supportMissingDependencyLock?: boolean
}

export interface JavaDependencyQuery extends Required<DependencyQuery> {
    parsedVersionRange: semver.Range
    supportMissingDependencyLock?: boolean
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
            if (
                context.supportMissingDependencyLock !== undefined &&
                typeof context.supportMissingDependencyLock !== 'boolean'
            ) {
                throw new Error('invalid supportMissingDependencyLock')
            }
            return {
                name: context.packageName,
                versionRange: context.matchVersion,
                parsedVersionRange: new semver.Range(context.matchVersion),
                supportMissingDependencyLock: context.supportMissingDependencyLock,
            }
        }
    )
}
