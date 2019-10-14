import { from, merge } from 'rxjs'
import { toArray, switchMap, filter } from 'rxjs/operators'
import { isDefined, propertyIsDefined } from '../../../../../../shared/src/util/types'
import { createExecServerClient } from '../../execServer/client'
import { memoizedFindTextInFiles } from '../../util'
import { PackageJsonDependencyManagementProvider, PackageJsonDependencyQuery } from '../providers'
import { lockTree } from './logicalTree'
import { provideDependencySpecification, editForCommands2, traversePackageJsonLockfile } from '../util'
import { DependencySpecification } from '../../dependencyManagement'

const npmExecClient = createExecServerClient('a8n-npm-exec')

const NPM_OPTS = ['--no-audit', '--package-lock-only', '--ignore-scripts']

export const npmPackageManager: PackageJsonDependencyManagementProvider = {
    type: 'npm',
    provideDependencySpecifications: (query, filters = '') =>
        from(
            memoizedFindTextInFiles(
                {
                    pattern: `'"${query.name}"' ${filters}`,
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
                            query,
                            getPackageLockDependency
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
        if (dep.declarations.length === 0) {
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
    parsedQuery: PackageJsonDependencyQuery
): Pick<DependencySpecification<PackageJsonDependencyQuery>, 'declarations' | 'resolutions'> {
    const tree = lockTree(JSON.parse(packageJson), JSON.parse(packageLock))
    return traversePackageJsonLockfile(tree, parsedQuery)
}
