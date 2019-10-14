import { from, combineLatest, Observable, merge } from 'rxjs'
import { toArray, map, switchMap, filter } from 'rxjs/operators'
import semver from 'semver'
import * as sourcegraph from 'sourcegraph'
import { isDefined, propertyIsDefined } from '../../../../../../shared/src/util/types'
import { createExecServerClient } from '../../execServer/client'
import { memoizedFindTextInFiles } from '../../util'
import {
    ResolvedDependency,
    PackageJsonDependencyQuery,
    PackageJsonDependencyManagementProvider,
} from '../packageManager'
import { editForCommands2 } from '../packageManagerCommon'
import { lockTree } from './logicalTree'
import { DependencySpecification } from '../../dependencyManagement'

const npmExecClient = createExecServerClient('a8n-npm-exec')

const NPM_OPTS = ['--no-audit', '--package-lock-only', '--ignore-scripts']

const provideDependencySpecification = (
    result: sourcegraph.TextSearchResult,
    query: PackageJsonDependencyQuery & { parsedVersionRange: semver.Range }
): Observable<DependencySpecification<PackageJsonDependencyQuery> | null> => {
    const packageJson = from(
        sourcegraph.workspace.openTextDocument(new URL(result.uri.replace(/package-lock\.json$/, 'package.json')))
    )
    const packageLockJson = from(sourcegraph.workspace.openTextDocument(new URL(result.uri)))
    return combineLatest([packageJson, packageLockJson]).pipe(
        map(([packageJson, packageLockJson]) => {
            try {
                // TODO!(sqs): support multiple versions in lockfile/package.json
                const dep = getPackageLockDependency(packageJson.text!, packageLockJson.text!, name)
                if (!dep) {
                    return null
                }
                if (!semver.satisfies(dep.version, query.parsedVersionRange)) {
                    return null
                }
                const spec: DependencySpecification<PackageJsonDependencyQuery> = {
                    query,
                    declarations: [
                        {
                            name: dep.name,
                            // requestedVersion: // TODO!(sqs): get from package.json
                            direct: dep.direct,
                            location: { uri: new URL(packageJson.uri) },
                        },
                    ],
                    resolutions: [
                        {
                            name: dep.name,
                            resolvedVersion: dep.version,
                            location: { uri: new URL(packageLockJson.uri) },
                        },
                    ],
                }
                return spec
            } catch (err) {
                console.error(`Error checking lockfile and package.json for ${result.uri}.`, err, {
                    packageLockJson: packageLockJson.text,
                    packagejson: packageJson.text,
                })
                return null
            }
        })
    )
}

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
                    pattern: `'"${name}"' ${filters}`,
                    type: 'regexp',
                },
                {
                    repositories: {
                        includes: [],
                        type: 'regexp',
                    },
                    files: {
                        includes: ['(^|/)package-lock.json$'],
                        excludes: ['node_modules'],
                        type: 'regexp',
                    },
                    maxResults: 100, // TODO!(sqs): increase
                }
            )
        ).pipe(
            switchMap(textSearchResults =>
                merge(
                    ...textSearchResults.map(textSearchResult =>
                        provideDependencySpecification(textSearchResult, parsedQuery)
                    )
                ).pipe(
                    filter(isDefined),
                    toArray()
                )
            )
        )
    },

    resolveDependencyUpgradeAction: (dep, version) => {
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
}

// TODO!(sqs) removeCommands: [['npm', 'uninstall', ...NPM_OPTS, '--', `${dep.dependency.name}`]],

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
