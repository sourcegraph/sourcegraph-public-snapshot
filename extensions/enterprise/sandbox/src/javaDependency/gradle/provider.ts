import { from, merge, Observable } from 'rxjs'
import { toArray, switchMap, filter, map } from 'rxjs/operators'
import { isDefined, propertyIsDefined } from '../../../../../../shared/src/util/types'
import { createExecServerClient } from '../../execServer/client'
import { memoizedFindTextInFiles } from '../../util'
import { JavaDependencyManagementProvider, JavaDependencyQuery } from '..'
import { DependencySpecification } from '../../dependencyManagement'
import { editForCommands } from '../../execServer/editsForCommands'
import { openTextDocument } from '../../dependencyManagement/util'

const gradleExecClient = createExecServerClient('a8n-java-gradle-exec')

const GRADLE_OPTS = ['--no-audit', '--package-lock-only', '--ignore-scripts']

const provideDependencySpecification = (
    buildGradle: URL,
    query: JavaDependencyQuery
): Observable<DependencySpecification<JavaDependencyQuery> | null> =>
    openTextDocument(buildGradle).pipe(switchMap(buildGradle => {}))

export const gradleDependencyManagementProvider: JavaDependencyManagementProvider = {
    type: 'gradle',
    provideDependencySpecifications: (query, filters = '') =>
        from(
            memoizedFindTextInFiles(
                {
                    pattern: `${JSON.stringify(query.name)} ${filters}`,
                    type: 'regexp',
                },
                {
                    repositories: {
                        type: 'regexp',
                    },
                    files: {
                        includes: ['(^|/)build.gradle$'],
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
                        provideDependencySpecification(new URL(textSearchResult.uri), query)
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
        return editForCommands(
            [
                ...dep.declarations.map(d => d.location.uri),
                ...dep.resolutions.filter(propertyIsDefined('location')).map(d => d.location.uri),
            ],
            [['gradle', 'install', ...GRADLE_OPTS, '--', `${dep.declarations[0].name}@${version}`]],
            gradleExecClient
        )
    },
    resolveDependencyBanAction: dep => {
        // TODO!(sqs): this is not correct w.r.t. indirect deps
        if (dep.declarations.length === 0) {
            console.error('Invalid declarations.', dep)
            throw new Error('invalid declarations')
        }
        return editForCommands(
            [
                ...dep.declarations.map(d => d.location.uri),
                ...dep.resolutions.filter(propertyIsDefined('location')).map(d => d.location.uri),
            ],
            [['gradle', 'uninstall', ...GRADLE_OPTS, '--', `${dep.declarations[0].name}`]],
            gradleExecClient
        )
    },
}

function getPackageLockDependency(
    packageJson: string,
    packageLock: string,
    parsedQuery: JavaDependencyQuery
): Pick<DependencySpecification<JavaDependencyQuery>, 'declarations' | 'resolutions'> {
    const tree = lockTree(JSON.parse(packageJson), JSON.parse(packageLock))
    return traverseJavaLockfile(tree, parsedQuery)
}
