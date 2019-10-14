/* eslint-disable @typescript-eslint/no-non-null-assertion */
import { flatten } from 'lodash'
import { from, merge, Observable, combineLatest } from 'rxjs'
import { toArray, map, switchMap, filter } from 'rxjs/operators'
import semver from 'semver'
import * as sourcegraph from 'sourcegraph'
import { isDefined, propertyIsDefined } from '../../../../../../shared/src/util/types'
import { createExecServerClient } from '../../execServer/client'
import { memoizedFindTextInFiles } from '../../util'
import {
    PackageJsonPackageManager,
    ResolvedDependency,
    ResolvedDependencyInPackage,
    PackageJsonDependencyManagementProvider,
    PackageJsonDependencyQuery,
} from '../packageManager'
import { editForDependencyAction, editPackageJson, editForCommands, editForCommands2 } from '../packageManagerCommon'
import { yarnLogicalTree } from './logicalTree'
import { DependencySpecification } from '../../dependencyManagement'

const yarnExecClient = createExecServerClient('a8n-yarn-exec', [])

const YARN_OPTS = [
    '--ignore-engines',
    '--ignore-platform',
    '--ignore-scripts',
    '--non-interactive',
    '--no-node-version-check',
    '--no-progress',
    '--silent',
    // '--mutex network',
    '--mutex',
    'file:/tmp/.yarn-mutex',
    // '--cache-folder',
    // '.sourcegraph-yarn-cache',
    '--skip-integrity-check',
    '--no-default-rc',
]

const provideDependencySpecification = (
    result: sourcegraph.TextSearchResult,
    query: PackageJsonDependencyQuery & { parsedVersionRange: semver.Range }
): Observable<DependencySpecification<PackageJsonDependencyQuery> | null> => {
    const packageJson = from(
        sourcegraph.workspace.openTextDocument(new URL(result.uri.replace(/yarn\.lock$/, 'package.json')))
    )
    const yarnLock = from(sourcegraph.workspace.openTextDocument(new URL(result.uri)))
    return combineLatest([packageJson, yarnLock]).pipe(
        map(([packageJson, yarnLock]) => {
            try {
                // TODO!(sqs): support multiple versions in lockfile/package.json
                const dep = getYarnLockDependency(packageJson.text!, yarnLock.text!, name)
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
                            location: { uri: new URL(yarnLock.uri) },
                        },
                    ],
                }
                return spec
            } catch (err) {
                console.error(`Error checking yarn.lock and package.json for ${result.uri}.`, err, {
                    yarnLock: yarnLock.text,
                    packagejson: packageJson.text,
                })
                return null
            }
        })
    )
}

export const yarnPackageManager: PackageJsonDependencyManagementProvider = {
    provideDependencySpecifications: (query, filters = '') => {
        const parsedQuery = {
            ...query,
            parsedVersionRange: new semver.Range(query.versionRange),
        }
        return from(
            memoizedFindTextInFiles(
                {
                    pattern: `\\b${name}\\b ${filters}`,
                    type: 'regexp',
                },
                {
                    repositories: {
                        includes: [],
                        type: 'regexp',
                    },
                    files: {
                        includes: ['(^|/)yarn.lock$'],
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
            [['yarn', 'upgrade', ...YARN_OPTS, '--', `${dep.declarations[0].name}@${version}`]],
            yarnExecClient
        )
    },
}

// TODO!(sqs) removeCommands: [['yarn', 'remove', ...YARN_OPTS, '--', dep.dependency.name]],

/*const addYarnResolutions = () => {
    // Handle indirect dep upgrades by adding to Yarn resolutions. This causes an error in `yarn
    // check` if the resolution is not compatible. TODO(sqs): Find the minimum upgrade path (if
    // any) to eliminate the old version dep without using resolutions.
    const workspaceEdit = editPackageJson(dep.packageJson, [
        { path: ['resolutions', dep.dependency.name], value: dep.dependency.version },
    ])
    const packageJsonObj = JSON.parse(dep.packageJson.text!)
    const edits2 = await editForCommands(
        {
            lockfile: dep.lockfile,
            packageJson: {
                uri: dep.packageJson.uri,
                text: JSON.stringify({
                    ...packageJsonObj,
                    resolutions: { ...packageJsonObj.resolutions, [dep.dependency.name]: dep.dependency.version },
                }),
            },
        },
        [['yarn', ...YARN_OPTS, 'install']],
        yarnExecClient
    )
    workspaceEdit.set(new URL(dep.lockfile.uri), edits2.get(new URL(dep.lockfile.uri)))
    return workspaceEdit
}*/

function getYarnLockDependency(packageJson: string, yarnLock: string, packageName: string): ResolvedDependency | null {
    const tree = yarnLogicalTree(JSON.parse(packageJson), yarnLock)
    let found: any
    // eslint-disable-next-line ban/ban
    tree.forEach((dep: { name: string }, next: () => void) => {
        if (dep.name === packageName) {
            found = dep
        } else {
            // eslint-disable-next-line callback-return
            next()
        }
    })
    return found ? { name: packageName, version: found.version, direct: !!tree.getDep(packageName) } : null
}
