import { from, merge } from 'rxjs'
import { toArray, switchMap, filter } from 'rxjs/operators'
import semver from 'semver'
import { isDefined, propertyIsDefined } from '../../../../../../shared/src/util/types'
import { createExecServerClient } from '../../execServer/client'
import { memoizedFindTextInFiles } from '../../util'
import { ResolvedDependency, PackageJsonDependencyManagementProvider } from '../packageManager'
import { editForCommands2 } from '../packageManagerCommon'
import { lockTree } from './logicalTree'
import { provideDependencySpecification } from '../util'

const npmExecClient = createExecServerClient('a8n-npm-exec')

const NPM_OPTS = ['--no-audit', '--package-lock-only', '--ignore-scripts']

export const npmPackageManager: PackageJsonDependencyManagementProvider = {
    type: 'npm',
    provideDependencySpecifications: (query, filters = '') => {
        const parsedQuery = {
            ...query,
            parsedVersionRange: new semver.Range(query.versionRange),
        }
        return from(
            memoizedFindTextInFiles(
                {
                    pattern: `'"${parsedQuery.name}"' ${filters}`,
                    type: 'regexp',
                },
                {
                    repositories: {
                        type: 'regexp',
                    },
                    files: {
                        includes: ['(^|/)package-lock.json$'],
                        excludes: ['node_modules'],
                        type: 'regexp',
                    },
                    maxResults: 99999999, // TODO!(sqs): un-hardcode
                }
            )
        ).pipe(
            switchMap(textSearchResults =>
                merge(
                    ...textSearchResults.map(textSearchResult =>
                        provideDependencySpecification(
                            new URL(textSearchResult.uri.replace(/package-lock\.json$/, 'package.json')),
                            new URL(textSearchResult.uri),
                            parsedQuery,
                            getPackageLockDependency
                        )
                    )
                ).pipe(
                    filter(isDefined),
                    toArray()
                )
            )
        )
    },
    resolveDependencyUpgradeAction: (dep, version) => {
        // TODO!(sqs): this is not correct w.r.t. indirect deps
        if (dep.declarations.length !== 1) {
            console.error('Invalid declarations.', dep)
            throw new Error('invalid declarations')
        }
        return editForCommands2(
            [
                ...dep.declarations.map(d => d.location.uri),
                ...dep.resolutions.filter(propertyIsDefined('location')).map(d => d.location.uri),
            ],
            [['npm', 'install', ...NPM_OPTS, '--', `${dep.declarations[0].name}@${version}`]],
            npmExecClient
        )
    },
    resolveDependencyBanAction: dep => {
        // TODO!(sqs): this is not correct w.r.t. indirect deps
        if (dep.declarations.length !== 1) {
            console.error('Invalid declarations.', dep)
            throw new Error('invalid declarations')
        }
        return editForCommands2(
            [
                ...dep.declarations.map(d => d.location.uri),
                ...dep.resolutions.filter(propertyIsDefined('location')).map(d => d.location.uri),
            ],
            [['npm', 'uninstall', ...NPM_OPTS, '--', `${dep.declarations[0].name}`]],
            npmExecClient
        )
    },
}

function getPackageLockDependency(
    packageJson: string,
    packageLock: string,
    packageName: string
): ResolvedDependency | null {
    const tree = lockTree(JSON.parse(packageJson), JSON.parse(packageLock))
    let found: any
    // eslint-disable-next-line ban/ban
    tree.forEach((dep: any, next: any) => {
        if (dep.name === packageName) {
            found = dep
        } else {
            // eslint-disable-next-line callback-return
            next()
        }
    })
    return found ? { name: packageName, version: found.version, direct: !!tree.getDep(packageName) } : null
}
