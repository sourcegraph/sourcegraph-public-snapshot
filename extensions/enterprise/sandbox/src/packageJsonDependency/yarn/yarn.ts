import { from, merge } from 'rxjs'
import { toArray, switchMap, filter } from 'rxjs/operators'
import { isDefined, propertyIsDefined } from '../../../../../../shared/src/util/types'
import { createExecServerClient } from '../../execServer/client'
import { memoizedFindTextInFiles } from '../../util'
import { PackageJsonDependencyManagementProvider, PackageJsonDependencyQuery } from '..'
import { yarnLogicalTree } from './logicalTree'
import { provideDependencySpecification, editForCommands2, editPackageJson, traversePackageJsonLockfile } from '../util'
import * as sourcegraph from 'sourcegraph'
import { DependencySpecification } from '../../dependencyManagement'
import { combineWorkspaceEdits } from '../../../../../../shared/src/api/types/workspaceEdit'

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

export const yarnDependencyManagementProvider: PackageJsonDependencyManagementProvider = {
    type: 'yarn',
    provideDependencySpecifications: (query, filters = '') =>
        from(
            memoizedFindTextInFiles(
                {
                    pattern: `\\b${query.name}\\b ${filters}`,
                    type: 'regexp',
                },
                {
                    repositories: {
                        type: 'regexp',
                    },
                    files: {
                        includes: ['(^|/)yarn.lock$'],
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
                            new URL(textSearchResult.uri.replace(/yarn\.lock$/, 'package.json')),
                            new URL(textSearchResult.uri),
                            query,
                            getYarnLockDependency
                        )
                    )
                ).pipe(
                    filter(isDefined),
                    toArray()
                )
            )
        ),
    resolveDependencyUpgradeAction: (dep, version) => {
        // TODO!(sqs): this is not correct w.r.t. indirect deps
        if (dep.declarations.length === 0) {
            console.error('Invalid declarations.', dep)
            throw new Error('invalid declarations')
        }
        // eslint-disable-next-line no-constant-condition
        if (1 * 2 === 4) {
            // TODO!(sqs)
            console.log(addYarnResolutions)
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
    resolveDependencyBanAction: dep => {
        // TODO!(sqs): this is not correct w.r.t. indirect deps
        if (dep.declarations.length === 0) {
            console.error('Invalid declarations.', dep)
            throw new Error('invalid declarations')
        }
        return editForCommands2(
            [
                ...dep.declarations.map(d => d.location.uri),
                ...dep.resolutions.filter(propertyIsDefined('location')).map(d => d.location.uri),
            ],
            [['yarn', 'remove', ...YARN_OPTS, '--', dep.declarations[0].name]],
            yarnExecClient
        )
    },
}

async function addYarnResolutions(
    dep: DependencySpecification<PackageJsonDependencyQuery>,
    version: string
): Promise<sourcegraph.WorkspaceEdit> {
    if (dep.declarations.length === 0) {
        console.error('Invalid declarations.', dep)
        throw new Error('invalid declarations')
    }

    // Handle indirect dep upgrades by adding to Yarn resolutions. This causes an error in `yarn
    // check` if the resolution is not compatible. TODO(sqs): Find the minimum upgrade path (if
    // any) to eliminate the old version dep without using resolutions.
    const packageJson = await sourcegraph.workspace.openTextDocument(dep.declarations[0].location.uri)
    const workspaceEdit = editPackageJson(packageJson, [
        { path: ['resolutions', dep.declarations[0].name], value: version },
    ])
    const packageJsonObj = JSON.parse(packageJson.text!)
    const otherFiles = dep.resolutions.filter(propertyIsDefined('location')).map(res => res.location.uri)
    const edits2 = await editForCommands2(
        [
            ...otherFiles,
            {
                uri: packageJson.uri,
                text: JSON.stringify({
                    ...packageJsonObj,
                    resolutions: { ...packageJsonObj.resolutions, [dep.declarations[0].name]: version },
                }),
            },
        ],
        [['yarn', ...YARN_OPTS, 'install']],
        yarnExecClient
    ).toPromise()
    return combineWorkspaceEdits([workspaceEdit, edits2])
}

function getYarnLockDependency(
    packageJson: string,
    yarnLock: string,
    parsedQuery: PackageJsonDependencyQuery
): Pick<DependencySpecification<PackageJsonDependencyQuery>, 'declarations' | 'resolutions'> {
    const tree = yarnLogicalTree(JSON.parse(packageJson), yarnLock)
    return traversePackageJsonLockfile(tree, parsedQuery)
}
